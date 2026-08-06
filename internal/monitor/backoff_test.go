package monitor

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	base := 30 * time.Second

	cases := []struct {
		failures int
		want     time.Duration
	}{
		{0, 30 * time.Second},
		{1, 60 * time.Second},
		{2, 120 * time.Second},
		{3, 240 * time.Second},
		{10, maxBackoffInterval}, // muito acima do teto de shift, deve saturar
	}

	for _, c := range cases {
		got := backoff(base, c.failures)
		if got != c.want {
			t.Errorf("backoff(%s, %d) = %s, want %s", base, c.failures, got, c.want)
		}
	}
}

func TestBackoffCapsAtMax(t *testing.T) {
	got := backoff(10*time.Minute, 5)
	if got != maxBackoffInterval {
		t.Errorf("esperava saturar em %s, obtive %s", maxBackoffInterval, got)
	}
}
