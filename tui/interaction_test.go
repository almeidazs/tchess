package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/almeidazs/tchess/chess"
)

func TestClickToMoveAndFlip(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}} {
		m := newGame(t, size[0], size[1], "none")
		click(m, chess.E2)
		if m.selected != chess.E2 || len(m.targets) != 2 {
			t.Fatalf("%v: select e2: selected=%v targets=%v", size, m.selected, m.targets)
		}
		click(m, chess.E4)
		pos := m.game.Position()
		if pos.PieceAt(chess.E4) != chess.MakePiece(chess.White, chess.Pawn) {
			t.Fatalf("%v: e4 not played", size)
		}
		if m.orientation != chess.Black {
			t.Fatalf("%v: board should flip to black", size)
		}
		// With the board flipped, clicking where e7/e5 are must still work.
		click(m, chess.E7)
		click(m, chess.E5)
		if h := m.game.History(); len(h) != 2 || h[1].SAN != "e5" {
			t.Fatalf("%v: history %v", size, h)
		}
	}
}

func TestClickIllegalTargetDeselects(t *testing.T) {
	m := newGame(t, 120, 40, "none")
	click(m, chess.E2)
	click(m, chess.E5)
	if m.selected != chess.NoSquare || len(m.game.History()) != 0 {
		t.Fatal("illegal target should just clear the selection")
	}
	click(m, chess.E7) // opponent piece: not selectable
	if m.selected != chess.NoSquare {
		t.Fatal("selected an opponent piece")
	}
}

func TestClickSelectedPieceDeselects(t *testing.T) {
	m := newGame(t, 160, 50, "none")
	mouseAt(m, chess.E2, tea.MouseActionPress)
	mouseAt(m, chess.E2, tea.MouseActionRelease)
	if m.selected != chess.E2 {
		t.Fatal("first click selects")
	}
	mouseAt(m, chess.E2, tea.MouseActionPress)
	mouseAt(m, chess.E2, tea.MouseActionRelease)
	if m.selected != chess.NoSquare {
		t.Fatal("second click on the same piece deselects")
	}
	// Click-click still moves.
	mouseAt(m, chess.E2, tea.MouseActionPress)
	mouseAt(m, chess.E2, tea.MouseActionRelease)
	mouseAt(m, chess.E4, tea.MouseActionPress)
	mouseAt(m, chess.E4, tea.MouseActionRelease)
	if len(m.game.History()) != 1 {
		t.Fatal("click-click move broke")
	}
}

func TestKeyboardMove(t *testing.T) {
	m := newGame(t, 120, 40, "none")
	key(m, "up") // first key just reveals the cursor at e2
	if !m.showCursor || m.squareAt(m.cursorCol, m.cursorRow) != chess.E2 {
		t.Fatal("cursor should start on e2")
	}
	key(m, "enter")
	key(m, "up")
	key(m, "up")
	key(m, "enter")
	if h := m.game.History(); len(h) != 1 || h[0].SAN != "e4" {
		t.Fatalf("history %v", h)
	}
}

func TestPromotionPicker(t *testing.T) {
	m := newGame(t, 120, 40, "none")
	g, err := chess.NewGameFromFEN("1n2k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	m.game = g

	click(m, chess.A7)
	click(m, chess.B8)
	if m.mode != modePromotion || len(m.promo) != 4 {
		t.Fatalf("expected promotion picker, mode=%v promo=%v", m.mode, m.promo)
	}
	key(m, "esc")
	if m.mode != modePlay || len(m.game.History()) != 0 {
		t.Fatal("esc must cancel the promotion")
	}

	click(m, chess.A7)
	click(m, chess.B8)
	// The picker stacks Q R B N downward from b8; the knight is 3 cells below.
	l := computeLayout(m.width, m.height)
	col, row := m.cellOf(chess.B8)
	m.Update(tea.MouseMsg{X: l.boardX + col*l.cellW + 1, Y: l.boardY + (row+3)*l.cellH + 1,
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	pos := m.game.Position()
	if pos.PieceAt(chess.B8) != chess.MakePiece(chess.White, chess.Knight) {
		t.Fatalf("expected knight promotion, got %s", pos.FEN())
	}
}

func TestResignAndDrawFlows(t *testing.T) {
	m := newGame(t, 120, 40, "none")
	key(m, "r")
	key(m, "n")
	if m.game.IsOver() {
		t.Fatal("declined resignation ended the game")
	}
	key(m, "r")
	key(m, "y")
	if m.game.Outcome() != chess.BlackWins || m.mode != modeGameOver {
		t.Fatal("white resignation")
	}
	key(m, "n") // new game -> menu
	if m.screen != screenMenu {
		t.Fatal("n should return to the menu")
	}

	m = newGame(t, 120, 40, "none")
	key(m, "d")
	if m.mode != modeDrawOffer {
		t.Fatal("d without a claim should offer a draw")
	}
	key(m, "y")
	if m.game.Termination() != chess.DrawAgreement {
		t.Fatal("accepted draw")
	}
}

func TestClaimThreefoldWithKey(t *testing.T) {
	m := newGame(t, 120, 40, "none")
	for range 2 {
		for _, mv := range [][2]chess.Square{{chess.G1, chess.F3}, {chess.G8, chess.F6}, {chess.F3, chess.G1}, {chess.F6, chess.G8}} {
			click(m, mv[0])
			click(m, mv[1])
		}
	}
	if !strings.Contains(ansi.Strip(m.View()), "reivindicar empate") {
		t.Fatal("status bar should offer the claim")
	}
	key(m, "d")
	if m.game.Termination() != chess.ThreefoldRepetition {
		t.Fatalf("got %v", m.game.Termination())
	}
}
