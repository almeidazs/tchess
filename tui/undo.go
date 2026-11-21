package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
)

// A take-back is asked by the player who just moved and needs the
// opponent's consent. It undoes one half-move. The requester's clock goes
// back to what it showed at the moment of the move (without the increment
// the move earned) and the opponent's to what it showed before its time
// started running.

// requestUndo asks the opponent to accept taking back the last move.
func (m *Model) requestUndo() {
	m.notice = ""
	if m.game.Ply() == 0 {
		m.notice = "Nenhum lance para voltar"
		return
	}
	m.clearSelection()
	m.mode = modeUndoOffer
}

// undo takes back the last move after the opponent accepted.
func (m *Model) undo() tea.Cmd {
	ply := m.game.Ply()
	mover := m.game.Turn().Other()
	if err := m.game.Undo(); err != nil {
		m.mode = modePlay
		m.notice = err.Error()
		return nil
	}
	m.moveTimes = m.moveTimes[:min(len(m.moveTimes), ply-1)]
	if m.clock != nil && ply-1 < len(m.moveClocks) {
		m.restoreClock(ply-1, mover)
	}
	m.moveClocks = m.moveClocks[:min(len(m.moveClocks), ply-1)]
	m.turnStart = m.now()
	m.anim = nil
	m.mode = modePlay
	m.orientation = m.facing()
	m.clearSelection()
	m.notice = "Lance voltado"
	return m.files.saveResume(m.gamePGN())
}

// restoreClock sets both clocks as they were when the move at index i
// (0-based ply) was made by mover.
func (m *Model) restoreClock(i int, mover chess.Color) {
	var remaining [2]time.Duration
	pr := m.moveClocks[i]
	remaining[mover] = pr.before
	remaining[mover.Other()] = m.tc.Base
	if i > 0 {
		remaining[mover.Other()] = m.moveClocks[i-1].after
	}
	m.clock.SetStartDelay(0)
	m.clock.Restore(remaining, mover, pr.running)
}
