package main

import (
	"image/color"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Terminal colors used when the output does not set one.
var (
	termFg = color.RGBA{0xC9, 0xD1, 0xD9, 0xFF}
	termBg = color.RGBA{0x0D, 0x11, 0x17, 0xFF}
)

type cell struct {
	r      rune
	fg, bg color.RGBA
	bold   bool
}

// parseANSI turns rendered terminal output into a w x h grid of cells. It
// understands the subset of SGR the interface emits (reset, bold, 24-bit
// foreground and background).
func parseANSI(s string, w, h int) [][]cell {
	grid := make([][]cell, h)
	for y := range grid {
		grid[y] = make([]cell, w)
		for x := range grid[y] {
			grid[y][x] = cell{r: ' ', fg: termFg, bg: termBg}
		}
	}
	fg, bg, bold := termFg, termBg, false
	x, y := 0, 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
				j++
			}
			if j < len(s) && s[j] == 'm' {
				fg, bg, bold = applySGR(s[i+2:j], fg, bg, bold)
			}
			i = j + 1
			continue
		}
		if s[i] == '\n' {
			x, y = 0, y+1
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if y < h && x < w {
			grid[y][x] = cell{r: r, fg: fg, bg: bg, bold: bold}
		}
		x++
	}
	return grid
}

// applySGR updates the pen with one SGR parameter list.
func applySGR(params string, fg, bg color.RGBA, bold bool) (color.RGBA, color.RGBA, bool) {
	p := strings.Split(params, ";")
	for i := 0; i < len(p); i++ {
		n, _ := strconv.Atoi(p[i])
		switch {
		case p[i] == "" || n == 0:
			fg, bg, bold = termFg, termBg, false
		case n == 1:
			bold = true
		case n == 22:
			bold = false
		case n == 39:
			fg = termFg
		case n == 49:
			bg = termBg
		case (n == 38 || n == 48) && i+4 < len(p) && p[i+1] == "2":
			r, _ := strconv.Atoi(p[i+2])
			g, _ := strconv.Atoi(p[i+3])
			b, _ := strconv.Atoi(p[i+4])
			c := color.RGBA{uint8(r), uint8(g), uint8(b), 0xFF}
			if n == 38 {
				fg = c
			} else {
				bg = c
			}
			i += 4
		}
	}
	return fg, bg, bold
}
