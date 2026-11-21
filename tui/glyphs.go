package tui

import (
	"github.com/almeidazs/tchess/chess"
)

// Classic diagram glyphs: outlined for White, filled for Black, both drawn in
// dark ink so they read well on light and dark squares alike.
var glyphs = [2][6]string{
	{"♙", "♘", "♗", "♖", "♕", "♔"},
	{"♟", "♞", "♝", "♜", "♛", "♚"},
}

// glyph returns the Unicode symbol of a piece.
func glyph(p chess.Piece) string { return glyphs[p.Color()][p.Type()] }

// kingGlyph is used next to player names.
func kingGlyph(c chess.Color) string { return glyphs[c][chess.King] }
