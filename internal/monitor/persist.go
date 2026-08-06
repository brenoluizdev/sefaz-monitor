package monitor

import (
	"encoding/json"
	"os"
)

// SaveStates grava o snapshot de estados em disco. Existe para que um
// reinício do processo — incluindo os disparados pela própria
// auto-atualização — não apague a memória do último status conhecido de
// cada UF. Sem isso, a primeira verificação após qualquer reinício é
// tratada como "estado inicial" (ver Seed) e a notificação da mudança que
// eventualmente aconteceu durante o reinício nunca é disparada.
func SaveStates(path string, states []UFState) error {
	b, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// LoadStates lê o snapshot salvo por SaveStates. Arquivo ausente ou
// corrompido não é erro — só significa que não há estado anterior
// conhecido (ex.: primeira execução).
func LoadStates(path string) []UFState {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var states []UFState
	if err := json.Unmarshal(b, &states); err != nil {
		return nil
	}
	return states
}

// Seed pré-popula o estado conhecido de cada UF a partir de um snapshot
// carregado (ver LoadStates). Precisa ser chamado antes de Start(): assim,
// a primeira verificação de fato compara contra o último estado real
// conhecido em vez de contra Unknown, e uma mudança real ocorrida enquanto
// o processo estava reiniciando é corretamente notificada.
func (m *Monitor) Seed(states []UFState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range states {
		m.states[s.UF] = s
	}
}
