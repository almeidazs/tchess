package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
)

func TestUndoNeedsConsent(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	key(m, "u")
	if m.mode != modePlay || !strings.Contains(m.notice, "Nenhum lance") {
		t.Fatalf("undo before any move: mode %v notice %q", m.mode, m.notice)
	}
	click(m, chess.E2)
	click(m, chess.E4)
	key(m, "u")
	if m.mode != modeUndoOffer || !strings.Contains(m.statusView(), "Brancas pedem para voltar o lance") {
		t.Fatalf("mode %v status %q", m.mode, m.statusView())
	}
	key(m, "n")
	if m.game.Ply() != 1 || m.mode != modePlay {
		t.Fatal("a refused take-back keeps the move")
	}
	key(m, "u")
	key(m, "y")
	if m.game.Ply() != 0 || m.orientation != chess.White || len(m.moveTimes) != 0 {
		t.Fatalf("after take-back: ply %d orientation %v times %v", m.game.Ply(), m.orientation, m.moveTimes)
	}
}

func TestUndoRestoresClocks(t *testing.T) {
	clk := &fakeTime{time.Unix(0, 0)}
	m := newDiskModel(t, t.TempDir(), clk, config.Default())
	m.startGame(clock.TimeControl{Base: 5 * time.Minute, Increment: 3 * time.Second})

	click(m, chess.E2) // White's first move starts Black's clock
	click(m, chess.E4)
	clk.advance(5 * time.Second)
	click(m, chess.E7) // Black: 4:55 at the move, 4:58 after it
	click(m, chess.E5)
	clk.advance(2 * time.Second)
	click(m, chess.G1) // White: 4:58 at the move, 5:01 after it
	click(m, chess.F3)
	clk.advance(10 * time.Second) // Black thinks, then accepts White's request

	key(m, "u")
	key(m, "y")
	if got := m.clock.Remaining(chess.White); got != 4*time.Minute+58*time.Second {
		t.Fatalf("white %v, want 4:58 (the increment is taken back)", got)
	}
	if got := m.clock.Remaining(chess.Black); got != 4*time.Minute+58*time.Second {
		t.Fatalf("black %v, want 4:58 (time spent deciding is not charged)", got)
	}
	clk.advance(time.Second)
	if got := m.clock.Remaining(chess.White); got != 4*time.Minute+57*time.Second {
		t.Fatalf("white's time should run again: %v", got)
	}

	// Taking back White's first move stops the clock again.
	key(m, "u") // Black now asks to take back e5
	key(m, "y")
	key(m, "u")
	key(m, "y")
	if m.game.Ply() != 0 || m.clock.Running() {
		t.Fatalf("ply %d running %v", m.game.Ply(), m.clock.Running())
	}
	if m.clock.Remaining(chess.White) != 5*time.Minute || m.clock.Remaining(chess.Black) != 5*time.Minute {
		t.Fatal("both clocks should be full before the first move")
	}
}

func TestUndoBackToReviewedPosition(t *testing.T) {
	clk := &fakeTime{time.Unix(0, 0)}
	m := newDiskModel(t, t.TempDir(), clk, config.Default())
	m.startGame(clock.TimeControl{Base: 5 * time.Minute, Increment: 3 * time.Second})
	for _, mv := range [][2]chess.Square{{chess.E2, chess.E4}, {chess.E7, chess.E5}, {chess.G1, chess.F3}, {chess.B8, chess.C6}, {chess.F1, chess.B5}} {
		clk.advance(4 * time.Second)
		click(m, mv[0])
		click(m, mv[1])
	}
	key(m, "home")
	key(m, "]")
	key(m, "]") // position after 1.e4 e5: White to move
	key(m, "u")
	if m.mode != modeUndoOffer || !strings.Contains(m.statusView(), "Brancas pedem para voltar 3 lances. Pretas, aceitam?") {
		t.Fatalf("mode %v status %q", m.mode, m.statusView())
	}
	if !m.reviewing {
		t.Fatal("the board keeps showing the target position while asking")
	}
	key(m, "y")
	if m.game.Ply() != 2 || m.reviewing || m.game.Turn() != chess.White || len(m.moveTimes) != 2 {
		t.Fatalf("after: ply %d reviewing %v times %v", m.game.Ply(), m.reviewing, m.moveTimes)
	}
	if !strings.Contains(m.notice, "3 lances voltados") {
		t.Fatalf("notice %q", m.notice)
	}
	// White's clock is back to what it showed at the moment of 2.Nf3 (4 s
	// after 1...e5 started it), without the increment Nf3 earned; Black's to
	// what it showed after 1...e5.
	if got, want := m.clock.Remaining(chess.White), 5*time.Minute-4*time.Second; got != want {
		t.Fatalf("white %v, want %v", got, want)
	}
	if got, want := m.clock.Remaining(chess.Black), 5*time.Minute-4*time.Second+3*time.Second; got != want {
		t.Fatalf("black %v, want %v", got, want)
	}
	if !m.clock.Running() {
		t.Fatal("White's time runs again")
	}

	// Refusing keeps everything, still reviewing.
	key(m, "home")
	key(m, "u")
	key(m, "n")
	if m.game.Ply() != 2 || m.mode != modePlay {
		t.Fatal("a refused take-back keeps the game")
	}
	key(m, "end")
	key(m, "u") // live again: one half-move
	if !strings.Contains(m.statusView(), "Pretas pedem para voltar o lance") {
		t.Fatalf("status %q", m.statusView())
	}
}

func TestUndoBeforeTheClockJoined(t *testing.T) {
	pg, err := chess.ParsePGN("1. e4 e5 *")
	if err != nil {
		t.Fatal(err)
	}
	clk := &fakeTime{time.Unix(0, 0)}
	m := New(Options{Config: config.Default(), Import: []*chess.PGNGame{pg}, Now: clk.now})
	m.Update(teaSize(140, 50))
	m.animate = false
	runCmd(m, m.Init())
	m.clocks.cursor = 1 // Blitz 3+2 (no record clocks: Bullet first)
	press(m, "enter")
	clk.advance(10 * time.Second)
	click(m, chess.G1)
	click(m, chess.F3)

	key(m, "home") // back past the moves that had no clock
	key(m, "u")
	key(m, "y")
	if m.game.Ply() != 0 || m.moveClocks != nil || m.clockFrom != 0 {
		t.Fatalf("ply %d clocks %v from %d", m.game.Ply(), m.moveClocks, m.clockFrom)
	}
	if m.clock.Remaining(chess.White) != 3*time.Minute-10*time.Second+2*time.Second {
		t.Fatalf("clocks without a record stay as they were: %v", m.clock.Remaining(chess.White))
	}
	clk.advance(time.Second)
	if m.clock.Active() != chess.White || m.clock.Remaining(chess.Black) != 3*time.Minute {
		t.Fatal("the clock now runs for White, the side to move")
	}
}

func TestUndoNotAfterGameOver(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	for _, sq := range []chess.Square{chess.F2, chess.F3, chess.E7, chess.E5, chess.G2, chess.G4, chess.D8, chess.H4} {
		click(m, sq)
	}
	key(m, "u")
	if m.mode != modeGameOver || m.game.Ply() != 4 {
		t.Fatal("a finished game cannot be taken back")
	}
}
