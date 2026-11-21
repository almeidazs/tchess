package tui

const (
	rankLabelW = 3
	gapW       = 3
	panelW     = 28
	headerH    = 2
	footerH    = 3
)

// boardSizes are tried from largest to smallest. Pixel sizes draw sprites of
// cellW x 2*cellH pixels (each terminal row holds two half-block pixels);
// the last size falls back to one Unicode glyph per square.
var boardSizes = []struct{ cellW, cellH int }{{11, 5}, {9, 4}, {7, 3}, {3, 1}}

// computeLayout picks the largest board that fits a w x h terminal and
// centers the game screen.
func computeLayout(w, h int) layout {
	var l layout
	for _, s := range boardSizes {
		l = layout{cellW: s.cellW, cellH: s.cellH, pixel: s.cellH > 1}
		l.totalW = rankLabelW + 8*l.cellW + gapW + panelW
		l.totalH = headerH + l.boardH() + footerH
		if l.fits(w, h) {
			break
		}
	}
	l.originX, l.originY = centerOffset(w, h, l.totalW, l.totalH)
	l.boardX = l.originX + rankLabelW
	l.boardY = l.originY + headerH
	return l
}

// boardH is the board block height: 8 rows plus the file labels.
func (l layout) boardH() int { return 8*l.cellH + 1 }

// compact reports the smallest layout, which uses one-line clocks.
func (l layout) compact() bool { return l.cellH == 1 }

// fits reports whether the terminal is large enough for this layout.
func (l layout) fits(w, h int) bool { return l.totalW <= w && l.totalH <= h }

// cellAt maps a terminal position to a board cell.
func (l layout) cellAt(x, y int) (col, row int, ok bool) {
	dx, dy := x-l.boardX, y-l.boardY
	if dx < 0 || dy < 0 {
		return 0, 0, false
	}
	col, row = dx/l.cellW, dy/l.cellH
	return col, row, col < 8 && row < 8
}

// centerOffset returns the top-left corner that centers a cw x ch block.
func centerOffset(w, h, cw, ch int) (int, int) {
	return max((w-cw)/2, 0), max((h-ch)/2, 0)
}
