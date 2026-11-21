package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// exportGame writes the game to a new PGN file in the games folder. A
// finished game is already there.
func (m *Model) exportGame() tea.Cmd {
	if m.mode == modeGameOver && m.savedPath != "" {
		m.notice = "The game is already saved in " + tildePath(m.savedPath)
		return nil
	}
	m.notice = "Saving the game…"
	return m.files.export(m.gamePGN(), m.exportName())
}

// saveFinished writes a game that just ended to the games folder.
func (m *Model) saveFinished() tea.Cmd {
	if m.files.gamesDir == "" {
		return nil
	}
	return m.files.export(m.gamePGN(), m.exportName())
}

// exportName names an exported file after the current moment.
func (m *Model) exportName() string {
	return m.now().Format("2006-01-02_15-04-05")
}

// onExported reports where the game was saved.
func (m *Model) onExported(msg exportedMsg) {
	if msg.err != nil {
		m.say("Could not save the game: " + msg.err.Error())
		return
	}
	if m.screen == screenGame && m.mode == modeGameOver {
		m.savedPath = msg.path
	}
	m.say("Game saved in " + tildePath(msg.path))
}
