package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// helpEntry is one key in the help card, with a full description and a
// short one for terminals too small for the full card.
type helpEntry struct{ key, long, short string }

// helpEntries lists the keys of the current screen, with the theme, sound
// and board-turning state shown next to their keys.
func (m *Model) helpEntries() []helpEntry {
	soundState := "ligado"
	if !m.soundOn {
		soundState = "desligado"
	}
	prefs := []helpEntry{
		{"s", "tema: " + m.theme().Name, "tema"},
		{"m", "som: " + soundState, "som"},
		{"f", "girar o tabuleiro: " + yesNo(m.flip), "girar"},
		{"?", "fechar a ajuda", "fechar"},
		{"q", "sair", "sair"},
	}
	if m.screen == screenMenu {
		return append([]helpEntry{
			{"↑↓  jk", "escolher", "escolher"},
			{"enter", "jogar ou abrir", "abrir"},
		}, prefs...)
	}
	return append([]helpEntry{
		{"mouse", "clicar ou arrastar a peça", "arrastar"},
		{"←↑↓→  hjkl", "mover o cursor", "cursor"},
		{"enter", "selecionar e jogar", "jogar"},
		{":", "digitar o lance (Nf3, e2e4)", "digitar"},
		{"esc", "cancelar", "cancelar"},
		{"u", "pedir para voltar o lance", "voltar"},
		{"[  ]", "rever lances (no fim, ←→)", "rever"},
		{"e", "exportar a partida em PGN", "exportar"},
		{"d", "propor ou reivindicar empate", "empate"},
		{"r", "desistir", "desistir"},
	}, prefs...)
}

// helpBox renders the help card for a screen maxH rows tall: with spacing
// when there is room, then without it, and in two columns of short labels
// when even that does not fit.
func (m *Model) helpBox(maxH int) string {
	entries := m.helpEntries()
	switch {
	case len(entries)+8 <= maxH: // title, footer, blank lines, padding, borders
		return m.helpCard(m.helpColumn(entries, false), true)
	case len(entries)+3 <= maxH:
		return m.helpCard(m.helpColumn(entries, false), false)
	}
	half := (len(entries) + 1) / 2
	left := m.helpColumn(entries[:half], true)
	right := m.helpColumn(entries[half:], true)
	gap := lipgloss.NewStyle().Background(panelBg).Render("  ")
	rows := make([]string, len(left))
	for i := range left {
		r := ""
		if i < len(right) {
			r = right[i]
		}
		rows[i] = left[i] + gap + r
	}
	return m.helpCard(rows, false)
}

// helpColumn renders entries as aligned "key  description" rows.
func (m *Model) helpColumn(entries []helpEntry, short bool) []string {
	keyW := 0
	for _, e := range entries {
		keyW = max(keyW, lipgloss.Width(e.key))
	}
	keyStyle := m.st.key.Background(panelBg).Width(keyW + 2)
	descStyle := textStyle.Background(panelBg)
	rows := make([]string, len(entries))
	width := 0
	for i, e := range entries {
		desc := e.long
		if short {
			desc = e.short
		}
		rows[i] = keyStyle.Render(e.key) + descStyle.Render(desc)
		width = max(width, lipgloss.Width(rows[i]))
	}
	for i := range rows {
		rows[i] = cardLine(rows[i], width, lipgloss.Left)
	}
	return rows
}

// helpCard frames the rows with the title and, when roomy, a footer and
// spacing.
func (m *Model) helpCard(rows []string, roomy bool) string {
	width := 0
	for _, r := range rows {
		width = max(width, lipgloss.Width(r))
	}
	title := cardLine(m.st.title.Background(panelBg).Render("Ajuda"), width, lipgloss.Center)
	body := []string{title}
	padY, padX := 0, 1
	if roomy {
		blank := cardLine("", width, lipgloss.Left)
		footer := cardLine(dimStyle.Background(panelBg).Render("? ou esc fecha"), width, lipgloss.Center)
		body = append(body, blank)
		body = append(body, rows...)
		body = append(body, blank, footer)
		padY, padX = 1, 2
	} else {
		body = append(body, rows...)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.st.accentColor).
		Background(panelBg).
		Padding(padY, padX).
		Render(strings.Join(body, "\n"))
}
