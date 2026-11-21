package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
)

func TestDragAndDrop(t *testing.T) {
	for _, size := range [][2]int{{160, 50}, {120, 40}, {95, 32}, {70, 20}} {
		m := newGame(t, size[0], size[1], "none")
		drag(m, chess.G1, chess.G2, chess.F3)
		if h := m.game.History(); len(h) != 1 || h[0].SAN != "Nf3" {
			t.Fatalf("%v: drag g1-f3: history %v", size, h)
		}
		// Board flipped: dragging black's pieces must map correctly too.
		drag(m, chess.D7, chess.D6, chess.D5)
		if h := m.game.History(); len(h) != 2 || h[1].SAN != "d5" {
			t.Fatalf("%v: drag d7-d5: history %v", size, h)
		}
	}
}

func TestDragShowsPieceUnderPointer(t *testing.T) {
	m := newGame(t, 160, 50, "none")
	mouseAt(m, chess.E2, tea.MouseActionPress)
	mouseAt(m, chess.E3, tea.MouseActionMotion)
	l := computeLayout(m.width, m.height)
	dv := m.dragView(l)
	if !dv.active || dv.from != chess.E2 || dv.hover != chess.E3 || dv.piece != chess.MakePiece(chess.White, chess.Pawn) {
		t.Fatalf("drag view %+v", dv)
	}
	mouseAt(m, chess.E5, tea.MouseActionMotion) // not a legal target
	if m.dragView(l).hover != chess.NoSquare {
		t.Fatal("illegal squares must not be highlighted")
	}
	if m.View() == "" {
		t.Fatal("empty frame while dragging")
	}
}

func TestDropOnIllegalSquareSnapsBack(t *testing.T) {
	m := newGame(t, 160, 50, "none")
	drag(m, chess.E2, chess.E3, chess.E5)
	if len(m.game.History()) != 0 {
		t.Fatal("illegal drop must not move")
	}
	if m.selected != chess.E2 || m.drag != nil {
		t.Fatal("piece should snap back and stay selected")
	}
	// Dropping outside the board also snaps back.
	mouseAt(m, chess.E2, tea.MouseActionPress)
	m.Update(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft})
	m.Update(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
	if len(m.game.History()) != 0 || m.selected != chess.E2 {
		t.Fatal("drop outside the board should snap back")
	}
}

func TestDragToPromotion(t *testing.T) {
	m := newGame(t, 160, 50, "none")
	g, _ := chess.NewGameFromFEN("1n2k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	m.game = g
	drag(m, chess.A7, chess.A8)
	if m.mode != modePromotion {
		t.Fatalf("dropping on the last rank should open the picker, mode=%v", m.mode)
	}
	key(m, "esc")
	if m.dropped {
		t.Fatal("canceled promotion must not leave the drop flag set")
	}
	drag(m, chess.A7, chess.B8)
	key(m, "n")
	pos := m.game.Position()
	if pos.PieceAt(chess.B8) != chess.MakePiece(chess.White, chess.Knight) {
		t.Fatalf("promotion by drag: %s", pos.FEN())
	}
}
