package tui

import (
	"github.com/almeidazs/tchess/chess"
)

// SquareCenter returns the terminal cell at the center of sq as currently
// drawn, for driving the interface programmatically (tests, recordings).
func (m *Model) SquareCenter(sq chess.Square) (x, y int) {
	l := computeLayout(m.width, m.height)
	col, row := m.cellOf(sq)
	return l.boardX + col*l.cellW + l.cellW/2, l.boardY + row*l.cellH + l.cellH/2
}

// squareAt maps a screen cell to a board square for the current orientation.
func (m *Model) squareAt(col, row int) chess.Square {
	if m.orientation == chess.White {
		return chess.NewSquare(col, 7-row)
	}
	return chess.NewSquare(7-col, row)
}

// cellOf maps a board square to its screen cell.
func (m *Model) cellOf(sq chess.Square) (col, row int) {
	if m.orientation == chess.White {
		return sq.File(), 7 - sq.Rank()
	}
	return 7 - sq.File(), sq.Rank()
}

// cellIn is cellOf for an explicit orientation.
func cellIn(orient chess.Color, sq chess.Square) (col, row int) {
	if orient == chess.White {
		return sq.File(), 7 - sq.Rank()
	}
	return 7 - sq.File(), sq.Rank()
}
