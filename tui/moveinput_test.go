package tui

import (
	"strings"
	"testing"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/internal/sound"
)

// matchFEN matches text in the position given by fen.
func matchFEN(t *testing.T, fen, text string) inputMatch {
	t.Helper()
	g, err := chess.NewGameFromFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	pos := g.Position()
	return matchInput(&pos, g.LegalMoves(), text)
}

func TestMatchInputSpellings(t *testing.T) {
	const (
		ambiguous = "rnbqkbnr/pppppppp/8/8/8/5N2/PPP1PPPP/RNBQKB1R w KQkq - 0 1"
		pawnOrB   = "4k3/8/8/8/8/2n5/1P1B4/4K3 w - - 0 1"
		castles   = "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"
		promotion = "4k3/P7/8/8/8/8/8/4K3 w - - 0 1"
	)
	tests := []struct {
		fen, text, want string // want: UCI of the single exact match
	}{
		{chess.StartFEN, "Nf3", "g1f3"},
		{chess.StartFEN, "nf3", "g1f3"},
		{chess.StartFEN, "g1f3", "g1f3"},
		{chess.StartFEN, "g1-f3", "g1f3"},
		{chess.StartFEN, "Ng1f3", "g1f3"},
		{chess.StartFEN, "Ngf3", "g1f3"},
		{chess.StartFEN, "Nf3!?", "g1f3"},
		{chess.StartFEN, "e4", "e2e4"},
		{chess.StartFEN, "E4", "e2e4"},
		{ambiguous, "Nbd2", "b1d2"},
		{ambiguous, "N1d2", "b1d2"},
		{ambiguous, "nfd2", "f3d2"},
		{pawnOrB, "bc3", "b2c3"},
		{pawnOrB, "bxc3", "b2c3"},
		{pawnOrB, "Bc3", "d2c3"},
		{pawnOrB, "Bxc3", "d2c3"},
		{pawnOrB, "BXC3", "b2c3"}, // all caps: case-insensitive, pawn preferred
		{castles, "O-O", "e1g1"},
		{castles, "0-0", "e1g1"},
		{castles, "o-o-o", "e1c1"},
		{castles, "O-O+", "e1g1"},
		{promotion, "a8=Q", "a7a8q"},
		{promotion, "a8Q", "a7a8q"},
		{promotion, "a8q", "a7a8q"},
		{promotion, "a7a8n", "a7a8n"},
		{promotion, "a8=N", "a7a8n"},
	}
	for _, tt := range tests {
		im := matchFEN(t, tt.fen, tt.text)
		if len(im.exact) != 1 || im.exact[0].UCI() != tt.want {
			t.Errorf("%q in %s: exact %v, want [%s]", tt.text, tt.fen, im.exact, tt.want)
		}
	}
}

func TestMatchInputPartial(t *testing.T) {
	im := matchFEN(t, chess.StartFEN, "N")
	if len(im.exact) != 0 || len(im.prefix) != 4 {
		t.Fatalf("'N': exact %v prefix %v, want 4 knight moves", im.exact, im.prefix)
	}
	im = matchFEN(t, "rnbqkbnr/pppppppp/8/8/8/5N2/PPP1PPPP/RNBQKB1R w KQkq - 0 1", "Nd2")
	if len(im.exact) != 2 {
		t.Fatalf("'Nd2' should be ambiguous, got %v", im.exact)
	}
	im = matchFEN(t, "4k3/P7/8/8/8/8/8/4K3 w - - 0 1", "a8")
	if len(im.exact) != 0 || !samePromotion(im.prefix) || len(im.prefix) != 4 {
		t.Fatalf("'a8': exact %v prefix %v, want the four promotions", im.exact, im.prefix)
	}
	if im := matchFEN(t, chess.StartFEN, "Nf4"); len(im.exact)+len(im.prefix) != 0 {
		t.Fatalf("'Nf4' matched %v / %v", im.exact, im.prefix)
	}
}

func typeKeys(m *Model, keys ...string) {
	for _, k := range keys {
		key(m, k)
	}
}

func TestPromptPlaysMove(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	typeKeys(m, ":", "n", "f", "3")
	if m.mode != modeInput {
		t.Fatalf("mode %v, want input", m.mode)
	}
	bs := m.boardState(computeLayout(m.width, m.height))
	if !bs.marked.Has(chess.G1) || !bs.targets.Has(chess.F3) {
		t.Fatalf("live highlight: marked %v targets %v", bs.marked, bs.targets)
	}
	if !strings.Contains(m.statusView(), "Nf3") {
		t.Fatalf("status should preview Nf3: %q", m.statusView())
	}
	key(m, "enter")
	if h := m.game.History(); len(h) != 1 || h[0].SAN != "Nf3" {
		t.Fatalf("history %v", h)
	}
	if m.mode != modePlay || m.input != "" {
		t.Fatalf("prompt should close: mode %v input %q", m.mode, m.input)
	}
}

func TestPromptMarksCandidates(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	typeKeys(m, ":", "N")
	bs := m.boardState(computeLayout(m.width, m.height))
	if bs.marked != chess.B1.Bitboard()|chess.G1.Bitboard() || bs.targets != 0 {
		t.Fatalf("marked %v targets %v, want the two knights only", bs.marked, bs.targets)
	}
}

func TestPromptAmbiguousIsRejected(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	rec := &recordPlayer{}
	m.player, m.soundOn = rec, true
	g, _ := chess.NewGameFromFEN("rnbqkbnr/pppppppp/8/8/8/5N2/PPP1PPPP/RNBQKB1R w KQkq - 0 1")
	m.game = g
	typeKeys(m, ":", "N", "d", "2", "enter")
	if m.mode != modeInput || !strings.Contains(m.inputErr, "Nbd2 ou Nfd2") {
		t.Fatalf("mode %v err %q", m.mode, m.inputErr)
	}
	if len(rec.got) != 1 || rec.got[0] != sound.Error {
		t.Fatalf("sounds %v, want the error sound", rec.got)
	}
	key(m, "backspace") // editing clears the error
	if m.inputErr != "" || m.input != "Nd" {
		t.Fatalf("after backspace: input %q err %q", m.input, m.inputErr)
	}
	key(m, "esc")
	if m.mode != modePlay || len(m.game.History()) != 0 {
		t.Fatalf("esc should cancel: mode %v", m.mode)
	}
}

func TestPromptOpensPromotionPicker(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	g, _ := chess.NewGameFromFEN("4k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	m.game = g
	typeKeys(m, ":", "a", "8", "enter")
	if m.mode != modePromotion || len(m.promo) != 4 {
		t.Fatalf("mode %v promo %v, want the picker", m.mode, m.promo)
	}
	key(m, "n")
	pos := m.game.Position()
	if pos.PieceAt(chess.A8) != chess.MakePiece(chess.White, chess.Knight) {
		t.Fatalf("expected knight promotion, got %s", pos.FEN())
	}
}

func TestPromptIgnoresOtherShortcuts(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	typeKeys(m, ":", "d", "r", "q", "s", "m")
	if m.mode != modeInput || m.input != "drqsm" {
		t.Fatalf("shortcuts leaked into the prompt: mode %v input %q", m.mode, m.input)
	}
}
