package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
)

func TestViewRenders(t *testing.T) {
	small := newGame(t, 40, 10, "5+3")
	if !strings.Contains(ansi.Strip(small.View()), "Terminal pequeno") {
		t.Error("tiny terminal should show a size warning")
	}
	for _, size := range [][2]int{{140, 45}, {80, 24}, {58, 14}} {
		m := newGame(t, size[0], size[1], "5+3")
		click(m, chess.E2)
		out := ansi.Strip(m.View())
		if !strings.Contains(out, "Brancas") || !strings.Contains(out, "5:00") {
			t.Errorf("%v: view missing pieces of UI:\n%s", size, out)
		}
		if testing.Verbose() {
			t.Logf("%dx%d:\n%s", size[0], size[1], out)
		}
	}
}

func TestGameOverOverlay(t *testing.T) {
	m := newGame(t, 120, 40, "none")
	for _, mv := range [][2]chess.Square{{chess.F2, chess.F3}, {chess.E7, chess.E5}, {chess.G2, chess.G4}, {chess.D8, chess.H4}} {
		click(m, mv[0])
		click(m, mv[1])
	}
	out := ansi.Strip(m.View())
	if m.mode != modeGameOver || !strings.Contains(out, "Pretas vencem") || !strings.Contains(out, "xeque-mate") {
		t.Fatalf("game over view:\n%s", out)
	}
	if testing.Verbose() {
		t.Logf("\n%s", out)
	}
}

func BenchmarkViewLargeBoard(b *testing.B) {
	m := New(Options{})
	m.Update(tea.WindowSizeMsg{Width: 160, Height: 50})
	m.startGame(clock.TimeControl{Base: 5 * time.Minute})
	b.ReportAllocs()
	for b.Loop() {
		_ = m.View()
	}
}
