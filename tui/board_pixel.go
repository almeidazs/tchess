package tui

import (
	"math"

	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/internal/pixel"
)

// pixelBoard draws the board as sprites on a canvas, one line per terminal
// row (without labels).
func (m *Model) pixelBoard(l layout) []string {
	return m.pixelCanvas(l).Lines(lipgloss.ColorProfile())
}

// pixelCanvas paints the board, highlights and pieces (or the current
// animation frame) into a framebuffer.
func (m *Model) pixelCanvas(l layout) *pixel.Canvas {
	th := m.theme()
	sw, sh := l.cellW, 2*l.cellH // square size in pixels
	cv := pixel.NewCanvas(8*sw, 8*sh)
	if m.anim != nil {
		for row := range 8 {
			for col := range 8 {
				cv.Fill(col*sw, row*sh, sw, sh, shade((col+row)%2 == 0, th.Light, th.Dark))
			}
		}
		m.drawAnimation(cv, sw, sh)
		return cv
	}

	bs := m.boardState(l)
	sprites := spriteSets[sw]
	small := float64(min(sw, sh))
	cx, cy := float64(sw-1)/2, float64(sh-1)/2
	for row := range 8 {
		for col := range 8 {
			lk := m.squareLook(&bs, col, row)
			x0, y0 := col*sw, row*sh
			mark := pixel.Blend(lk.bg, th.Mark, 0.45)
			for y := range sh {
				for x := range sw {
					c := lk.bg
					d := math.Hypot(float64(x)-cx, float64(y)-cy)
					switch {
					case lk.check:
						c = pixel.Blend(checkRGB, lk.base, d/(0.75*float64(max(sw, sh))))
					case lk.dot && d <= small/6+0.1:
						c = mark
					case lk.capture && float64(min(x, sw-1-x)+min(y, sh-1-y)) < small/4:
						c = mark
					}
					cv.Set(x0+x, y0+y, c)
				}
			}
			switch {
			case lk.ghost:
				sprites[lk.piece.Type()].drawAlpha(cv, x0, y0, lk.piece.Color(), 0.3)
			case lk.piece != chess.NoPiece:
				sprites[lk.piece.Type()].draw(cv, x0, y0, lk.piece.Color())
			}
			if lk.cursor {
				for x := range sw {
					cv.Set(x0+x, y0, th.Cursor)
					cv.Set(x0+x, y0+sh-1, th.Cursor)
				}
				for y := range sh {
					cv.Set(x0, y0+y, th.Cursor)
					cv.Set(x0+sw-1, y0+y, th.Cursor)
				}
			}
		}
	}
	if d := bs.drag; d.active {
		// The dragged piece follows the pointer, centered on its cell.
		x := d.x - sw/2
		y := 2*d.y + 1 - sh/2
		sprites[d.piece.Type()].draw(cv, x, y, d.piece.Color())
	}
	return cv
}
