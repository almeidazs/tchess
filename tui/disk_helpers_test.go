package tui

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
)

// fakeTime is a clock tests move by hand.
type fakeTime struct{ t time.Time }

func (f *fakeTime) now() time.Time          { return f.t }
func (f *fakeTime) advance(d time.Duration) { f.t = f.t.Add(d) }

// newDiskModel creates a model that keeps its games under dir.
func newDiskModel(t *testing.T, dir string, clk *fakeTime, cfg config.Config) *Model {
	t.Helper()
	m := New(Options{
		Config:     cfg,
		Now:        clk.now,
		GamesDir:   filepath.Join(dir, "games"),
		ResumePath: filepath.Join(dir, "current.pgn"),
	})
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	m.animate = false
	return m
}

// clickRun clicks a square and runs the commands it returns, so files are
// written before the test looks at them.
func clickRun(m *Model, sq chess.Square) {
	x, y := m.SquareCenter(sq)
	_, cmd := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	runCmd(m, cmd)
	_, cmd = m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
	runCmd(m, cmd)
}

// moveRun plays a move given by its squares.
func moveRun(m *Model, from, to chess.Square) {
	clickRun(m, from)
	clickRun(m, to)
}

// teaSize is a window-size message.
func teaSize(w, h int) tea.WindowSizeMsg { return tea.WindowSizeMsg{Width: w, Height: h} }

// clockNone is the time control of a game without a clock.
func clockNone() clock.TimeControl { return clock.TimeControl{} }
