package chess

import (
	"testing"
)

func TestSAN(t *testing.T) {
	cases := []struct{ fen, uci, san string }{
		{StartFEN, "g1f3", "Nf3"},
		{StartFEN, "e2e4", "e4"},
		{"4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1", "e1g1", "O-O"},
		{"4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1", "e1c1", "O-O-O"},
		{"4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1", "a1a8", "Ra8+"},
		{"7k/8/6K1/8/8/8/8/R7 w - - 0 1", "a1a8", "Ra8#"},
		{"4k3/8/8/8/8/8/8/R4RK1 w - - 0 1", "a1d1", "Rad1"},
		{"4k3/8/8/8/R7/8/8/R5K1 w - - 0 1", "a1a2", "R1a2"},
		{"4k3/8/8/8/8/1N3N2/8/5NK1 w - - 0 1", "f3d2", "Nf3d2"}, // f1 shares the file, b3 the rank
		{"1n2k3/P7/8/8/8/8/8/4K3 w - - 0 1", "a7b8q", "axb8=Q+"},
		{"rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3", "e5d6", "exd6"},
	}
	for _, c := range cases {
		p := mustFEN(t, c.fen)
		m, err := p.ParseUCI(c.uci)
		if err != nil {
			t.Fatalf("%s: %v", c.uci, err)
		}
		if got := p.SAN(m); got != c.san {
			t.Errorf("SAN(%s) in %s = %q, want %q", c.uci, c.fen, got, c.san)
		}
		back, err := p.ParseSAN(c.san)
		if err != nil || back != m {
			t.Errorf("ParseSAN(%q) = %v, %v", c.san, back, err)
		}
	}
}
