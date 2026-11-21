package tui

import (
	"fmt"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/internal/pixel"
)

// sprite is pixel art for one piece, drawn in the piece's colors. Legend:
//
//	.  transparent      #  outline
//	o  body             x  detail (dark on White pieces, light on Black)
//
// The art itself lives in sprite_data.go.
type sprite struct {
	w, h int
	px   []byte
}

// mustSprite builds a sprite from rows of equal width.
func mustSprite(rows ...string) sprite {
	s := sprite{w: len(rows[0]), h: len(rows)}
	for _, r := range rows {
		if len(r) != s.w {
			panic(fmt.Sprintf("sprite row %q is not %d wide", r, s.w))
		}
		s.px = append(s.px, r...)
	}
	return s
}

// piecePalette colors the sprites, in the spirit of printed diagrams: White
// pieces are ivory with a dark outline, Black pieces near-black with light
// details.
var piecePalette = [2]struct{ outline, body, detail pixel.RGB }{
	chess.White: {pixel.Hex("#1B1A18"), pixel.Hex("#FAF6EE"), pixel.Hex("#1B1A18")},
	chess.Black: {pixel.Hex("#0B0A09"), pixel.Hex("#3A3632"), pixel.Hex("#E9E2D5")},
}

// draw paints the sprite with its top-left corner at (x, y).
func (s sprite) draw(cv *pixel.Canvas, x, y int, color chess.Color) {
	s.drawAlpha(cv, x, y, color, 1)
}

// drawAlpha paints the sprite blended over the canvas (alpha in [0, 1]).
func (s sprite) drawAlpha(cv *pixel.Canvas, x, y int, color chess.Color, alpha float64) {
	pal := piecePalette[color]
	for i, ch := range s.px {
		var col pixel.RGB
		switch ch {
		case '#':
			col = pal.outline
		case 'o':
			col = pal.body
		case 'x':
			col = pal.detail
		default:
			continue
		}
		px, py := x+i%s.w, y+i/s.w
		if !cv.Contains(px, py) {
			continue
		}
		if alpha < 1 {
			col = pixel.Blend(cv.At(px, py), col, alpha)
		}
		cv.Set(px, py, col)
	}
}
