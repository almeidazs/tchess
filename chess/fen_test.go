package chess

import (
	"testing"
)

func TestFENRoundTrip(t *testing.T) {
	for _, tc := range perftSuite {
		p := mustFEN(t, tc.fen)
		if got := p.FEN(); got != tc.fen {
			t.Errorf("FEN round trip: got %q want %q", got, tc.fen)
		}
	}
}

func TestFENRejectsInvalid(t *testing.T) {
	bad := []string{
		"",
		"8/8/8/8/8/8/8/8 w - - 0 1", // no kings
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1", // side
		"rnbqkbnr/pppppppp/9/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // 9 files
		"Pnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // pawn on rank 8
		"4k3/8/8/8/8/8/8/4K2r b - - 0 1",                           // white (not to move) is in check
	}
	for _, fen := range bad {
		if _, err := ParseFEN(fen); err == nil {
			t.Errorf("ParseFEN(%q) accepted an invalid position", fen)
		}
	}
}

func TestFENNormalizesRights(t *testing.T) {
	// Castling rights without a rook and an unusable ep square are dropped.
	p := mustFEN(t, "4k3/8/8/8/4P3/8/8/4K3 b KQkq e3 0 1")
	if p.Castling() != NoCastling || p.EnPassant() != NoSquare {
		t.Fatalf("got castling %b ep %s", p.Castling(), p.EnPassant())
	}
}
