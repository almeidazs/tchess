package chess

import (
	"errors"
	"slices"
	"testing"
)

func TestCheckmate(t *testing.T) {
	g := NewGame()
	playSAN(t, g, "f3", "e5", "g4", "Qh4#")
	if g.Outcome() != BlackWins || g.Termination() != Checkmate {
		t.Fatalf("got %v / %v", g.Outcome(), g.Termination())
	}
	if !errors.Is(g.Play(NewMove(A2, A3, FlagQuiet)), ErrGameOver) {
		t.Fatal("moves must be rejected after the game ends")
	}
	if h := g.History(); h[len(h)-1].SAN != "Qh4#" {
		t.Fatalf("last SAN %q", h[len(h)-1].SAN)
	}
}

func TestStalemate(t *testing.T) {
	g, _ := NewGameFromFEN("7k/8/6Q1/8/8/8/8/K7 w - - 0 1")
	playSAN(t, g, "Qf7")
	if g.Outcome() != Draw || g.Termination() != Stalemate {
		t.Fatalf("got %v / %v", g.Outcome(), g.Termination())
	}
}

func TestInsufficientMaterial(t *testing.T) {
	dead := []string{
		"8/8/8/4k3/8/8/8/4K3 w - - 0 1",
		"8/8/8/4k3/8/8/8/4KN2 w - - 0 1",
		"8/8/8/4k3/8/8/8/4KB2 w - - 0 1",
		"8/8/8/4k3/2b5/8/8/4KB2 w - - 0 1", // f1 and c4 are both light squares
	}
	alive := []string{
		"8/8/8/2b1k3/8/8/8/4KB2 w - - 0 1", // opposite-colored bishops (c5 dark, f1 light)
		"8/8/8/4k3/8/8/8/3NKN2 w - - 0 1",
		"8/8/8/4k3/8/8/4P3/4K3 w - - 0 1",
		"8/8/8/4kn2/8/8/8/4KB2 w - - 0 1",
	}
	for _, f := range dead {
		if !mustFEN(t, f).IsInsufficientMaterial() {
			t.Errorf("%s should be insufficient material", f)
		}
	}
	for _, f := range alive {
		if mustFEN(t, f).IsInsufficientMaterial() {
			t.Errorf("%s should not be insufficient material", f)
		}
	}
	g, _ := NewGameFromFEN("8/8/8/4k3/8/8/3p4/4KN2 w - - 0 1")
	playSAN(t, g, "Kxd2")
	if g.Termination() != InsufficientMaterial {
		t.Fatalf("got %v", g.Termination())
	}
}

func TestRepetition(t *testing.T) {
	g := NewGame()
	shuffle := []string{"Nf3", "Nf6", "Ng1", "Ng8"}
	playSAN(t, g, shuffle...)
	if g.RepetitionCount() != 2 || g.ClaimableDraw() != NoTermination {
		t.Fatalf("after one cycle: count %d", g.RepetitionCount())
	}
	playSAN(t, g, shuffle...)
	if g.ClaimableDraw() != ThreefoldRepetition {
		t.Fatalf("threefold should be claimable, count %d", g.RepetitionCount())
	}
	if g.IsOver() {
		t.Fatal("threefold must not end the game automatically")
	}
	playSAN(t, g, shuffle...)
	playSAN(t, g, shuffle...)
	if g.Termination() != FivefoldRepetition {
		t.Fatalf("fivefold should end the game, got %v (count %d)", g.Termination(), g.RepetitionCount())
	}
}

func TestClaimThreefold(t *testing.T) {
	g := NewGame()
	if !errors.Is(g.ClaimDraw(), ErrNoClaim) {
		t.Fatal("no claim available at the start")
	}
	for range 2 {
		playSAN(t, g, "Nc3", "Nc6", "Nb1", "Nb8")
	}
	if err := g.ClaimDraw(); err != nil {
		t.Fatal(err)
	}
	if g.Outcome() != Draw || g.Termination() != ThreefoldRepetition {
		t.Fatalf("got %v / %v", g.Outcome(), g.Termination())
	}
}

func TestRepetitionIgnoresUnusableEnPassant(t *testing.T) {
	// After 1.e4 no black pawn can capture en passant, so the position after
	// the knight shuffle is the same as right after 1.e4.
	g := NewGame()
	playSAN(t, g, "e4", "Nf6", "Nf3", "Ng8", "Ng1", "Nf6", "Nf3", "Ng8", "Ng1")
	if g.RepetitionCount() != 3 {
		t.Fatalf("count %d, want 3", g.RepetitionCount())
	}
}

func TestFiftyAndSeventyFiveMoveRules(t *testing.T) {
	g, _ := NewGameFromFEN("8/8/8/4k3/8/8/8/R3K3 w - - 99 80")
	playSAN(t, g, "Ra2")
	if g.ClaimableDraw() != FiftyMoveRule || g.IsOver() {
		t.Fatalf("50-move draw should be claimable, got %v", g.ClaimableDraw())
	}
	g, _ = NewGameFromFEN("8/8/8/4k3/8/8/8/R3K3 w - - 149 100")
	playSAN(t, g, "Ra2")
	if g.Termination() != SeventyFiveMoveRule {
		t.Fatalf("75-move rule should end the game, got %v", g.Termination())
	}
}

func TestCheckmateBeatsSeventyFive(t *testing.T) {
	g, _ := NewGameFromFEN("7k/8/6K1/8/8/8/8/R7 w - - 149 100")
	playSAN(t, g, "Ra8#")
	if g.Termination() != Checkmate {
		t.Fatalf("mate on the 75th move must stand, got %v", g.Termination())
	}
}

func TestResignAgreeAndFlag(t *testing.T) {
	g := NewGame()
	g.Resign(White)
	if g.Outcome() != BlackWins || g.Termination() != Resignation {
		t.Fatal("resign")
	}
	g = NewGame()
	g.AgreeDraw()
	if g.Outcome() != Draw || g.Termination() != DrawAgreement {
		t.Fatal("agree")
	}
	g = NewGame()
	g.FlagFall(Black)
	if g.Outcome() != WhiteWins || g.Termination() != Timeout {
		t.Fatal("flag")
	}
	// White flags but black has only king + knight: draw.
	g, _ = NewGameFromFEN("8/8/8/4k3/4n3/8/4P3/4K3 w - - 0 1")
	g.FlagFall(White)
	if g.Outcome() != Draw || g.Termination() != TimeoutVsInsufficientMaterial {
		t.Fatalf("got %v / %v", g.Outcome(), g.Termination())
	}
}

func TestCapturedPiecesAndMaterial(t *testing.T) {
	g := NewGame()
	playSAN(t, g, "e4", "d5", "exd5", "Qxd5", "Nc3", "Qxd2+", "Bxd2")
	white, black := g.Captured(White), g.Captured(Black)
	if want := []Piece{MakePiece(Black, Pawn), MakePiece(Black, Queen)}; !slices.Equal(white, want) {
		t.Fatalf("white captured %v, want %v", white, want)
	}
	if want := []Piece{MakePiece(White, Pawn), MakePiece(White, Pawn)}; !slices.Equal(black, want) {
		t.Fatalf("black captured %v, want %v", black, want)
	}
	pos := g.Position()
	if diff := pos.Material(White) - pos.Material(Black); diff != 8 {
		t.Fatalf("material diff %d, want 8", diff)
	}

	g = NewGame()
	playSAN(t, g, "e4", "a6", "e5", "d5", "exd6")
	if got := g.Captured(White); len(got) != 1 || got[0] != MakePiece(Black, Pawn) {
		t.Fatalf("en passant capture not recorded: %v", got)
	}
}
