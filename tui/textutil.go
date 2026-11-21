package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// --- small rendering helpers ---------------------------------------------

// helpLine renders "key action · key action" hints.
func (m *Model) helpLine(pairs [][2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = m.st.key.Render(p[0]) + " " + dimStyle.Render(p[1])
	}
	return strings.Join(parts, dimStyle.Render("  ·  "))
}

// pad right-pads s (which may contain ANSI codes) to width w.
func pad(s string, w int) string {
	return s + strings.Repeat(" ", max(w-lipgloss.Width(s), 0))
}

// center pads s (which may contain ANSI codes) to width w, centered.
func center(s string, w int) string {
	gap := max(w-lipgloss.Width(s), 0)
	return strings.Repeat(" ", gap/2) + s + strings.Repeat(" ", gap-gap/2)
}

// cardLine fits s to width w on the panel background, so text drawn over
// the board shows no gaps: every padding cell gets the card's color.
func cardLine(s string, w int, pos lipgloss.Position) string {
	return lipgloss.PlaceHorizontal(w, pos, s, lipgloss.WithWhitespaceBackground(panelBg))
}

// truncate cuts s to w cells, ending with an ellipsis.
func truncate(s string, w int) string { return ansi.Truncate(s, w, "…") }

// place shifts a block right by x columns and down by y rows.
func place(s string, x, y int) string {
	indent := strings.Repeat(" ", x)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = indent + lines[i]
	}
	return strings.Repeat("\n", y) + strings.Join(lines, "\n")
}

// overlay draws fg on top of bg with its top-left corner at (x, y).
func overlay(bg, fg []string, x, y int) []string {
	out := append([]string(nil), bg...)
	for i, fl := range fg {
		row := y + i
		if row < 0 || row >= len(out) {
			continue
		}
		base := pad(out[row], x+lipgloss.Width(fl))
		left := ansi.Truncate(base, x, "")
		right := ansi.TruncateLeft(base, x+lipgloss.Width(fl), "")
		out[row] = left + "\x1b[0m" + fl + "\x1b[0m" + right
	}
	return out
}
