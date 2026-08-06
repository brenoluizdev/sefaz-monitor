package ui

import (
	"time"

	"github.com/lxn/walk"

	"sefazmonitor/internal/monitor"
	"sefazmonitor/internal/ufs"
)

// ufRow é uma linha da tabela de UFs: metadados fixos + último estado
// conhecido reportado pelo Monitor.
type ufRow struct {
	UF          ufs.UF
	Checked     bool
	Status      monitor.Status
	Message     string
	LastChecked time.Time
}

// ufTableModel alimenta o walk.TableView da janela de configurações. A
// caixa de seleção de cada linha (Checked) é o que define quais UFs entram
// em config.SelectedUFs ao salvar.
type ufTableModel struct {
	walk.TableModelBase
	rows []*ufRow
}

func newUFTableModel(selected map[string]bool) *ufTableModel {
	rows := make([]*ufRow, len(ufs.All))
	for i, u := range ufs.All {
		rows[i] = &ufRow{UF: u, Checked: selected[u.Code]}
	}
	return &ufTableModel{rows: rows}
}

func (m *ufTableModel) RowCount() int { return len(m.rows) }

func (m *ufTableModel) Value(row, col int) interface{} {
	r := m.rows[row]
	switch col {
	case 0:
		return r.UF.Code
	case 1:
		return r.UF.Name
	case 2:
		return r.Status.String()
	case 3:
		return r.Message
	case 4:
		if r.LastChecked.IsZero() {
			return ""
		}
		return r.LastChecked.Format("02/01 15:04:05")
	}
	panic("coluna inesperada")
}

func (m *ufTableModel) Checked(row int) bool { return m.rows[row].Checked }

func (m *ufTableModel) SetChecked(row int, checked bool) error {
	m.rows[row].Checked = checked
	return nil
}

// selectedCodes devolve as siglas das UFs atualmente marcadas.
func (m *ufTableModel) selectedCodes() []string {
	var out []string
	for _, r := range m.rows {
		if r.Checked {
			out = append(out, r.UF.Code)
		}
	}
	return out
}

// applyStates atualiza as colunas de status a partir de um snapshot do
// Monitor e re-renderiza a tabela, preservando o estado marcado/desmarcado
// de cada linha.
func (m *ufTableModel) applyStates(states []monitor.UFState) {
	byUF := make(map[string]monitor.UFState, len(states))
	for _, s := range states {
		byUF[s.UF] = s
	}
	for _, r := range m.rows {
		if s, ok := byUF[r.UF.Code]; ok {
			r.Status = s.Status
			r.Message = s.Message
			r.LastChecked = s.LastChecked
		}
	}
	m.PublishRowsReset()
}
