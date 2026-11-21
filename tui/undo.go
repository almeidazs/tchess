package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
)

// A take-back needs the opponent's consent. Outside review it undoes the
// last half-move, asked by the player who made it. While reviewing, it
// takes the game back to the position on the board, asked by the side to
// move there. The requester's clock goes back to what it showed when that
// move was made (without the increment it earned), the opponent's to what
// it showed before its time started running.

// requestUndo asks the opponent to accept going back to the position after
// target half-moves.
func (m *Model) requestUndo(target int) {
	m.notice = ""
	if m.game.Ply() == 0 || target >= m.game.Ply() {
		m.notice = "No move to take back"
		return
	}
	m.clearSelection()
	m.undoTarget = max(target, 0)
	m.mode = modeUndoOffer
}

// undoRequester is the side asking: the one to move after the take-back.
func (m *Model) undoRequester() chess.Color {
	pos := m.game.PositionAt(m.undoTarget)
	return pos.SideToMove()
}

// undoStatus asks the opponent to accept the take-back.
func (m *Model) undoStatus() string {
	who := m.undoRequester()
	n := m.game.Ply() - m.undoTarget
	what := "the last move"
	if n > 1 {
		what = fmt.Sprintf("%d moves", n)
	}
	return colorName(who) + " asks to take back " + what + ". " + colorName(who.Other()) + ", accept?  "
}

// undo takes the game back after the opponent accepted.
func (m *Model) undo() tea.Cmd {
	target := m.undoTarget
	taken := m.game.Ply() - target
	for m.game.Ply() > target {
		if err := m.game.Undo(); err != nil {
			m.mode = modePlay
			m.notice = err.Error()
			return nil
		}
	}
	m.moveTimes = m.moveTimes[:min(len(m.moveTimes), target)]
	if j := target - m.clockFrom; m.clock != nil && j >= 0 && j < len(m.moveClocks) {
		m.restoreClock(j, m.game.Turn())
	} else if m.clock != nil {
		// Nothing recorded to go back to: keep both times, but run the
		// clock of the side now to move.
		remaining := [2]time.Duration{m.clock.Remaining(chess.White), m.clock.Remaining(chess.Black)}
		m.clock.SetStartDelay(0)
		m.clock.Restore(remaining, m.game.Turn(), m.clock.Running())
	}
	if j := target - m.clockFrom; j >= 0 {
		m.moveClocks = m.moveClocks[:min(len(m.moveClocks), j)]
	} else {
		m.moveClocks, m.clockFrom = nil, target
	}
	m.reviewing = false
	m.turnStart = m.now()
	m.anim = nil
	m.mode = modePlay
	m.orientation = m.facing()
	m.clearSelection()
	m.notice = "Move taken back"
	if taken > 1 {
		m.notice = fmt.Sprintf("%d moves taken back", taken)
	}
	return m.files.saveResume(m.gamePGN())
}

// restoreClock sets both clocks as they were when the move moveClocks[j]
// was made by mover.
func (m *Model) restoreClock(j int, mover chess.Color) {
	var remaining [2]time.Duration
	mc := m.moveClocks[j]
	remaining[mover] = mc.before
	remaining[mover.Other()] = m.tc.Base
	if j > 0 {
		remaining[mover.Other()] = m.moveClocks[j-1].after
	}
	m.clock.SetStartDelay(0)
	m.clock.Restore(remaining, mover, mc.running)
}
