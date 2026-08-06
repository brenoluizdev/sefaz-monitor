// Package scraper obtém e interpreta a tabela pública de disponibilidade
// dos webservices da NFe, publicada em
// https://www.nfe.fazenda.gov.br/portal/disponibilidade.aspx.
//
// Diferente do webservice NFeStatusServico4 (que exige TLS mútuo com um
// certificado digital ICP-Brasil mesmo só para consultar status), essa
// página é pública e é a mesma fonte visual (bolinha verde/amarela/
// vermelha) que motivou o pedido original: ela já implementa a máquina de
// estados descrita no próprio portal — vermelho após respostas negativas
// seguidas, amarelo na primeira falha (por até 10 min), verde na primeira
// resposta positiva.
package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"
)

const pageURL = "https://www.nfe.fazenda.gov.br/portal/disponibilidade.aspx"

// Color é o estado do semáforo publicado pela SEFAZ para um serviço.
type Color int

const (
	Unknown Color = iota
	Green
	Yellow
	Red
)

func (c Color) String() string {
	switch c {
	case Green:
		return "Verde"
	case Yellow:
		return "Amarelo"
	case Red:
		return "Vermelho"
	default:
		return "Desconhecido"
	}
}

// nomes das colunas de serviço, na ordem em que aparecem na tabela (depois
// da primeira coluna, que é o nome do autorizador/UF).
var serviceColumns = []string{
	"Autorização",
	"Retorno Autorização",
	"Inutilização",
	"Consulta Protocolo",
	"Status Serviço",
	"", // Tempo Médio: não é uma bolinha de status, ignorado
	"Consulta Cadastro",
	"Recepção Evento",
}

// Row é o status de todos os serviços de um autorizador (UF própria, SVRS
// ou SVAN) em um dado instante.
type Row struct {
	Services map[string]Color // chave: nome em serviceColumns (exceto "")
}

// Worst devolve o pior (mais severo) status entre os serviços da linha,
// ignorando colunas não aplicáveis a este autorizador ("-" ou vazias).
func (r Row) Worst() Color {
	worst := Unknown
	for _, c := range r.Services {
		if c > worst {
			worst = c
		}
	}
	return worst
}

// Offenders lista, em português, os serviços que não estão em Verde —
// usado para compor a mensagem de alerta.
func (r Row) Offenders() []string {
	var out []string
	for _, name := range serviceColumns {
		if name == "" {
			continue
		}
		if c, ok := r.Services[name]; ok && c != Green && c != Unknown {
			out = append(out, fmt.Sprintf("%s: %s", name, c))
		}
	}
	return out
}

// Snapshot é o resultado de uma consulta à página de disponibilidade.
type Snapshot struct {
	Rows            map[string]Row // chave: "AM", "BA", ..., "SVRS", "SVAN"
	ServerCheckedAt string         // timestamp que a própria SEFAZ informa na página ("Última Verificação")
	FetchedAt       time.Time
}

var (
	reTable   = regexp.MustCompile(`(?s)id="ctl00_ContentPlaceHolder1_gdvDisponibilidade2"[^>]*>(.*?)</table>`)
	reRow     = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	reCell    = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	reBall    = regexp.MustCompile(`bola_(verde|amarela?|vermelho)`)
	reChecked = regexp.MustCompile(`Última Verificação:\s*([0-9/:\s]+)`)
)

func cellColor(cell string) Color {
	m := reBall.FindStringSubmatch(cell)
	if m == nil {
		return Unknown
	}
	switch {
	case strings.HasPrefix(m[1], "verde"):
		return Green
	case strings.HasPrefix(m[1], "amarel"):
		return Yellow
	case strings.HasPrefix(m[1], "vermelho"):
		return Red
	}
	return Unknown
}

// Fetch baixa e interpreta a página de disponibilidade.
func Fetch(timeout time.Duration) (Snapshot, error) {
	snap := Snapshot{Rows: make(map[string]Row), FetchedAt: time.Now()}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return snap, err
	}
	client := &http.Client{Timeout: timeout, Jar: jar}

	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return snap, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) SefazMonitor/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return snap, fmt.Errorf("falha ao acessar o portal da NFe: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return snap, err
	}
	if resp.StatusCode != http.StatusOK {
		return snap, fmt.Errorf("portal da NFe respondeu HTTP %d", resp.StatusCode)
	}
	html := string(body)

	if m := reChecked.FindStringSubmatch(html); m != nil {
		snap.ServerCheckedAt = strings.TrimSpace(m[1])
	}

	tableMatch := reTable.FindStringSubmatch(html)
	if tableMatch == nil {
		return snap, fmt.Errorf("tabela de disponibilidade não encontrada na página (layout pode ter mudado)")
	}

	for _, rowMatch := range reRow.FindAllStringSubmatch(tableMatch[1], -1) {
		cells := reCell.FindAllStringSubmatch(rowMatch[1], -1)
		if len(cells) == 0 {
			continue // linha de cabeçalho (usa <th>, não <td>)
		}
		env := strings.TrimSpace(cells[0][1])
		if env == "" {
			continue
		}

		row := Row{Services: make(map[string]Color)}
		for i, name := range serviceColumns {
			if name == "" {
				continue
			}
			cellIdx := i + 1 // cells[0] é o código do autorizador
			if cellIdx >= len(cells) {
				continue
			}
			row.Services[name] = cellColor(cells[cellIdx][1])
		}
		snap.Rows[env] = row
	}

	if len(snap.Rows) == 0 {
		return snap, fmt.Errorf("nenhuma linha de status encontrada na página (layout pode ter mudado)")
	}
	return snap, nil
}
