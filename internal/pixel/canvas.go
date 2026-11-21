// Package pixel is a small framebuffer for terminals. Each terminal cell
// shows two vertically stacked pixels using the upper half block "▀": the
// foreground color paints the top pixel and the background the bottom one.
// With the usual 1:2 cell aspect ratio the pixels come out roughly square.
package pixel

import (
	"strings"
	"sync"

	"github.com/muesli/termenv"
)

// Canvas is a w x h framebuffer, h in pixels (two per terminal row).
type Canvas struct {
	w, h int
	px   []RGB
}

// NewCanvas returns a black canvas. The height should be even; an odd last
// pixel row is not printed.
func NewCanvas(w, h int) *Canvas { return &Canvas{w: w, h: h, px: make([]RGB, w*h)} }

// Width returns the width in pixels.
func (c *Canvas) Width() int { return c.w }

// Height returns the height in pixels.
func (c *Canvas) Height() int { return c.h }

// Contains reports whether (x, y) lies on the canvas.
func (c *Canvas) Contains(x, y int) bool { return x >= 0 && y >= 0 && x < c.w && y < c.h }

// Set paints one pixel; coordinates off the canvas are ignored.
func (c *Canvas) Set(x, y int, col RGB) {
	if c.Contains(x, y) {
		c.px[y*c.w+x] = col
	}
}

// At returns the color of a pixel on the canvas.
func (c *Canvas) At(x, y int) RGB { return c.px[y*c.w+x] }

// Fill paints a rectangle, clipped to the canvas.
func (c *Canvas) Fill(x, y, w, h int, col RGB) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			c.Set(xx, yy, col)
		}
	}
}

// Lines encodes the canvas as h/2 terminal lines for the given color
// profile, emitting color codes only when they change. Rows of two equal
// pixels are printed as colored spaces.
func (c *Canvas) Lines(p termenv.Profile) []string {
	seq := sequencesFor(p)
	out := make([]string, c.h/2)
	var sb strings.Builder
	for row := range out {
		sb.Reset()
		curFg, curBg := "\x00", "\x00"
		top, bot := c.px[2*row*c.w:], c.px[(2*row+1)*c.w:]
		for x := range c.w {
			if bg := seq.get(bot[x], true); bg != curBg {
				sb.WriteString(bg)
				curBg = bg
			}
			if top[x] == bot[x] {
				sb.WriteByte(' ')
				continue
			}
			if fg := seq.get(top[x], false); fg != curFg {
				sb.WriteString(fg)
				curFg = fg
			}
			sb.WriteString("▀")
		}
		sb.WriteString(termenv.CSI + termenv.ResetSeq + "m")
		out[row] = sb.String()
	}
	return out
}

// sequences memoizes SGR escape codes per profile: a board uses only a few
// dozen colors, so encoding is mostly map lookups.
type sequences struct {
	profile termenv.Profile
	mu      sync.Mutex
	cache   map[seqKey]string
}

type seqKey struct {
	color RGB
	bg    bool
}

var (
	profilesMu sync.Mutex
	profiles   = map[termenv.Profile]*sequences{}
)

// sequencesFor returns the shared escape-code cache for a profile.
func sequencesFor(p termenv.Profile) *sequences {
	profilesMu.Lock()
	defer profilesMu.Unlock()
	s, ok := profiles[p]
	if !ok {
		s = &sequences{profile: p, cache: map[seqKey]string{}}
		profiles[p] = s
	}
	return s
}

// get returns the SGR sequence setting c as foreground or background.
func (s *sequences) get(c RGB, bg bool) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := seqKey{c, bg}
	if v, ok := s.cache[k]; ok {
		return v
	}
	v := s.profile.Color(c.Hex()).Sequence(bg)
	if v != "" {
		v = termenv.CSI + v + "m"
	}
	s.cache[k] = v
	return v
}
