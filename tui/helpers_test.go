package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
)

func newGame(t *testing.T, w, h int, tc string) *Model {
	t.Helper()
	m := New(Options{Config: config.Default()})
	m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	m.animate = false // interaction tests don't wait for animations
	parsed, err := clock.Parse(tc)
	if err != nil {
		t.Fatal(err)
	}
	m.startGame(parsed)
	return m
}

// click presses the left mouse button on the center of a square's cell.
func click(m *Model, sq chess.Square) {
	l := computeLayout(m.width, m.height)
	col, row := m.cellOf(sq)
	m.Update(tea.MouseMsg{
		X: l.boardX + col*l.cellW + l.cellW/2, Y: l.boardY + row*l.cellH + l.cellH/2,
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})
}

func key(m *Model, k string) {
	var msg tea.KeyMsg
	switch k {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	m.Update(msg)
}

// mouseAt sends a mouse event on the center of a square.
func mouseAt(m *Model, sq chess.Square, action tea.MouseAction) {
	l := computeLayout(m.width, m.height)
	col, row := m.cellOf(sq)
	m.Update(tea.MouseMsg{
		X: l.boardX + col*l.cellW + l.cellW/2, Y: l.boardY + row*l.cellH + l.cellH/2,
		Action: action, Button: tea.MouseButtonLeft,
	})
}

func drag(m *Model, path ...chess.Square) {
	mouseAt(m, path[0], tea.MouseActionPress)
	for _, sq := range path[1:] {
		mouseAt(m, sq, tea.MouseActionMotion)
	}
	mouseAt(m, path[len(path)-1], tea.MouseActionRelease)
}

func cursorSquare(m *Model) chess.Square { return m.squareAt(m.cursorCol, m.cursorRow) }
