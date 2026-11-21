package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/clock"
)

var menuItems = []menuItem{
	{"Bullet", "1+0"},
	{"Blitz", "3+2"},
	{"Blitz", "5+3"},
	{"Rápida", "10+0"},
	{"Rápida", "15+10"},
	{"Sem relógio", "none"},
	{"Personalizado…", ""},
}

const (
	menuInnerW = 34
	menuPadX   = 3
	menuPadY   = 1
)

// menuEntries lists the selectable lines: the saved game, if any, the
// time controls and the options.
func (m *Model) menuEntries() []menuEntry {
	var out []menuEntry
	if m.resume != nil {
		out = append(out, menuEntry{kind: entryResume, label: "Continuar", right: m.resume.label})
	}
	for _, it := range menuItems {
		e := menuEntry{kind: entryTime, label: it.label, tc: it.tc}
		switch it.tc {
		case "":
			e.kind = entryCustom
		case "none":
		default:
			e.right = it.tc
		}
		out = append(out, e)
	}
	return append(out,
		menuEntry{kind: entryTheme, label: "Tema", right: m.theme().Name + " ›"},
		menuEntry{kind: entryFlip, label: "Girar tabuleiro", right: yesNo(m.flip)},
		menuEntry{kind: entryOpen, label: "Abrir PGN", right: "›"},
	)
}

// menuLines returns the menu content and, for each entry, its line.
func (m *Model) menuLines() (lines []string, entryLine []int) {
	ms := &m.menu
	entries := m.menuEntries()
	lines = append(lines,
		center(m.st.title.Render("♜ ♞ ♝ ♛ ♚ ♝ ♞ ♜"), menuInnerW),
		center(m.st.title.Render("X  A  D  R  E  Z"), menuInnerW),
		"",
	)
	for i, e := range entries {
		switch {
		case e.kind == entryResume:
			lines = append(lines, dimStyle.Render("Partida salva"))
		case e.kind == entryTime && (i == 0 || entries[i-1].kind == entryResume):
			if i > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, dimStyle.Render("Controle de tempo"))
		case e.kind == entryTheme:
			lines = append(lines, "", dimStyle.Render("Opções"))
		}
		entryLine = append(entryLine, len(lines))
		row := menuRow(e.label, e.right)
		if i == ms.cursor {
			lines = append(lines, m.st.accent.Render("▸ "+row+"  "))
		} else {
			lines = append(lines, textStyle.Render("  "+row+"  "))
		}
	}
	lines = append(lines, "")
	switch {
	case ms.editing:
		lines = append(lines, textStyle.Render("min+inc: ")+m.st.accent.Render(ms.input+"▌"))
	case ms.err != "":
		lines = append(lines, dangerStyle.Render(truncate(ms.err, menuInnerW)))
	case ms.notice != "":
		lines = append(lines, dimStyle.Render(truncate(ms.notice, menuInnerW)))
	default:
		lines = append(lines, "")
	}
	lines = append(lines, "", m.menuHelp())
	return lines, entryLine
}

// menuRow lays out a label and its right-aligned detail in the width of a
// menu line (minus the cursor and margin), keeping at least two spaces
// between them; a detail too long for that is cut with an ellipsis.
func menuRow(label, right string) string {
	const width, minGap = menuInnerW - 4, 2
	right = truncate(right, max(width-lipgloss.Width(label)-minGap, 0))
	return pad(label, width-lipgloss.Width(right)) + right
}

// menuHelp is the key hint line under the menu.
func (m *Model) menuHelp() string {
	if m.menu.editing {
		return m.helpLine([][2]string{{"enter", "jogar"}, {"esc", "voltar"}})
	}
	return m.helpLine([][2]string{{"enter", "abrir"}, {"?", "ajuda"}, {"q", "sair"}})
}

// menuView renders the menu, or the help card over it.
func (m *Model) menuView() string {
	if m.help {
		return m.helpBox(m.height)
	}
	lines, _ := m.menuLines()
	return m.st.menuBox.Width(menuInnerW + 2*menuPadX).Render(strings.Join(lines, "\n"))
}

// menuKey handles a key press on the menu.
func (m *Model) menuKey(msg tea.KeyMsg) tea.Cmd {
	ms := &m.menu
	key := msg.String()
	if ms.editing {
		switch key {
		case "esc":
			ms.editing, ms.input, ms.err = false, "", ""
		case "enter":
			tc, err := clock.Parse(ms.input)
			if err != nil || ms.input == "" {
				ms.editing = false
				ms.err = "Formato: minutos+segundos, ex. 7+2"
				if err != nil {
					ms.err = err.Error()
				}
				return nil
			}
			ms.editing, ms.custom = false, ms.input
			return tea.Batch(m.rememberTimeControl(tc), m.startGame(tc))
		case "backspace":
			if ms.input != "" {
				ms.input = ms.input[:len(ms.input)-1]
			}
		default:
			if len(key) == 1 && strings.IndexByte("0123456789+.", key[0]) >= 0 && len(ms.input) < 9 {
				ms.input += key
			}
		}
		return nil
	}
	ms.notice = ""
	if cmd, ok := m.prefKey(key); ok {
		return cmd
	}
	n := len(m.menuEntries())
	switch key {
	case "up", "k":
		ms.cursor = (ms.cursor + n - 1) % n
	case "down", "j", "tab":
		ms.cursor = (ms.cursor + 1) % n
	case "enter", " ", "space":
		return m.chooseEntry()
	case "q", "esc":
		return tea.Quit
	}
	return nil
}

// chooseEntry acts on the entry under the cursor.
func (m *Model) chooseEntry() tea.Cmd {
	entries := m.menuEntries()
	m.menu.cursor = min(m.menu.cursor, len(entries)-1)
	e := entries[m.menu.cursor]
	m.menu.err = ""
	switch e.kind {
	case entryResume:
		return m.loadRecord(m.resume.record, true)
	case entryCustom:
		m.menu.editing, m.menu.input = true, m.menu.custom
	case entryTime:
		tc, _ := clock.Parse(e.tc)
		return tea.Batch(m.rememberTimeControl(tc), m.startGame(tc))
	case entryTheme:
		m.openThemes()
	case entryFlip:
		return m.toggleFlip()
	case entryOpen:
		m.openPicker()
	}
	return nil
}

// restore puts the cursor on the saved game when there is one, otherwise
// on the saved time control. A time control that is not a preset is
// offered again when the player opens the custom input.
func (ms *menuState) restore(entries []menuEntry, saved string) {
	ms.cursor = 0 // the saved game, or the first preset
	want, err := clock.Parse(saved)
	if saved == "" || err != nil {
		return
	}
	match := -1
	for i, e := range entries {
		if tc, _ := clock.Parse(e.tc); e.kind == entryTime && tc == want {
			match = i
			break
		}
	}
	if match < 0 {
		ms.custom = saved
		for i, e := range entries {
			if e.kind == entryCustom {
				match = i
			}
		}
	}
	if entries[0].kind != entryResume {
		ms.cursor = match
	}
}

// menuClick acts on the entry under the pointer.
func (m *Model) menuClick(x, y int) tea.Cmd {
	if m.menu.editing {
		return nil
	}
	view := m.menuView()
	ox, oy := centerOffset(m.width, m.height, lipgloss.Width(view), lipgloss.Height(view))
	lines, entryLine := m.menuLines()
	row := y - (oy + 1 + menuPadY)
	col := x - (ox + 1 + menuPadX)
	if row < 0 || row >= len(lines) || col < 0 || col >= menuInnerW {
		return nil
	}
	for i, l := range entryLine {
		if l == row {
			m.menu.cursor = i
			return m.chooseEntry()
		}
	}
	return nil
}
