package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
)

func playOpening(m *Model) {
	for _, sq := range []chess.Square{chess.E2, chess.E4, chess.E7, chess.E5, chess.G1, chess.F3, chess.B8, chess.C6} {
		click(m, sq)
	}
}

func TestReviewKeys(t *testing.T) {
	m := newGame(t, 140, 50, "5+3")
	playOpening(m)
	l := computeLayout(m.width, m.height)

	key(m, "[")
	if !m.reviewing || m.reviewPly != 3 {
		t.Fatalf("[ should show the previous move: reviewing %v ply %d", m.reviewing, m.reviewPly)
	}
	if bs := m.boardState(l); bs.pos != m.game.PositionAt(3) || bs.last.UCI() != "g1f3" {
		t.Fatalf("board shows %s, last %v", bs.pos.FEN(), bs.last)
	}
	if !strings.Contains(m.statusView(), "2.Nf3") || !strings.Contains(m.statusView(), "3 of 4") {
		t.Fatalf("status %q", m.statusView())
	}
	key(m, "home")
	if m.reviewPly != 0 || !strings.Contains(m.statusView(), "starting position") {
		t.Fatalf("home: ply %d", m.reviewPly)
	}
	key(m, "[") // already at the start
	if m.reviewPly != 0 {
		t.Fatal("[ at the start stays there")
	}
	key(m, "]")
	if m.reviewPly != 1 {
		t.Fatalf("] : ply %d", m.reviewPly)
	}

	// Past positions cannot be played on.
	click(m, chess.D2)
	if m.reviewing || m.game.Ply() != 4 {
		t.Fatal("a click on the board returns to the live position without playing")
	}
	key(m, "home")
	for _, k := range []string{"enter", "d", "r", ":"} {
		key(m, k)
	}
	if m.mode != modePlay || m.game.Ply() != 4 || m.selected != chess.NoSquare {
		t.Fatalf("keys leaked into play: mode %v", m.mode)
	}
	key(m, "end")
	if m.reviewing {
		t.Fatal("end returns to the live position")
	}
	key(m, "[")
	key(m, "esc")
	if m.reviewing {
		t.Fatal("esc returns to the live position")
	}
}

func TestReviewByClickingHistory(t *testing.T) {
	m := newGame(t, 140, 50, "5+3")
	playOpening(m)
	l := computeLayout(m.width, m.height)
	top, _ := m.historyRows(l)
	x := l.boardX + 8*l.cellW + gapW + 6
	y := l.boardY + top + 2 + 1 // second line: Black's 1...e5
	m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if !m.reviewing || m.reviewPly != 2 {
		t.Fatalf("clicking 1...e5 should show ply 2, got reviewing %v ply %d", m.reviewing, m.reviewPly)
	}
	m.Update(tea.MouseMsg{X: x, Y: y + 2, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if m.reviewing {
		t.Fatal("clicking the last move returns to the live position")
	}
}

func TestReviewAfterGameOver(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	for _, sq := range []chess.Square{chess.F2, chess.F3, chess.E7, chess.E5, chess.G2, chess.G4, chess.D8, chess.H4} {
		click(m, sq)
	}
	key(m, "[")
	if !m.reviewing || strings.Contains(m.View(), "new game  ·  q quit\n") {
		t.Fatal("the result card should step aside while reviewing")
	}
	key(m, "esc")
	key(m, "n")
	if m.screen != screenMenu {
		t.Fatal("n still starts over after reviewing")
	}
}
