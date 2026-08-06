// Package monitor implementa o polling periódico da página pública de
// disponibilidade da NFe e a máquina de estados (verde/amarelo/vermelho)
// usada para decidir quando disparar um alerta de transição por UF.
package monitor

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"sefazmonitor/internal/config"
	"sefazmonitor/internal/scraper"
	"sefazmonitor/internal/ufs"
)

// Status é o nível de saúde classificado para uma UF.
type Status int

const (
	Unknown Status = iota
	OK
	Degraded
	Down
)

func (s Status) String() string {
	switch s {
	case OK:
		return "Operacional"
	case Degraded:
		return "Instável"
	case Down:
		return "Indisponível"
	default:
		return "Desconhecido"
	}
}

const fetchTimeout = 20 * time.Second

// UFState é o snapshot de estado mais recente de uma UF monitorada.
type UFState struct {
	UF          string
	Status      Status
	Message     string
	LastChecked time.Time
	LastChanged time.Time
}

func colorToStatus(c scraper.Color) Status {
	switch c {
	case scraper.Green:
		return OK
	case scraper.Yellow:
		return Degraded
	case scraper.Red:
		return Down
	default:
		return Unknown
	}
}

// classify decide o status de uma UF a partir do snapshot da página de
// disponibilidade. err != nil indica que não foi possível nem buscar a
// página (ex.: sem internet) — neste caso todas as UFs ficam "Indisponível"
// com a mensagem explicando a falha, já que não há como saber o estado real
// da SEFAZ nesse momento.
func classify(snap scraper.Snapshot, fetchErr error, u ufs.UF) (Status, string) {
	if fetchErr != nil {
		return Down, fmt.Sprintf("Falha ao consultar o portal da NFe: %v", fetchErr)
	}

	row, ok := snap.Rows[u.EnvKey()]
	if !ok {
		return Unknown, fmt.Sprintf("Autorizador %q não encontrado na página de disponibilidade", u.EnvKey())
	}

	status := colorToStatus(row.Worst())
	msg := "Todos os serviços operacionais"
	if offenders := row.Offenders(); len(offenders) > 0 {
		msg = strings.Join(offenders, "; ")
	}
	if snap.ServerCheckedAt != "" {
		msg += fmt.Sprintf(" (SEFAZ verificou às %s)", snap.ServerCheckedAt)
	}
	return status, msg
}

// Monitor executa o polling periódico e mantém o último estado conhecido de
// cada UF configurada.
type Monitor struct {
	mu     sync.Mutex
	cfg    config.Config
	states map[string]UFState

	onTransition func(old, new UFState)
	onUpdate     func()

	stopCh  chan struct{}
	running bool
}

// New cria um Monitor. onTransition é chamado sempre que o Status de uma UF
// muda (inclusive na primeira classificação, saindo de Unknown). onUpdate é
// chamado uma vez ao final de cada ciclo de verificação, para a UI se
// atualizar.
func New(onTransition func(old, new UFState), onUpdate func()) *Monitor {
	return &Monitor{
		states:       make(map[string]UFState),
		onTransition: onTransition,
		onUpdate:     onUpdate,
	}
}

// SetConfig atualiza a configuração (UFs monitoradas e intervalo). Pode ser
// chamado com o monitor já rodando.
func (m *Monitor) SetConfig(cfg config.Config) {
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
}

// Start inicia o loop de polling em background. Idempotente.
func (m *Monitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.stopCh = make(chan struct{})
	stop := m.stopCh
	m.mu.Unlock()

	go m.loop(stop)
}

// Stop encerra o loop de polling.
func (m *Monitor) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	close(m.stopCh)
	m.mu.Unlock()
}

func (m *Monitor) loop(stop chan struct{}) {
	m.CheckNow()

	for {
		m.mu.Lock()
		interval := time.Duration(m.cfg.IntervalMinutes) * time.Minute
		m.mu.Unlock()
		if interval <= 0 {
			interval = 10 * time.Minute
		}

		select {
		case <-stop:
			return
		case <-time.After(interval):
			m.CheckNow()
		}
	}
}

// CheckNow dispara imediatamente um ciclo de verificação de todas as UFs
// configuradas, sem esperar o intervalo do ticker. Faz uma única busca à
// página de disponibilidade (não uma por UF), já que ela traz todos os
// autorizadores de uma vez.
func (m *Monitor) CheckNow() {
	m.mu.Lock()
	codes := append([]string(nil), m.cfg.SelectedUFs...)
	m.mu.Unlock()

	if len(codes) == 0 {
		if m.onUpdate != nil {
			m.onUpdate()
		}
		return
	}

	snap, err := scraper.Fetch(fetchTimeout)

	now := time.Now()
	for _, code := range codes {
		u, ok := ufs.ByCode(code)
		if !ok {
			continue
		}
		status, msg := classify(snap, err, u)

		m.mu.Lock()
		old, existed := m.states[code]
		newState := UFState{
			UF:          code,
			Status:      status,
			Message:     msg,
			LastChecked: now,
			LastChanged: old.LastChanged,
		}
		changed := !existed || old.Status != status
		if changed {
			newState.LastChanged = now
		}
		m.states[code] = newState
		m.mu.Unlock()

		if changed && m.onTransition != nil {
			m.onTransition(old, newState)
		}
	}

	if m.onUpdate != nil {
		m.onUpdate()
	}
}

// States devolve um snapshot ordenado (por sigla de UF) do estado atual.
func (m *Monitor) States() []UFState {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]UFState, 0, len(m.states))
	for _, s := range m.states {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UF < out[j].UF })
	return out
}

// Worst devolve o pior status entre todas as UFs monitoradas (usado para
// decidir a cor do ícone da bandeja). Retorna Unknown se nada monitorado.
func (m *Monitor) Worst() Status {
	m.mu.Lock()
	defer m.mu.Unlock()

	worst := Unknown
	for _, s := range m.states {
		if s.Status > worst {
			worst = s.Status
		}
	}
	return worst
}
