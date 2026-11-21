package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The opponent screen asks who to play: another person on this terminal or
// the computer. For the computer it then asks the level and the human's
// color. Without an engine, it explains how to get one.

// pickerRow is one choice on the opponent screen.
type pickerRow struct {
	label, right string
	disabled     bool
}

// colorChoices are the sides the human can take against the computer.
var colorChoices = []struct{ key, label string }{
	{"white", "White"}, {"black", "Black"}, {"random", "Random"},
}

// noEngineHelp explains how to get an engine.
var noEngineHelp = []string{
	"No UCI engine found. Install Stockfish:",
	"  apt install stockfish   (Debian, Ubuntu)",
	"  dnf install stockfish   (Fedora)",
	"  brew install stockfish  (macOS)",
	"  or get it at stockfishchess.org",
	"Or point to an engine with --engine, or",
	`engine = "/path" in ~/.tchess/config.toml.`,
}

// openOpponent shows the opponent screen on the current choice.
func (m *Model) openOpponent() {
	m.opp = opponentPicker{choice: m.choice}
	if m.choice.computer {
		m.opp.cursor = 1
	}
	m.screen = screenOpponent
}

// opponentRows lists the choices of the current step.
func (m *Model) opponentRows() []pickerRow {
	switch m.opp.step {
	case 1:
		rows := make([]pickerRow, 0, len(levels)+1)
		for _, l := range levels {
			right := "full strength"
			if l.elo > 0 {
				right = fmt.Sprintf("~%d", l.elo)
			}
			rows = append(rows, pickerRow{label: l.name, right: right})
		}
		custom := ""
		if !isPreset(m.opp.choice.elo) {
			custom = strconv.Itoa(m.opp.choice.elo)
		}
		return append(rows, pickerRow{label: "Custom Elo…", right: custom})
	case 2:
		rows := make([]pickerRow, len(colorChoices))
		for i, c := range colorChoices {
			rows[i] = pickerRow{label: c.label}
		}
		return rows
	}
	computer := pickerRow{label: "Computer", right: m.computerName()}
	if !m.engine.available() {
		computer = pickerRow{label: "Computer", right: "no engine found", disabled: true}
	}
	return []pickerRow{{label: "Human", right: "two players"}, computer}
}

// isPreset reports whether elo is one of the named levels.
func isPreset(elo int) bool {
	for _, l := range levels {
		if l.elo == elo {
			return true
		}
	}
	return false
}

// opponentKey handles a key on the opponent screen.
func (m *Model) opponentKey(key string) tea.Cmd {
	p := &m.opp
	if p.editing {
		switch key {
		case "esc":
			p.editing, p.input, p.err = false, "", ""
		case "enter":
			elo, ok := ParseLevel(p.input)
			if !ok || elo == 0 {
				p.err = fmt.Sprintf("Elo between %d and %d", minElo, maxElo)
				return nil
			}
			p.choice.elo, p.editing, p.err = elo, false, ""
			p.step, p.cursor = 2, colorIndex(p.choice.color)
		case "backspace":
			if p.input != "" {
				p.input = p.input[:len(p.input)-1]
			}
		default:
			if len(key) == 1 && key[0] >= '0' && key[0] <= '9' && len(p.input) < 4 {
				p.input += key
				p.err = ""
			}
		}
		return nil
	}
	n := len(m.opponentRows())
	switch key {
	case "up", "k":
		p.cursor = (p.cursor + n - 1) % n
	case "down", "j", "tab":
		p.cursor = (p.cursor + 1) % n
	case "enter", " ", "space":
		return m.chooseOpponentRow()
	case "esc", "q":
		p.err = ""
		if p.step == 0 {
			m.screen = screenMenu
			return nil
		}
		p.step--
		p.cursor = 0
		if p.step == 0 {
			p.cursor = 1 // back to "Computer"
		}
	}
	return nil
}

// colorIndex is the row of a color choice.
func colorIndex(color string) int {
	for i, c := range colorChoices {
		if c.key == color {
			return i
		}
	}
	return 0
}

// chooseOpponentRow acts on the highlighted row.
func (m *Model) chooseOpponentRow() tea.Cmd {
	p := &m.opp
	switch p.step {
	case 0:
		if p.cursor == 0 {
			p.choice.computer = false
			return m.keepOpponent()
		}
		if !m.engine.available() {
			p.err = "Install an engine first"
			return nil
		}
		p.choice.computer = true
		p.step, p.cursor = 1, 2 // Medium by default
		for i, l := range levels {
			if l.elo == p.choice.elo {
				p.cursor = i
			}
		}
		if !isPreset(p.choice.elo) {
			p.cursor = len(levels)
		}
	case 1:
		if p.cursor == len(levels) {
			p.editing, p.input, p.err = true, "", ""
			if !isPreset(p.choice.elo) {
				p.input = strconv.Itoa(p.choice.elo)
			}
			return nil
		}
		p.choice.elo = levels[p.cursor].elo
		p.step, p.cursor = 2, colorIndex(p.choice.color)
	case 2:
		p.choice.color = colorChoices[p.cursor].key
		return m.keepOpponent()
	}
	return nil
}

// keepOpponent saves the choice and returns to the menu.
func (m *Model) keepOpponent() tea.Cmd {
	m.choice = m.opp.choice
	m.cfg.Opponent = "human"
	if m.choice.computer {
		m.cfg.Opponent = "computer"
	}
	m.cfg.ComputerElo, m.cfg.HumanColor = m.choice.elo, m.choice.color
	m.screen = screenMenu
	m.menu.notice = "Opponent: " + m.opponentSummary()
	return m.saver.save(m.cfg)
}

// opponentSummary describes the menu choice: "Human" or "Computer (Medium)".
func (m *Model) opponentSummary() string {
	if !m.choice.computer {
		return "Human"
	}
	if !m.engine.available() {
		return "Computer (no engine)"
	}
	return "Computer (" + levelName(m.choice.elo) + ")"
}

// opponentWidth is the content width: wider when the install help shows.
func (m *Model) opponentWidth() int {
	width := menuInnerW
	if !m.engine.available() && m.opp.step == 0 {
		for _, l := range noEngineHelp {
			width = max(width, lipgloss.Width(l))
		}
	}
	return width
}

// opponentLines returns the screen's content; rows start on line 3.
func (m *Model) opponentLines() []string {
	w := m.opponentWidth()
	p := &m.opp
	titles := []string{"Opponent", "Computer level", "Your color"}
	sub := "Engine: " + m.computerName()
	if !m.engine.available() {
		sub = "No engine found"
	}
	lines := []string{center(m.st.title.Render(titles[p.step]), w), center(dimStyle.Render(truncate(sub, w)), w), ""}
	for i, r := range m.opponentRows() {
		row := menuRow(r.label, r.right)
		switch {
		case i == p.cursor && r.disabled:
			lines = append(lines, dimStyle.Render("▸ "+row+"  "))
		case i == p.cursor:
			lines = append(lines, m.st.accent.Render("▸ "+row+"  "))
		case r.disabled:
			lines = append(lines, dimStyle.Render("  "+row+"  "))
		default:
			lines = append(lines, textStyle.Render("  "+row+"  "))
		}
	}
	lines = append(lines, "")
	switch {
	case p.editing:
		lines = append(lines, textStyle.Render("Elo: ")+m.st.accent.Render(p.input+"▌"))
	case p.err != "":
		lines = append(lines, dangerStyle.Render(truncate(p.err, menuInnerW)))
	default:
		lines = append(lines, "")
	}
	if p.step == 0 && !m.engine.available() {
		lines = append(lines, "")
		for _, l := range noEngineHelp {
			lines = append(lines, dimStyle.Render(l))
		}
	}
	return append(lines, "", m.helpLine([][2]string{{"enter", "choose"}, {"esc", "back"}}))
}

// opponentView renders the opponent screen.
func (m *Model) opponentView() string {
	return m.st.menuBox.Width(m.opponentWidth() + 2*menuPadX).Render(strings.Join(m.opponentLines(), "\n"))
}

// opponentClick chooses the row under the pointer.
func (m *Model) opponentClick(x, y int) tea.Cmd {
	if m.opp.editing {
		return nil
	}
	view := m.opponentView()
	ox, oy := centerOffset(m.width, m.height, lipgloss.Width(view), lipgloss.Height(view))
	row := y - (oy + 1 + menuPadY + 3)
	col := x - (ox + 1 + menuPadX)
	if row < 0 || row >= len(m.opponentRows()) || col < 0 || col >= menuInnerW {
		return nil
	}
	m.opp.cursor = row
	return m.chooseOpponentRow()
}
