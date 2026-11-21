package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current screen, centered in the terminal.
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}
	var page string
	switch m.screen {
	case screenMenu:
		page = m.menuView()
	case screenThemes:
		page = m.themesView()
	case screenOpen:
		page = m.openView()
	case screenClocks:
		page = m.clocksView()
	}
	if page != "" {
		ox, oy := centerOffset(m.width, m.height, lipgloss.Width(page), lipgloss.Height(page))
		return place(page, ox, oy)
	}
	l := computeLayout(m.width, m.height)
	if !l.fits(m.width, m.height) {
		msg := lipgloss.JoinVertical(lipgloss.Center,
			m.st.accent.Render("Terminal too small"),
			dimStyle.Render(fmt.Sprintf("needs %dx%d, now %dx%d", l.totalW, l.totalH, m.width, m.height)))
		ox, oy := centerOffset(m.width, m.height, lipgloss.Width(msg), lipgloss.Height(msg))
		return place(msg, ox, oy)
	}
	return place(m.gameView(l), l.originX, l.originY)
}

// gameView composes header, board, panel and footer, plus the result
// card when the game is over.
func (m *Model) gameView(l layout) string {
	board := m.boardView(l)
	panel := strings.Split(m.panelView(l), "\n")
	gap := strings.Repeat(" ", gapW)

	lines := make([]string, 0, l.totalH)
	lines = append(lines, m.headerView(l), "")
	for i := range board {
		p := ""
		if i < len(panel) {
			p = panel[i]
		}
		lines = append(lines, board[i]+gap+p)
	}
	lines = append(lines, "", truncate(m.statusView(), l.totalW), truncate(m.helpView(), l.totalW))

	if m.mode == modeGameOver && m.anim == nil && !m.reviewing && !m.resultHidden {
		box := m.resultBox()
		bw, bh := lipgloss.Width(box), lipgloss.Height(box)
		x := max(rankLabelW+(8*l.cellW-bw)/2, 0)
		y := headerH + max((8*l.cellH-bh)/2, 0)
		lines = overlay(lines, strings.Split(box, "\n"), x, y)
	}
	if m.help {
		box := m.helpBox(l.totalH)
		bw, bh := lipgloss.Width(box), lipgloss.Height(box)
		lines = overlay(lines, strings.Split(box, "\n"), max((l.totalW-bw)/2, 0), max((l.totalH-bh)/2, 0))
	}
	return strings.Join(lines, "\n")
}
