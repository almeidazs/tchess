package main

import (
	"image"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const frameStep = 50 * time.Millisecond

// 20 fps

type frame struct {
	img   *image.RGBA
	delay time.Duration
}

type recorder struct {
	s      *session
	fs     *fontSet
	frames []frame
}

// snap records the current screen, shown for hold.
func (r *recorder) snap(hold time.Duration) {
	img := r.fs.window(r.fs.renderGrid(parseANSI(r.s.view(), r.s.w, r.s.h)), title, false)
	r.frames = append(r.frames, frame{img, hold})
}

// animate advances the fake clock frame by frame through a move animation.
func (r *recorder) animate(d time.Duration) {
	for t := time.Duration(0); t < d; t += frameStep {
		r.s.now = r.s.now.Add(frameStep)
		r.snap(frameStep)
	}
}

// clickMove plays by clicking the piece and then its destination.
func (r *recorder) clickMove(san string) {
	mv := r.s.resolve(san)
	r.s.now = r.s.now.Add(700 * time.Millisecond)
	r.s.click(mv.From())
	r.snap(700 * time.Millisecond) // selection with legal moves
	r.s.click(mv.To())
	if err := r.s.g.Play(mv); err != nil {
		fail(err)
	}
	r.animate(800 * time.Millisecond)
	r.snap(900 * time.Millisecond)
}

// dragMove plays by dragging the piece across the board.
func (r *recorder) dragMove(san string) {
	mv := r.s.resolve(san)
	x0, y0 := r.s.m.SquareCenter(mv.From())
	x1, y1 := r.s.m.SquareCenter(mv.To())
	r.s.now = r.s.now.Add(500 * time.Millisecond)
	r.s.mouse(x0, y0, tea.MouseActionPress)
	r.snap(350 * time.Millisecond)
	const steps = 7
	for i := 1; i <= steps; i++ {
		x := x0 + (x1-x0)*i/steps
		y := y0 + (y1-y0)*i/steps
		r.s.mouse(x, y, tea.MouseActionMotion)
		r.s.now = r.s.now.Add(60 * time.Millisecond)
		r.snap(60 * time.Millisecond)
	}
	r.snap(250 * time.Millisecond)
	r.s.mouse(x1, y1, tea.MouseActionRelease)
	if err := r.s.g.Play(mv); err != nil {
		fail(err)
	}
	r.animate(650 * time.Millisecond)
	r.snap(900 * time.Millisecond)
}

// typeMove plays by typing the move after ':', one key at a time.
func (r *recorder) typeMove(san string) {
	mv := r.s.resolve(san)
	r.s.now = r.s.now.Add(500 * time.Millisecond)
	r.s.runes(":")
	r.snap(400 * time.Millisecond)
	for _, c := range san {
		r.s.runes(string(c))
		r.snap(220 * time.Millisecond)
	}
	r.snap(600 * time.Millisecond) // the move is highlighted on the board
	r.s.key(tea.KeyEnter)
	if err := r.s.g.Play(mv); err != nil {
		fail(err)
	}
	r.animate(800 * time.Millisecond)
	r.snap(900 * time.Millisecond)
}

// demo records the animated GIF: choosing a theme in the menu, then moves
// by drag, click and typing, a take-back, a look at earlier moves and the
// help card.
func demo(path string) {
	s := newSession(126, 47, nil)
	r := &recorder{s: s, fs: newFontSet(8)}

	r.snap(1200 * time.Millisecond) // menu
	for range 7 {                   // down to "Tema"
		s.key(tea.KeyDown)
		r.snap(90 * time.Millisecond)
	}
	r.snap(400 * time.Millisecond)
	s.key(tea.KeyEnter) // theme picker
	r.snap(900 * time.Millisecond)
	for _, k := range []tea.KeyType{tea.KeyDown, tea.KeyDown, tea.KeyDown, tea.KeyUp, tea.KeyUp} {
		s.key(k)
		r.snap(650 * time.Millisecond)
	}
	s.key(tea.KeyEnter) // keep Floresta
	r.snap(700 * time.Millisecond)
	for range 5 { // back up to Blitz 5+3
		s.key(tea.KeyUp)
		r.snap(90 * time.Millisecond)
	}
	r.snap(400 * time.Millisecond)
	s.key(tea.KeyEnter)
	r.snap(1000 * time.Millisecond)

	r.dragMove("e4")
	r.clickMove("e5")
	r.typeMove("Nf3")
	r.clickMove("Nc6")
	r.dragMove("Bb5")
	r.typeMove("a6")
	r.dragMove("Bxc6")

	// White asks to take the capture back and Black agrees.
	s.now = s.now.Add(600 * time.Millisecond)
	s.runes("u")
	r.snap(1400 * time.Millisecond)
	s.runes("y")
	if err := s.g.Undo(); err != nil {
		fail(err)
	}
	r.snap(1200 * time.Millisecond)
	r.dragMove("Ba4")

	// A look back at the opening, then back to the game.
	for range 4 {
		s.runes("[")
		r.snap(600 * time.Millisecond)
	}
	r.snap(400 * time.Millisecond)
	s.key(tea.KeyEsc)
	r.snap(800 * time.Millisecond)

	s.runes("?")
	r.snap(2500 * time.Millisecond)
	s.runes("?")
	r.snap(1500 * time.Millisecond)

	writeGIF(path, r.frames)
}
