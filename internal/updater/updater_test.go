package updater

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want             bool
	}{
		{"v1.2.0", "1.1.0", true},
		{"1.2.0", "v1.1.0", true},
		{"v1.1.0", "v1.1.0", false},
		{"v1.0.9", "v1.1.0", false},
		{"v2.0.0", "v1.9.9", true},
		{"v1.1.1", "v1.1.0", true},
		{"garbage", "v1.0.0", false},
		{"v1.0.0", "dev", false},
	}

	for _, c := range cases {
		if got := IsNewer(c.latest, c.current); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}
