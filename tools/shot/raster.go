package main

import (
	"image"
	"image/color"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// renderGrid rasterizes a cell grid.
func (fs *fontSet) renderGrid(grid [][]cell) *image.RGBA {
	h, w := len(grid), len(grid[0])
	img := image.NewRGBA(image.Rect(0, 0, w*fs.cellW, h*fs.cellH))
	for y, row := range grid {
		for x, c := range row {
			fs.drawCell(img, x*fs.cellW, y*fs.cellH, c)
		}
	}
	return img
}

// drawCell paints one terminal cell: block characters and box drawing
// geometrically, everything else with the font.
func (fs *fontSet) drawCell(img *image.RGBA, x0, y0 int, c cell) {
	cw, ch := fs.cellW, fs.cellH
	fillRect(img, x0, y0, x0+cw, y0+ch, c.bg)
	switch {
	case c.r == ' ':
	case c.r == '▀':
		fillRect(img, x0, y0, x0+cw, y0+ch/2, c.fg)
	case c.r == '█':
		fillRect(img, x0, y0, x0+cw, y0+ch, c.fg)
	case c.r >= '▉' && c.r <= '▏': // left eighth blocks: ▉ = 7/8 ... ▏ = 1/8
		eighths := int('▏'-c.r) + 1
		fillRect(img, x0, y0, x0+cw*eighths/8, y0+ch, c.fg)
	case strings.ContainsRune("─│╭╮╰╯", c.r):
		fs.drawBox(img, x0, y0, c)
	default:
		f := fs.faceFor(c.r, c.bold)
		m := f.Metrics()
		adv := font.MeasureString(f, string(c.r))
		baseline := fixed.I(y0) + (fixed.I(ch)-(m.Ascent+m.Descent))/2 + m.Ascent
		d := font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(c.fg),
			Face: f,
			Dot:  fixed.Point26_6{X: fixed.I(x0) + (fixed.I(cw)-adv)/2, Y: baseline},
		}
		d.DrawString(string(c.r))
	}
}

// drawBox draws box-drawing characters geometrically so lines join
// seamlessly across cells; corners are quarter circles.
func (fs *fontSet) drawBox(img *image.RGBA, x0, y0 int, c cell) {
	cw, ch := fs.cellW, fs.cellH
	t := max(cw/10, 1) // line thickness
	mx, my := x0+cw/2, y0+ch/2
	hline := func(xa, xb int) { fillRect(img, xa, my-t/2, xb, my-t/2+t, c.fg) }
	vline := func(ya, yb int) { fillRect(img, mx-t/2, ya, mx-t/2+t, yb, c.fg) }
	r := cw / 2
	arc := func(cx, cy float64, a0 float64) {
		for i := range 65 {
			a := a0 + math.Pi/2*float64(i)/64
			px := int(math.Round(cx + float64(r)*math.Cos(a)))
			py := int(math.Round(cy + float64(r)*math.Sin(a)))
			fillRect(img, px-t/2, py-t/2, px-t/2+t, py-t/2+t, c.fg)
		}
	}
	switch c.r {
	case '─':
		hline(x0, x0+cw)
	case '│':
		vline(y0, y0+ch)
	case '╭':
		arc(float64(mx+r), float64(my+r), math.Pi)
		hline(mx+r, x0+cw)
		vline(my+r, y0+ch)
	case '╮':
		arc(float64(mx-r), float64(my+r), -math.Pi/2)
		hline(x0, mx-r)
		vline(my+r, y0+ch)
	case '╰':
		arc(float64(mx+r), float64(my-r), math.Pi/2)
		hline(mx+r, x0+cw)
		vline(y0, my-r)
	case '╯':
		arc(float64(mx-r), float64(my-r), 0)
		hline(x0, mx-r)
		vline(y0, my-r)
	}
}

// fillRect paints the rectangle [x0, x1) x [y0, y1).
func fillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// mix blends a toward b by t.
func mix(a, b color.RGBA, t float64) color.RGBA {
	l := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.RGBA{l(a.R, b.R), l(a.G, b.G), l(a.B, b.B), 0xFF}
}
