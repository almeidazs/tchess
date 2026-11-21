package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/clock"
	"github.com/almeidazs/chess/tui"
)

// session drives a model with a fake clock and mirrors the game to resolve
// moves written in SAN.
type session struct {
	m    *tui.Model
	g    *chess.Game
	now  time.Time
	w, h int
}

// newSession creates a model of the given size, optionally skipping the menu.
func newSession(w, h int, tc *clock.TimeControl) *session {
	s := &session{g: chess.NewGame(), now: time.Date(2026, 1, 1, 20, 0, 0, 0, time.UTC), w: w, h: h}
	s.m = tui.New(tui.Options{TimeControl: tc, Now: func() time.Time { return s.now }})
	s.m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	s.m.Init()
	return s
}

// view renders the current frame.
func (s *session) view() string { return s.m.View() }

// mouse sends a left-button event at (x, y).
func (s *session) mouse(x, y int, a tea.MouseAction) {
	s.m.Update(tea.MouseMsg{X: x, Y: y, Action: a, Button: tea.MouseButtonLeft})
}

// key sends a key press.
func (s *session) key(k tea.KeyType) { s.m.Update(tea.KeyMsg{Type: k}) }

// click presses and releases the mouse on a square.
func (s *session) click(sq chess.Square) {
	x, y := s.m.SquareCenter(sq)
	s.mouse(x, y, tea.MouseActionPress)
	s.mouse(x, y, tea.MouseActionRelease)
}

// resolve finds a SAN move in the mirrored game.
func (s *session) resolve(san string) chess.Move {
	pos := s.g.Position()
	mv, err := pos.ParseSAN(san)
	if err != nil {
		fail(err)
	}
	return mv
}

// play makes a move by clicking origin then destination after "thinking".
func (s *session) play(san string, think time.Duration) {
	mv := s.resolve(san)
	s.now = s.now.Add(think)
	s.click(mv.From())
	s.click(mv.To())
	if err := s.g.Play(mv); err != nil {
		fail(err)
	}
	s.now = s.now.Add(time.Second) // let the turn animation finish
}
