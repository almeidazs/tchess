package tui

import (
	"github.com/almeidazs/chess/chess"
)

// The keyboard cursor only stops on squares where Enter does something
// useful: with a piece selected, its legal destinations (and the piece
// itself, to step back to it); otherwise, our pieces that can move.

// cursorStops returns the screen cells the cursor may stop on.
func (m *Model) cursorStops() []cell {
	seen := map[cell]bool{}
	var out []cell
	add := func(sq chess.Square) {
		c, r := m.cellOf(sq)
		if k := (cell{c, r}); !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	if m.selected != chess.NoSquare {
		add(m.selected)
		for _, mv := range m.targets {
			add(mv.To())
		}
		return out
	}
	for _, mv := range m.game.LegalMoves() {
		add(mv.From())
	}
	return out
}

// revealCursor shows the cursor, snapping it to the nearest valid stop.
func (m *Model) revealCursor() {
	m.showCursor = true
	m.snapCursor(m.cursorStops())
}

// moveCursor moves to the nearest stop in the direction (dc, dr). The first
// key press only reveals the cursor.
func (m *Model) moveCursor(dc, dr int) {
	if !m.showCursor {
		m.revealCursor()
		return
	}
	stops := m.cursorStops()
	if next, ok := nextStop(stops, cell{m.cursorCol, m.cursorRow}, dc, dr); ok {
		m.cursorCol, m.cursorRow = next.col, next.row
		return
	}
	m.snapCursor(stops) // nothing that way: at least stand on a valid stop
}

// snapCursor moves the cursor to the nearest stop unless it is on one.
func (m *Model) snapCursor(stops []cell) {
	cur := cell{m.cursorCol, m.cursorRow}
	best, bestD := cur, -1
	for _, s := range stops {
		if s == cur {
			return
		}
		if d := sq(s.col-cur.col) + sq(s.row-cur.row); bestD < 0 || d < bestD {
			best, bestD = s, d
		}
	}
	m.cursorCol, m.cursorRow = best.col, best.row
}

// nextStop picks the stop ahead in direction (dc, dr) that is closest,
// weighting sideways distance double so the cursor favors staying in line.
func nextStop(stops []cell, from cell, dc, dr int) (cell, bool) {
	var best cell
	bestScore, found := 0, false
	for _, s := range stops {
		vx, vy := s.col-from.col, s.row-from.row
		along := vx*dc + vy*dr
		if along <= 0 {
			continue
		}
		side := abs(vx*dr - vy*dc)
		score := along + 2*side
		if !found || score < bestScore || (score == bestScore && side < abs((best.col-from.col)*dr-(best.row-from.row)*dc)) {
			best, bestScore, found = s, score, true
		}
	}
	return best, found
}
