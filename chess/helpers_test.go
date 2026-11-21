package chess

import (
	"slices"
	"testing"
)

func mustFEN(t *testing.T, fen string) *Position {
	t.Helper()
	p, err := ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN(%q): %v", fen, err)
	}
	return p
}

func legalUCI(p *Position) []string {
	var ml MoveList
	p.LegalMoves(&ml)
	out := make([]string, 0, ml.Len())
	for _, m := range ml.Slice() {
		out = append(out, m.UCI())
	}
	slices.Sort(out)
	return out
}

func playSAN(t *testing.T, g *Game, moves ...string) {
	t.Helper()
	for _, s := range moves {
		pos := g.Position()
		m, err := pos.ParseSAN(s)
		if err != nil {
			t.Fatalf("move %q at %s: %v", s, pos.FEN(), err)
		}
		if err := g.Play(m); err != nil {
			t.Fatalf("play %q: %v", s, err)
		}
	}
}
