package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/chess/clock"
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

// menuLines returns the menu content and the index of the first item line.
func (m *Model) menuLines() (lines []string, firstItem int) {
	ms := &m.menu
	lines = append(lines,
		center(m.st.title.Render("♜ ♞ ♝ ♛ ♚ ♝ ♞ ♜"), menuInnerW),
		center(m.st.title.Render("X  A  D  R  E  Z"), menuInnerW),
		"",
		dimStyle.Render("Controle de tempo"),
		"",
	)
	firstItem = len(lines)
	for i, it := range menuItems {
		label := it.label
		right := ""
		if it.tc != "" && it.tc != "none" {
			right = it.tc
		}
		row := pad(label, menuInnerW-4-len(right)) + right
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
	return lines, firstItem
}

// menuHelp is the key hint line under the menu.
func (m *Model) menuHelp() string {
	if m.menu.editing {
		return m.helpLine([][2]string{{"enter", "jogar"}, {"esc", "voltar"}})
	}
	return m.helpLine([][2]string{{"enter", "jogar"}, {"?", "ajuda"}, {"q", "sair"}})
}

// menuView renders the time-control menu, or the help card over it.
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
	switch key {
	case "up", "k":
		ms.cursor = (ms.cursor + len(menuItems) - 1) % len(menuItems)
	case "down", "j", "tab":
		ms.cursor = (ms.cursor + 1) % len(menuItems)
	case "enter", " ", "space":
		return m.chooseMenuItem()
	case "q", "esc":
		return tea.Quit
	}
	return nil
}

// chooseMenuItem starts a game with the selected preset, or opens the
// custom time-control input.
func (m *Model) chooseMenuItem() tea.Cmd {
	it := menuItems[m.menu.cursor]
	m.menu.err = ""
	if it.tc == "" {
		m.menu.editing, m.menu.input = true, m.menu.custom
		return nil
	}
	tc, _ := clock.Parse(it.tc)
	return tea.Batch(m.rememberTimeControl(tc), m.startGame(tc))
}

// restore puts the cursor on the saved time control. One that is not a
// preset is offered again when the player opens the custom input.
func (ms *menuState) restore(saved string) {
	if saved == "" {
		return
	}
	want, err := clock.Parse(saved)
	if err != nil {
		return
	}
	for i, it := range menuItems {
		if it.tc == "" {
			ms.cursor, ms.custom = i, saved
			return
		}
		if tc, _ := clock.Parse(it.tc); tc == want {
			ms.cursor = i
			return
		}
	}
}

// menuClick selects and starts the preset under the pointer.
func (m *Model) menuClick(x, y int) tea.Cmd {
	if m.menu.editing || m.help {
		return nil
	}
	view := m.menuView()
	ox, oy := centerOffset(m.width, m.height, lipgloss.Width(view), lipgloss.Height(view))
	_, first := m.menuLines()
	row := y - (oy + 1 + menuPadY + first)
	col := x - (ox + 1 + menuPadX)
	if row < 0 || row >= len(menuItems) || col < 0 || col >= menuInnerW {
		return nil
	}
	m.menu.cursor = row
	return m.chooseMenuItem()
}
