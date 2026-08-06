// Package config gerencia a configuração persistida do usuário: quais UFs
// monitorar, o intervalo de consulta e o ambiente (produção/homologação).
package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
)

// IntervalFloorSeconds é o menor intervalo aceito entre verificações. Existe
// para evitar bater no portal público da NFe com mais frequência do que
// qualquer sistema automatizado deveria (risco de bloqueio por WAF), mas
// permite ir bem mais rápido que o padrão para quem quiser.
const IntervalFloorSeconds = 15

// Config é o estado configurável pelo usuário, salvo em disco.
type Config struct {
	SelectedUFs     []string `json:"selectedUFs"`
	IntervalSeconds int      `json:"intervalSeconds"`
}

// Default retorna a configuração inicial usada quando não há nada salvo
// ainda: nenhuma UF selecionada, intervalo de 10 minutos (razoável para não
// sobrecarregar o portal da NFe, mas ajustável nas configurações).
func Default() Config {
	return Config{
		SelectedUFs:     nil,
		IntervalSeconds: 600,
	}
}

// Dir retorna (e garante que exista) o diretório de dados do app em
// %APPDATA%\SefazMonitor.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "SefazMonitor")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load lê a configuração salva. Um arquivo ausente, corrompido ou
// ilegível nunca é fatal: sempre devolve uma Config utilizável (Default()
// nesses casos), já que perder a configuração salva não deveria impedir o
// app de abrir — o usuário apenas precisa reconfigurar as UFs.
func Load() Config {
	p, err := path()
	if err != nil {
		return Default()
	}

	b, err := os.ReadFile(p)
	if err != nil {
		return Default()
	}
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF}) // BOM UTF-8, se presente

	cfg := Default()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Default()
	}
	if cfg.IntervalSeconds < IntervalFloorSeconds {
		cfg.IntervalSeconds = IntervalFloorSeconds
	}
	return cfg
}

// Save grava a configuração em disco.
func Save(cfg Config) error {
	p, err := path()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}
