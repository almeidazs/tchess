package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/chess/chess"
)

// gameKey handles a key press on the game screen, according to the mode.
func (m *Model) gameKey(key string) tea.Cmd {
	switch m.mode {
	case modeGameOver:
		if cmd, ok := m.prefKey(key); ok {
			return cmd
		}
		switch key {
		case "n":
			m.toMenu()
		case "q", "esc":
			return tea.Quit
		}

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
			m.conclude(m.game.Resign(m.game.Turn()))
		case "n", "esc":
			m.mode = modePlay
		}

	case modeDrawOffer:
		switch key {
		case "y", "s":
			m.conclude(m.game.AgreeDraw())
		case "n", "esc":
			m.mode = modePlay
			m.notice = "Empate recusado"
		}

	case modeConfirmQuit:
		switch key {
		case "y", "s":
			return tea.Quit
		case "n", "esc":
			m.mode = modePlay
		}

	case modeInput:
		return m.inputKey(key)

	case modePlay:
		if cmd, ok := m.prefKey(key); ok {
			return cmd
		}
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
		case "d":
			m.notice = ""
			if m.game.ClaimableDraw() != chess.NoTermination {
				m.conclude(m.game.ClaimDraw())
			} else {
				m.mode = modeDrawOffer
			}
		case "r":
			m.notice = ""
			m.mode = modeConfirmResign
		case "q":
			m.mode = modeConfirmQuit
		}
	}
	return nil
}
