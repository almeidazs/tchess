package tui

import (
	"github.com/almeidazs/tchess/chess"
)

// While reviewing, the board shows the position after reviewPly half-moves
// instead of the live one. The game goes on underneath: the clock keeps
// running, and moves are blocked until the player returns to the present.

// reviewKey handles [ ] home end and esc. It reports whether it used the key.
func (m *Model) reviewKey(key string) bool {
	ply := m.game.Ply()
	at := ply
	if m.reviewing {
		at = m.reviewPly
	}
	switch key {
	case "[":
		m.reviewAt(at - 1)
	case "]":
		m.reviewAt(at + 1)
	case "home":
		m.reviewAt(0)
	case "end":
		m.reviewAt(ply)
	case "esc":
		if !m.reviewing {
			return false
		}
		m.stopReview()
	default:
		return false
	}
	return true
}

// reviewAt shows the position after ply half-moves; the current ply
// returns to the live position.
func (m *Model) reviewAt(ply int) {
	ply = min(max(ply, 0), m.game.Ply())
	if ply == m.game.Ply() {
		m.stopReview()
		return
	}
	m.reviewing, m.reviewPly = true, ply
	m.anim = nil
	m.clearSelection()
	m.notice = ""
}

// stopReview returns to the live position.
func (m *Model) stopReview() {
	m.reviewing = false
}

// shownPosition is the position on the board: the reviewed one or the live one.
func (m *Model) shownPosition() chess.Position {
	if m.reviewing {
		return m.game.PositionAt(m.reviewPly)
	}
	return m.game.Position()
}

// shownLastMove is the move that led to the shown position.
func (m *Model) shownLastMove() (chess.Move, bool) {
	if !m.reviewing {
		return m.game.LastMove()
	}
	if m.reviewPly == 0 {
		return chess.NullMove, false
	}
	return m.game.History()[m.reviewPly-1].Move, true
}
