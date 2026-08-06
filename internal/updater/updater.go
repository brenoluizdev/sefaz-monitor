// Package updater verifica, baixa e instala novas versões do app a partir
// de GitHub Releases.
//
// Fluxo de segurança: o instalador baixado só é executado se seu SHA-256
// bater com o checksum publicado no mesmo release (também baixado do
// GitHub) — isso protege contra download corrompido ou incompleto, mas não
// contra um release comprometido na origem (para isso seria necessário
// assinatura de código, fora do escopo deste projeto pessoal).
package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sefazmonitor/internal/version"
)

const (
	repo                = "brenoluizdev/sefaz-monitor"
	installerAssetName  = "SefazMonitorSetup.exe"
	checksumAssetName   = "SefazMonitorSetup.exe.sha256"
	apiTimeout          = 20 * time.Second
	downloadTimeout     = 3 * time.Minute
)

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

func findAsset(rel release, name string) (asset, bool) {
	for _, a := range rel.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return asset{}, false
}

func latestRelease() (release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "SefazMonitor-Updater/"+version.Current)

	client := &http.Client{Timeout: apiTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("GitHub API respondeu HTTP %d", resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return release{}, err
	}
	return rel, nil
}

// parseVersion interpreta "v1.2.3" ou "1.2.3" como (1, 2, 3, true).
func parseVersion(tag string) (major, minor, patch int, ok bool) {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	parts := strings.SplitN(tag, ".", 3)
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return 0, 0, 0, false
		}
		nums[i] = n
	}
	return nums[0], nums[1], nums[2], true
}

// IsNewer diz se latestTag representa uma versão mais nova que current.
// Tags que não seguem major.minor.patch nunca são consideradas mais novas
// (mais seguro deixar de notificar uma atualização do que instalar algo
// inesperado por engano de parsing).
func IsNewer(latestTag, current string) bool {
	lm, ln, lp, ok1 := parseVersion(latestTag)
	cm, cn, cp, ok2 := parseVersion(current)
	if !ok1 || !ok2 {
		return false
	}
	if lm != cm {
		return lm > cm
	}
	if ln != cn {
		return ln > cn
	}
	return lp > cp
}

func downloadToTemp(url, suffix string) (string, error) {
	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download respondeu HTTP %d", resp.StatusCode)
	}

	f, err := os.CreateTemp("", "sefazmonitor-update-*"+suffix)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func readChecksum(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Aceita tanto uma linha só com o hash quanto o formato "hash  arquivo".
	fields := strings.Fields(string(b))
	if len(fields) == 0 {
		return "", fmt.Errorf("arquivo de checksum vazio")
	}
	return strings.ToLower(fields[0]), nil
}

// CheckAndInstall verifica se há uma versão publicada mais nova que a
// atual; se houver, baixa o instalador e o checksum, confirma a
// integridade, e então dispara a instalação silenciosa (que se encarrega de
// fechar o app em execução, sobrescrever o binário, e reabri-lo).
//
// onBeforeInstall é chamado (com a tag da nova versão) só depois que o
// download e a verificação de integridade já deram certo, imediatamente
// antes de disparar o instalador — é o sinal para o chamador se despedir do
// usuário e encerrar o processo.
//
// found=true e err=nil significa que a instalação foi disparada com
// sucesso; o processo atual deve terminar logo em seguida (o arquivo do
// próprio executável precisa ficar livre para o instalador sobrescrever).
func CheckAndInstall(onBeforeInstall func(newVersion string)) (found bool, err error) {
	if version.IsDev() {
		return false, nil
	}

	rel, err := latestRelease()
	if err != nil {
		return false, fmt.Errorf("consultar releases: %w", err)
	}
	if !IsNewer(rel.TagName, version.Current) {
		return false, nil
	}

	installerA, ok := findAsset(rel, installerAssetName)
	if !ok {
		return false, fmt.Errorf("release %s não tem o asset %s", rel.TagName, installerAssetName)
	}
	checksumA, ok := findAsset(rel, checksumAssetName)
	if !ok {
		return false, fmt.Errorf("release %s não tem o asset %s", rel.TagName, checksumAssetName)
	}

	installerPath, err := downloadToTemp(installerA.BrowserDownloadURL, ".exe")
	if err != nil {
		return false, fmt.Errorf("baixar instalador: %w", err)
	}
	checksumPath, err := downloadToTemp(checksumA.BrowserDownloadURL, ".sha256")
	if err != nil {
		os.Remove(installerPath)
		return false, fmt.Errorf("baixar checksum: %w", err)
	}
	defer os.Remove(checksumPath)

	wantHex, err := readChecksum(checksumPath)
	if err != nil {
		os.Remove(installerPath)
		return false, fmt.Errorf("ler checksum: %w", err)
	}
	gotHex, err := sha256File(installerPath)
	if err != nil {
		os.Remove(installerPath)
		return false, fmt.Errorf("calcular checksum: %w", err)
	}
	if !strings.EqualFold(wantHex, gotHex) {
		os.Remove(installerPath)
		return false, fmt.Errorf("checksum não corresponde (baixado %s, esperado %s) — download descartado por segurança", gotHex, wantHex)
	}

	if onBeforeInstall != nil {
		onBeforeInstall(rel.TagName)
	}

	// /CLOSEAPPLICATIONS + /RESTARTAPPLICATIONS: o instalador fecha
	// qualquer processo detectado pelo AppMutex (ver [Setup] no .iss) antes
	// de sobrescrever o executável, e reabre o app depois. O renamedPath
	// serve só de cosmético para o Gerenciador de Tarefas durante a
	// instalação, não afeta a lógica.
	renamedPath := filepath.Join(filepath.Dir(installerPath), "SefazMonitorSetup.exe")
	if renamedPath != installerPath {
		if err := os.Rename(installerPath, renamedPath); err == nil {
			installerPath = renamedPath
		}
	}

	cmd := exec.Command(installerPath, "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/CLOSEAPPLICATIONS", "/RESTARTAPPLICATIONS")
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("iniciar instalador: %w", err)
	}

	return true, nil
}
