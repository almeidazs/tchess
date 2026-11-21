package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
)

// activate handles a click or Enter on sq: move the selected piece there,
// select one of our pieces, or clear the selection.
func (m *Model) activate(sq chess.Square) tea.Cmd {
	m.notice = ""
	if m.selected != chess.NoSquare {
		var cands []chess.Move
		for _, mv := range m.targets {
			if mv.To() == sq {
				cands = append(cands, mv)
			}
		}
		switch {
		case len(cands) == 1:
			return m.play(cands[0])
		case len(cands) > 1: // promotion: let the player pick the piece
			m.promo = orderPromotions(cands)
			m.promoIdx = 0
			m.mode = modePromotion
			return nil
		}
	}
	pos := m.game.Position()
	if pc := pos.PieceAt(sq); pc != chess.NoPiece && pc.Color() == m.game.Turn() && sq != m.selected {
		m.selected = sq
		m.targets = m.game.MovesFrom(sq)
		return nil
	}
	m.clearSelection()
	return nil
}

// isTarget reports whether the selected piece can move to sq.
func (m *Model) isTarget(sq chess.Square) bool {
	for _, mv := range m.targets {
		if mv.To() == sq {
			return true
		}
	}
	return false
}

// clearSelection drops the selection, drag and pending promotion.
func (m *Model) clearSelection() {
	m.drag = nil
	m.dropped = false
	m.selected = chess.NoSquare
	m.targets = nil
	m.promo = nil
}
