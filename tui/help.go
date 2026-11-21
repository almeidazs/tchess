package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// helpEntries lists the keys of the current screen, with the theme and
// sound state shown next to their keys.
func (m *Model) helpEntries() [][2]string {
	soundState := "ligado"
	if !m.soundOn {
		soundState = "desligado"
	}
	prefs := [][2]string{
		{"s", "tema: " + m.theme().Name},
		{"m", "som: " + soundState},
		{"?", "fechar a ajuda"},
		{"q", "sair"},
	}
	if m.screen == screenMenu {
		return append([][2]string{
			{"↑↓  jk", "escolher o ritmo"},
			{"enter", "jogar"},
		}, prefs...)
	}
	return append([][2]string{
		{"mouse", "clicar ou arrastar a peça"},
		{"←↑↓→  hjkl", "mover o cursor"},
		{"enter  espaço", "selecionar e jogar"},
		{":", "digitar o lance (Nf3, e2e4)"},
		{"esc", "cancelar"},
		{"d", "propor ou reivindicar empate"},
		{"r", "desistir"},
	}, prefs...)
}

// helpBox renders the help card. When it would be taller than maxH rows it
// drops its spacing and footer, so it fits even the smallest board.
func (m *Model) helpBox(maxH int) string {
	entries := m.helpEntries()
	keyW := 0
	for _, e := range entries {
		keyW = max(keyW, lipgloss.Width(e[0]))
	}
	keyStyle := m.st.key.Background(panelBg).Width(keyW + 3)
	descStyle := textStyle.Background(panelBg)
	rows := make([]string, 0, len(entries))
	width := 0
	for _, e := range entries {
		row := keyStyle.Render(e[0]) + descStyle.Render(e[1])
		rows = append(rows, row)
		width = max(width, lipgloss.Width(row))
	}
	for i := range rows {
		rows[i] = cardLine(rows[i], width, lipgloss.Left)
	}
	title := cardLine(m.st.title.Background(panelBg).Render("Ajuda"), width, lipgloss.Center)
	footer := cardLine(dimStyle.Background(panelBg).Render("? ou esc fecha"), width, lipgloss.Center)

	roomy := len(rows)+4+2+2 <= maxH // title, footer, blank lines, padding, border
	body := []string{title}
	blank := cardLine("", width, lipgloss.Left)
	if roomy {
		body = append(body, blank)
	}
	body = append(body, rows...)
	if roomy {
		body = append(body, blank, footer)
	}

	padY := 0
	if roomy {
		padY = 1
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.st.accentColor).
		Background(panelBg).
		Padding(padY, 2).
		Render(strings.Join(body, "\n"))
}
