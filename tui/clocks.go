package tui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
)

// Before playing on an unfinished imported game, the player picks how to
// time the rest of it: the record's own clocks when it has them, one of
// the menu's time controls with both clocks full, or no clock.

// clockNoTime is the time control of a game without a clock.
var clockNoTime = clock.TimeControl{}

// askClocks shows the time-control question for pg.
func (m *Model) askClocks(pg *chess.PGNGame, source string, back screen) {
	var opts []clockOption
	if _, _, remaining, ok := recordClocks(pg); ok {
		opts = append(opts, clockOption{
			label: "Relógios do PGN",
			right: formatClock(remaining[chess.White]) + " · " + formatClock(remaining[chess.Black]),
			src:   clocksFromRecord,
		})
	}
	for _, it := range menuItems {
		switch it.tc {
		case "": // custom input: not offered here
		case "none":
			opts = append(opts, clockOption{label: it.label, src: clocksOff})
		default:
			tc, _ := clock.Parse(it.tc)
			opts = append(opts, clockOption{label: it.label, right: it.tc, src: clocksFresh, tc: tc})
		}
	}
	m.clocks = clockChoice{record: pg, source: source, options: opts, back: back}
	m.screen = screenClocks
}

// clocksKey handles a key on the time-control question.
func (m *Model) clocksKey(key string) tea.Cmd {
	c := &m.clocks
	n := len(c.options)
	switch key {
	case "up", "k":
		c.cursor = (c.cursor + n - 1) % n
	case "down", "j", "tab":
		c.cursor = (c.cursor + 1) % n
	case "enter", " ", "space":
		return m.chooseClock()
	case "esc", "q":
		m.screen = c.back
	}
	return nil
}

// chooseClock starts playing with the highlighted option.
func (m *Model) chooseClock() tea.Cmd {
	c := &m.clocks
	o := c.options[c.cursor]
	cmd := m.loadRecord(c.record, o.src, o.tc)
	m.savedPath = c.source
	return cmd
}

// clocksLines returns the question's content; options start on line 3.
func (m *Model) clocksLines() []string {
	c := &m.clocks
	pos := c.record.Game.Position()
	lines := []string{
		center(m.st.title.Render("Continuar a partida"), menuInnerW),
		center(dimStyle.Render("lance "+strconv.Itoa(pos.FullmoveNumber())+", vez das "+strings.ToLower(colorName(pos.SideToMove()))), menuInnerW),
		"",
	}
	for i, o := range c.options {
		row := menuRow(o.label, o.right)
		if i == c.cursor {
			lines = append(lines, m.st.accent.Render("▸ "+row+"  "))
		} else {
			lines = append(lines, textStyle.Render("  "+row+"  "))
		}
	}
	return append(lines, "", m.helpLine([][2]string{{"enter", "jogar"}, {"esc", "voltar"}}))
}

// clocksView renders the time-control question.
func (m *Model) clocksView() string {
	return m.st.menuBox.Width(menuInnerW + 2*menuPadX).Render(strings.Join(m.clocksLines(), "\n"))
}

// clocksClick starts with the option under the pointer.
func (m *Model) clocksClick(x, y int) tea.Cmd {
	view := m.clocksView()
	ox, oy := centerOffset(m.width, m.height, lipgloss.Width(view), lipgloss.Height(view))
	row := y - (oy + 1 + menuPadY + 3)
	col := x - (ox + 1 + menuPadX)
	if row < 0 || row >= len(m.clocks.options) || col < 0 || col >= menuInnerW {
		return nil
	}
	m.clocks.cursor = row
	return m.chooseClock()
}
