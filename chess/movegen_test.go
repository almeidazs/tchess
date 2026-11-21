package chess

import (
	"slices"
	"testing"
)

func TestPiecesCannotJump(t *testing.T) {
	p := StartingPosition()
	for _, uci := range []string{"a1a3", "c1e3", "d1d3", "f1c4", "e1e2"} {
		if _, err := p.ParseUCI(uci); err == nil {
			t.Errorf("%s should be blocked in the initial position", uci)
		}
	}
	// Knights do jump.
	if _, err := p.ParseUCI("g1f3"); err != nil {
		t.Error(err)
	}
}

func TestEnPassant(t *testing.T) {
	g := NewGame()
	playSAN(t, g, "e4", "a6", "e5", "d5")
	pos := g.Position()
	m, err := pos.ParseSAN("exd6")
	if err != nil {
		t.Fatal(err)
	}
	if !m.IsEnPassant() {
		t.Fatal("exd6 should be an en passant capture")
	}
	if err := g.Play(m); err != nil {
		t.Fatal(err)
	}
	after := g.Position()
	if after.PieceAt(D5) != NoPiece || after.PieceAt(D6) != MakePiece(White, Pawn) {
		t.Fatalf("en passant left %s", after.FEN())
	}
}

func TestEnPassantOnlyImmediately(t *testing.T) {
	g := NewGame()
	playSAN(t, g, "e4", "a6", "e5", "d5", "Nf3", "Nf6")
	pos := g.Position()
	if _, err := pos.ParseSAN("exd6"); err == nil {
		t.Fatal("en passant must be played immediately")
	}
}

func TestEnPassantDiscoveredCheck(t *testing.T) {
	// Capturing en passant would remove both pawns from the 5th rank and
	// expose the white king to the rook.
	p := mustFEN(t, "8/8/8/K2pP2r/8/8/8/7k w - d6 0 1")
	if slices.Contains(legalUCI(p), "e5d6") {
		t.Fatal("en passant exposing the king must be illegal")
	}
}

func TestEnPassantResolvesCheck(t *testing.T) {
	// The pawn that just double-pushed gives check; capturing it en passant
	// is legal.
	p := mustFEN(t, "8/8/8/2k5/3Pp3/8/8/4K3 b - d3 0 1")
	if !slices.Contains(legalUCI(p), "e4d3") {
		t.Fatalf("en passant capturing the checker should be legal: %v", legalUCI(p))
	}
}

func TestCastling(t *testing.T) {
	p := mustFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	moves := legalUCI(p)
	for _, want := range []string{"e1g1", "e1c1"} {
		if !slices.Contains(moves, want) {
			t.Errorf("missing %s", want)
		}
	}
	m, _ := p.ParseUCI("e1g1")
	p.Make(m)
	if p.PieceAt(F1) != MakePiece(White, Rook) || p.PieceAt(G1) != MakePiece(White, King) {
		t.Fatalf("castling result: %s", p.FEN())
	}
	if p.Castling()&(WhiteKingside|WhiteQueenside) != 0 {
		t.Fatal("white keeps castling rights after castling")
	}
}

func TestCastlingRestrictions(t *testing.T) {
	cases := []struct {
		name, fen, move string
	}{
		{"in check", "r3k2r/8/8/8/4r3/8/8/R3K2R w KQkq - 0 1", "e1g1"},
		{"through attacked f1", "r3k2r/8/8/8/5r2/8/8/R3K2R w KQkq - 0 1", "e1g1"},
		{"into attacked g1", "r3k2r/8/8/8/6r1/8/8/R3K2R w KQkq - 0 1", "e1g1"},
		{"through attacked d1", "r3k2r/8/8/8/3r4/8/8/R3K2R w KQkq - 0 1", "e1c1"},
		{"blocked b1", "r3k2r/8/8/8/8/8/8/RN2K2R w KQkq - 0 1", "e1c1"},
		{"no right", "r3k2r/8/8/8/8/8/8/R3K2R w Qkq - 0 1", "e1g1"},
	}
	for _, c := range cases {
		if slices.Contains(legalUCI(mustFEN(t, c.fen)), c.move) {
			t.Errorf("%s: %s should be illegal", c.name, c.move)
		}
	}
	// b1 attacked does not prevent queenside castling.
	if !slices.Contains(legalUCI(mustFEN(t, "r3k2r/8/8/8/1r6/8/8/R3K2R w KQkq - 0 1")), "e1c1") {
		t.Error("queenside castling with b1 attacked should be legal")
	}
}

func TestCastlingRightsLostWhenRookCaptured(t *testing.T) {
	g, _ := NewGameFromFEN("r3k2r/8/8/8/8/8/6B1/R3K2R w KQkq - 0 1")
	playSAN(t, g, "Bxa8")
	pos := g.Position()
	if pos.Castling()&BlackQueenside != 0 {
		t.Fatal("black keeps queenside rights after losing the a8 rook")
	}
}

func TestPinnedPieceCannotLeaveLine(t *testing.T) {
	p := mustFEN(t, "4k3/4r3/8/8/8/8/4N3/4K3 w - - 0 1")
	for _, m := range legalUCI(p) {
		if m[:2] == "e2" {
			t.Fatalf("pinned knight moved: %s", m)
		}
	}
	// A pinned rook slides along the pin.
	p = mustFEN(t, "4k3/4r3/8/8/8/8/4R3/4K3 w - - 0 1")
	moves := legalUCI(p)
	if !slices.Contains(moves, "e2e7") || slices.Contains(moves, "e2d2") {
		t.Fatalf("pinned rook moves wrong: %v", moves)
	}
}

func TestDoubleCheckOnlyKingMoves(t *testing.T) {
	p := mustFEN(t, "4k3/8/8/8/1b6/8/2N5/r3K3 w - - 0 1")
	for _, m := range legalUCI(p) {
		if m[:2] != "e1" {
			t.Fatalf("non-king move in double check: %s", m)
		}
	}
}

func TestPromotion(t *testing.T) {
	p := mustFEN(t, "1n2k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	moves := legalUCI(p)
	for _, want := range []string{"a7a8q", "a7a8r", "a7a8b", "a7a8n", "a7b8q", "a7b8n"} {
		if !slices.Contains(moves, want) {
			t.Errorf("missing promotion %s", want)
		}
	}
	m, _ := p.ParseUCI("a7b8n")
	p.Make(m)
	if p.PieceAt(B8) != MakePiece(White, Knight) {
		t.Fatalf("promotion result: %s", p.FEN())
	}
}
