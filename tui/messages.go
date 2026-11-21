package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg drives the clock display. gen discards ticks from previous games.
type tickMsg struct{ gen int }

const tickInterval = 100 * time.Millisecond

// tick schedules the next clock refresh.
func tick(gen int) tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{gen} })
}

type animMsg struct{ gen int }

// animTick schedules the next animation frame.
func animTick(gen int) tea.Cmd {
	return tea.Tick(frameRate, func(time.Time) tea.Msg { return animMsg{gen} })
}
