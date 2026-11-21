package main

import (
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

// fontSet draws text cells: a monospace face with fallbacks for symbols.
type fontSet struct {
	cellW, cellH int
	regular      []face // fallback chain
	bold         []face
}

type face struct {
	sf *sfnt.Font
	f  font.Face
}

const fontDir = "/usr/share/fonts/truetype/"

// loadFace opens a TrueType font at the given size.
func loadFace(path string, size float64) face {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	sf, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}
	f, err := opentype.NewFace(sf, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic(err)
	}
	return face{sf, f}
}

// newFontSet builds faces for a cell cellW pixels wide (cell height is
// twice that, so half-block pixels come out square).
func newFontSet(cellW int) *fontSet {
	size := float64(cellW) / 0.602 // DejaVu Sans Mono advance is 0.602 em
	sym := size * 1.12             // DejaVu Sans chess glyphs run small
	return &fontSet{
		cellW: cellW, cellH: 2 * cellW,
		regular: []face{
			loadFace(fontDir+"dejavu/DejaVuSansMono.ttf", size),
			loadFace(fontDir+"dejavu/DejaVuSans.ttf", sym),
			loadFace(fontDir+"noto/NotoSansSymbols2-Regular.ttf", sym),
		},
		bold: []face{
			loadFace(fontDir+"dejavu/DejaVuSansMono-Bold.ttf", size),
			loadFace(fontDir+"dejavu/DejaVuSans-Bold.ttf", sym),
			loadFace(fontDir+"noto/NotoSansSymbols2-Regular.ttf", sym),
		},
	}
}

// faceFor returns the first face in the chain that has a glyph for r.
func (fs *fontSet) faceFor(r rune, bold bool) font.Face {
	chain := fs.regular
	if bold {
		chain = fs.bold
	}
	var buf sfnt.Buffer
	for _, f := range chain {
		if idx, err := f.sf.GlyphIndex(&buf, r); err == nil && idx != 0 {
			return f.f
		}
	}
	return chain[0].f
}
