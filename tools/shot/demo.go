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

// demo records the animated GIF: menu, then a few moves by drag and click.
func demo(path string) {
	s := newSession(126, 47, nil)
	r := &recorder{s: s, fs: newFontSet(8)}

	r.snap(1200 * time.Millisecond) // menu
	s.key(tea.KeyDown)
	r.snap(300 * time.Millisecond)
	s.key(tea.KeyDown)
	r.snap(500 * time.Millisecond)
	s.key(tea.KeyEnter) // Blitz 5+3
	r.snap(1000 * time.Millisecond)

	r.dragMove("e4")
	r.clickMove("e5")
	r.dragMove("Nf3")
	r.clickMove("Nc6")
	r.dragMove("Bb5")
	r.clickMove("a6")
	r.dragMove("Bxc6")
	r.clickMove("dxc6")
	r.frames[len(r.frames)-1].delay = 2500 * time.Millisecond

	writeGIF(path, r.frames)
}
