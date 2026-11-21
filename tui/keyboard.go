package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
)

// gameKey handles a key press on the game screen, according to the mode.
func (m *Model) gameKey(key string) tea.Cmd {
	switch m.mode {
	case modeGameOver, modePlay:
		if m.mode == modeGameOver { // no cursor here: the arrows step through the moves
			switch key {
			case "left", "h":
				key = "["
			case "right", "l":
				key = "]"
			}
		}
		if cmd, ok := m.boardKey(key); ok {
			return cmd
		}
		if m.reviewing && key != "q" {
			return nil // a past position cannot be played on
		}
		if m.mode == modeGameOver {
			return m.gameOverKey(key)
		}
		return m.playKey(key)

	case modeInput:
		return m.inputKey(key)

	case modePromotion:
		switch key {
		case "up", "k":
			m.promoIdx = (m.promoIdx + len(m.promo) - 1) % len(m.promo)
		case "down", "j":
			m.promoIdx = (m.promoIdx + 1) % len(m.promo)
		case "enter", " ", "space":
			return m.play(m.promo[m.promoIdx])
		case "q", "r", "b", "n":
			return m.play(m.promo[map[string]int{"q": 0, "r": 1, "b": 2, "n": 3}[key]])
		case "esc":
			m.mode = modePlay
			m.clearSelection()
		}

	case modeConfirmResign:
		switch key {
		case "y", "s":
			return m.conclude(m.game.Resign(m.game.Turn()))
		case "n", "esc":
			m.mode = modePlay
		}

	case modeDrawOffer:
		switch key {
		case "y", "s":
			return m.conclude(m.game.AgreeDraw())
		case "n", "esc":
			m.mode = modePlay
			m.notice = "Empate recusado"
		}

	case modeUndoOffer:
		switch key {
		case "y", "s":
			return m.undo()
		case "n", "esc":
			m.mode = modePlay
			m.notice = "Pedido para voltar o lance recusado"
		}

	case modeConfirmQuit:
		switch key {
		case "y", "s":
			return tea.Quit
		case "n", "esc":
			m.mode = modePlay
		}
	}
	return nil
}

// boardKey handles the keys shared by live play, review and the game-over
// screen: reviewing moves, preferences and export.
func (m *Model) boardKey(key string) (tea.Cmd, bool) {
	if m.reviewKey(key) {
		return nil, true
	}
	if cmd, ok := m.prefKey(key); ok {
		return cmd, true
	}
	if key == "e" {
		return m.exportGame(), true
	}
	return nil, false
}

// gameOverKey handles the result screen.
func (m *Model) gameOverKey(key string) tea.Cmd {
	switch key {
	case "n":
		m.toMenu()
	case "esc":
		m.resultHidden = !m.resultHidden
	case "q":
		return tea.Quit
	}
	return nil
}

// playKey handles a key while waiting for a move.
func (m *Model) playKey(key string) tea.Cmd {
	switch key {
	case ":":
		m.openInput()
	case "up", "k":
		m.moveCursor(0, -1)
	case "down", "j":
		m.moveCursor(0, 1)
	case "left", "h":
		m.moveCursor(-1, 0)
	case "right", "l":
		m.moveCursor(1, 0)
	case "enter", " ", "space":
		if !m.showCursor {
			m.revealCursor()
			return nil
		}
		return m.activate(m.squareAt(m.cursorCol, m.cursorRow))
	case "esc":
		m.clearSelection()
	case "u":
		m.requestUndo()
	case "d":
		m.notice = ""
		if m.game.ClaimableDraw() != chess.NoTermination {
			return m.conclude(m.game.ClaimDraw())
		}
		m.mode = modeDrawOffer
	case "r":
		m.notice = ""
		m.mode = modeConfirmResign
	case "q":
		m.mode = modeConfirmQuit
	}
	return nil
}
