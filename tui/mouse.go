package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
)

// hover returns the square under the pointer, if it is on the board.
func (d *dragState) hover(m *Model, l layout) (chess.Square, bool) {
	col, row, ok := l.cellAt(d.x, d.y)
	if !ok {
		return chess.NoSquare, false
	}
	return m.squareAt(col, row), true
}

// handleMouse dispatches mouse presses, drags and releases.
func (m *Model) handleMouse(msg tea.MouseMsg) tea.Cmd {
	switch {
	case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
		m.anim = nil // any input skips the animation
		if m.help {
			return nil
		}
		switch m.screen {
		case screenMenu:
			return m.menuClick(msg.X, msg.Y)
		case screenThemes:
			return m.themesClick(msg.X, msg.Y)
		case screenOpen:
			return m.openClick(msg.X, msg.Y)
		case screenClocks:
			return m.clocksClick(msg.X, msg.Y)
		}
		return m.mouseDown(msg.X, msg.Y)
	case msg.Action == tea.MouseActionMotion && m.drag != nil:
		m.mouseDrag(msg.X, msg.Y)
	case msg.Action == tea.MouseActionRelease && m.drag != nil:
		return m.mouseUp(msg.X, msg.Y)
	}
	return nil
}

// mouseDown handles a left press: promotion choice, click-to-move, or
// picking up one of our pieces.
func (m *Model) mouseDown(x, y int) tea.Cmd {
	l := computeLayout(m.width, m.height)
	if !l.fits(m.width, m.height) {
		return nil
	}
	col, row, onBoard := l.cellAt(x, y)
	if m.mode == modeInput { // a click leaves the prompt and acts normally
		m.closeInput()
	}
	if m.mode == modePlay || m.mode == modeGameOver {
		if ply, ok := m.historyAt(x, y); ok { // a move in the list: review it
			m.reviewAt(ply)
			return nil
		}
		if m.reviewing { // the board: back to the live position
			m.stopReview()
			return nil
		}
	}
	switch m.mode {
	case modePromotion:
		pc, pr0 := m.cellOf(m.promo[0].To())
		if onBoard && col == pc {
			if i := abs(row - pr0); i < len(m.promo) {
				return m.play(m.promo[i])
			}
		}
		m.mode = modePlay
		m.clearSelection()
	case modePlay:
		if !onBoard {
			m.clearSelection()
			return nil
		}
		m.showCursor = false
		m.cursorCol, m.cursorRow = col, row
		sq := m.squareAt(col, row)
		if m.isTarget(sq) { // click-click move
			return m.activate(sq)
		}
		pos := m.game.Position()
		if pc := pos.PieceAt(sq); pc != chess.NoPiece && pc.Color() == m.game.Turn() {
			m.drag = &dragState{from: sq, x: x, y: y, wasSelected: sq == m.selected}
			m.selected = sq
			m.targets = m.game.MovesFrom(sq)
			return nil
		}
		m.clearSelection()
	}
	return nil
}

// mouseDrag follows the pointer while a piece is held.
func (m *Model) mouseDrag(x, y int) {
	d := m.drag
	d.x, d.y = x, y
	if sq, ok := d.hover(m, computeLayout(m.width, m.height)); !ok || sq != d.from {
		d.moved = true
	}
}

// mouseUp drops the held piece: a move on a legal square, a click on its
// own square, or a snap back anywhere else.
func (m *Model) mouseUp(x, y int) tea.Cmd {
	d := m.drag
	m.drag = nil
	d.x, d.y = x, y
	sq, onBoard := d.hover(m, computeLayout(m.width, m.height))
	switch {
	case onBoard && sq == d.from:
		// A plain click on an already selected piece deselects it.
		if d.wasSelected && !d.moved {
			m.clearSelection()
		}
	case onBoard && m.isTarget(sq):
		m.cursorCol, m.cursorRow = m.cellOf(sq)
		m.dropped = true // the piece is already there: no slide animation
		return m.activate(sq)
	}
	// Dropped anywhere else: the piece snaps back and stays selected.
	return nil
}

// dragView summarizes the drag for the renderers.
func (m *Model) dragView(l layout) dragView {
	d := m.drag
	if d == nil || !d.moved || m.mode != modePlay {
		return dragView{hover: chess.NoSquare, from: chess.NoSquare}
	}
	pos := m.game.Position()
	dv := dragView{active: true, piece: pos.PieceAt(d.from), from: d.from, hover: chess.NoSquare,
		x: d.x - l.boardX, y: d.y - l.boardY}
	if sq, ok := d.hover(m, l); ok && m.isTarget(sq) {
		dv.hover = sq
	}
	return dv
}
