package monitor

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadStatesRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "states.json")

	want := []UFState{
		{UF: "SP", Status: OK, Message: "tudo certo", LastChecked: time.Now().Truncate(time.Second)},
		{UF: "MG", Status: Degraded, Message: "instável", LastChecked: time.Now().Truncate(time.Second)},
	}

	if err := SaveStates(path, want); err != nil {
		t.Fatalf("SaveStates: %v", err)
	}

	got := LoadStates(path)
	if len(got) != len(want) {
		t.Fatalf("LoadStates devolveu %d itens, esperava %d", len(got), len(want))
	}
	for i := range want {
		if got[i].UF != want[i].UF || got[i].Status != want[i].Status {
			t.Errorf("item %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadStatesMissingFile(t *testing.T) {
	got := LoadStates(filepath.Join(t.TempDir(), "nao-existe.json"))
	if got != nil {
		t.Errorf("esperava nil para arquivo ausente, obtive %+v", got)
	}
}

// TestSeedMakesPriorStateVisible confirma o núcleo do fix: depois de Seed,
// o estado de uma UF deixa de ser "desconhecido" — é exatamente essa
// diferença que faz onTransition (na camada ui) não silenciar uma mudança
// real ocorrida enquanto o processo estava reiniciando.
func TestSeedMakesPriorStateVisible(t *testing.T) {
	m := New(nil, nil)

	before := m.States()
	if len(before) != 0 {
		t.Fatalf("esperava nenhum estado antes do Seed, obtive %+v", before)
	}

	m.Seed([]UFState{{UF: "SP", Status: Degraded, Message: "instável"}})

	after := m.States()
	if len(after) != 1 || after[0].UF != "SP" || after[0].Status != Degraded {
		t.Fatalf("esperava SP=Degraded depois do Seed, obtive %+v", after)
	}
}
