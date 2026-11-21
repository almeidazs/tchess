package tui

import (
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
)

// The theme picker lists the themes beside a board drawn in the highlighted
// one. Moving the cursor previews the theme everywhere, accent included;
// Enter keeps it and Esc goes back to the previous one.

// themeListW is the width of the theme names column.
const themeListW = 12

// previewGame is the position the preview shows: a few moves into the Two
// Knights, with White's knight on g5 selected so the preview has a last
// move, legal-move dots and a capture.
var previewGame = sync.OnceValue(func() *chess.Game {
	g := chess.NewGame()
	for _, san := range []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Nf6", "Ng5", "d5", "exd5", "Nxd5", "Bxd5", "Qxd5"} {
		pos := g.Position()
		mv, err := pos.ParseSAN(san)
		if err != nil {
			panic(err) // a fixed, legal line
		}
		_ = g.Play(mv)
	}
	return g
})

// openThemes shows the theme picker.
func (m *Model) openThemes() {
	m.themes = themePicker{cursor: m.themeIdx, before: m.themeIdx}
	m.screen = screenThemes
}

// themesKey handles a key on the theme picker.
func (m *Model) themesKey(key string) tea.Cmd {
	switch key {
	case "up", "k":
		m.pickTheme((m.themes.cursor + len(themes) - 1) % len(themes))
	case "down", "j", "tab":
		m.pickTheme((m.themes.cursor + 1) % len(themes))
	case "enter", " ", "space":
		return m.keepTheme()
	case "esc", "q":
		m.setTheme(m.themes.before)
		m.screen = screenMenu
	}
	return nil
}

// pickTheme highlights theme i and previews it.
func (m *Model) pickTheme(i int) {
	m.themes.cursor = i
	m.setTheme(i)
}

// keepTheme saves the highlighted theme and returns to the menu.
func (m *Model) keepTheme() tea.Cmd {
	m.setTheme(m.themes.cursor)
	m.screen = screenMenu
	m.cfg.Theme = m.theme().Name
	m.menu.notice = "Tema: " + m.theme().Name
	return m.saver.save(m.cfg)
}

// previewLayout picks the largest preview board that fits the terminal
// beside the list, falling back to one glyph per square.
func previewLayout(w, h int) layout {
	var l layout
	for _, s := range boardSizes {
		l = layout{cellW: s.cellW, cellH: s.cellH, pixel: s.cellH > 1}
		width := 2 + 2*menuPadX + themeListW + gapW + 8*l.cellW
		height := 2 + 2*menuPadY + 4 + 8*l.cellH // borders, padding, title and hints
		if width <= w && height <= h {
			break
		}
	}
	return l
}

// themePreview draws the preview position in theme i.
func (m *Model) themePreview(i int, l layout) []string {
	g := previewGame()
	p := &Model{game: g, orientation: chess.White, selected: chess.G5, targets: g.MovesFrom(chess.G5), now: m.now, mode: modePlay}
	p.setTheme(i)
	if l.pixel {
		return p.pixelBoard(l)
	}
	return p.glyphBoard(l)
}

// themesLines returns the picker's content; theme names start on line 2.
func (m *Model) themesLines() []string {
	l := previewLayout(m.width, m.height)
	board := m.themePreview(m.themes.cursor, l)
	gap := strings.Repeat(" ", gapW)
	width := themeListW + gapW + 8*l.cellW

	lines := []string{center(m.st.title.Render("Tema"), width), ""}
	for row, b := range board {
		name := ""
		if row < len(themes) {
			if row == m.themes.cursor {
				name = m.st.accent.Render("▸ " + themes[row].Name)
			} else {
				name = textStyle.Render("  " + themes[row].Name)
			}
		}
		lines = append(lines, pad(name, themeListW)+gap+b)
	}
	return append(lines, "", m.helpLine([][2]string{{"↑↓", "escolher"}, {"enter", "usar"}, {"esc", "voltar"}}))
}

// themesView renders the theme picker.
func (m *Model) themesView() string {
	return m.st.menuBox.Render(strings.Join(m.themesLines(), "\n"))
}

// themesClick previews the theme under the pointer, or keeps it when it is
// already highlighted.
func (m *Model) themesClick(x, y int) tea.Cmd {
	view := m.themesView()
	ox, oy := centerOffset(m.width, m.height, lipgloss.Width(view), lipgloss.Height(view))
	row := y - (oy + 1 + menuPadY + 2)
	col := x - (ox + 1 + menuPadX)
	if row < 0 || row >= len(themes) || col < 0 || col >= themeListW {
		return nil
	}
	if row == m.themes.cursor {
		return m.keepTheme()
	}
	m.pickTheme(row)
	return nil
}
