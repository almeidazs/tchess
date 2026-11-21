package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
)

// panelView renders the clocks (opponent on top, player to move at the
// bottom, matching the board) and the move list between them.
func (m *Model) panelView(l layout) string {
	orient := m.viewOrientation()
	top := m.clockView(orient.Other(), l.compact())
	bottom := m.clockView(orient, l.compact())
	histH := l.boardH() - lipgloss.Height(top) - lipgloss.Height(bottom)
	return lipgloss.JoinVertical(lipgloss.Left, top, m.historyView(histH), bottom)
}

// clockView shows a player's name and clock, and below them the pieces
// they have captured plus their material advantage.
func (m *Model) clockView(c chess.Color, compact bool) string {
	active := !m.game.IsOver() && m.game.Turn() == c
	name := kingGlyph(c) + " " + colorName(c)
	timeStr, low := "", false
	if m.clock != nil {
		rem := m.clock.Remaining(c)
		timeStr = formatClock(rem)
		low = rem < lowTime
	} else if active {
		timeStr = "◀ vez"
	}

	nameStyle, timeStyle := dimStyle, dimStyle
	if active {
		nameStyle, timeStyle = m.st.accent, m.st.accent
	}
	if low && m.clock != nil {
		timeStyle = dangerStyle
	}
	inner := panelW - 4
	content := nameStyle.Render(name) +
		strings.Repeat(" ", max(inner-lipgloss.Width(name)-lipgloss.Width(timeStr), 1)) +
		timeStyle.Render(timeStr) + "\n" +
		truncate(m.capturedLine(c), inner)

	if compact {
		st := lipgloss.NewStyle().Width(panelW).Padding(0, 2)
		if active {
			st = st.Background(panelBg)
		}
		return st.Render(content)
	}
	if active {
		return m.st.activeBox.Width(panelW - 2).Render(content)
	}
	return boxStyle.Width(panelW - 2).Render(content)
}

// capturedLine lists the pieces c has captured, grouped by kind from pawns
// to queens, followed by c's material advantage when ahead.
func (m *Model) capturedLine(c chess.Color) string {
	var count [6]int
	for _, pc := range m.game.Captured(c) {
		count[pc.Type()]++
	}
	var sb strings.Builder
	for t := chess.Pawn; t < chess.King; t++ {
		for range count[t] {
			sb.WriteString(glyphs[c.Other()][t])
		}
	}
	line := capturedStyle.Render(sb.String())
	pos := m.game.Position()
	if adv := pos.Material(c) - pos.Material(c.Other()); adv > 0 {
		if sb.Len() > 0 {
			line += " "
		}
		line += m.st.accent.Render(fmt.Sprintf("+%d", adv))
	}
	return line
}

// historyView lists one half-move per line: move number, SAN, a bar
// proportional to the thinking time and the time itself.
func (m *Model) historyView(height int) string {
	inner := max(height-2, 1) // minus borders
	history := m.game.History()

	var longest time.Duration
	for _, d := range m.moveTimes {
		longest = max(longest, d)
	}
	var rows []string
	for i, rec := range history {
		num := "    "
		if rec.Color == chess.White {
			num = fmt.Sprintf("%3d.", i/2+1)
		}
		var spent time.Duration
		if i < len(m.moveTimes) {
			spent = m.moveTimes[i]
		}
		bar := whiteBarStyle
		if rec.Color == chess.Black {
			bar = blackBarStyle
		}
		rows = append(rows, dimStyle.Render(num)+" "+pad(m.sanCell(history, i), 8)+
			bar.Render(timeBar(spent, longest))+" "+dimStyle.Render(fmt.Sprintf("%6s", formatMoveTime(spent))))
	}
	visible := inner - 1 // first line is the title
	start := m.historyStart(len(rows), visible)
	rows = rows[start:min(len(rows), start+visible)]
	if len(rows) == 0 {
		rows = []string{dimStyle.Render("  nenhum lance ainda")}
	}
	body := append([]string{dimStyle.Render("Lances")}, rows...)
	for len(body) < inner {
		body = append(body, "")
	}
	return boxStyle.Width(panelW - 2).Render(strings.Join(body, "\n"))
}

// historyStart is the first history row shown when visible rows fit: the
// latest moves, or, while reviewing an earlier move, a window holding it.
// Windows start on White's move so each move's two lines stay together.
func (m *Model) historyStart(n, visible int) int {
	if visible <= 0 || n <= visible {
		return 0
	}
	start := n - visible
	if start%2 == 1 {
		start++
	}
	if focus := m.reviewPly - 1; m.reviewing && focus < start {
		start = max(focus-focus%2, 0)
	}
	return start
}

// historyRows is the height of the move list inside the history box, and
// the box's top row relative to the board.
func (m *Model) historyRows(l layout) (top, visible int) {
	orient := m.viewOrientation()
	topH := lipgloss.Height(m.clockView(orient.Other(), l.compact()))
	bottomH := lipgloss.Height(m.clockView(orient, l.compact()))
	height := l.boardH() - topH - bottomH
	return topH, max(height-2, 1) - 1 // borders, then the title line
}

// historyAt maps a click to the move list: the ply after the clicked move.
func (m *Model) historyAt(x, y int) (int, bool) {
	l := computeLayout(m.width, m.height)
	panelX := l.boardX + 8*l.cellW + gapW
	if x < panelX || x >= panelX+panelW {
		return 0, false
	}
	top, visible := m.historyRows(l)
	row := y - (l.boardY + top + 2) // border and title
	n := m.game.Ply()
	if row < 0 || row >= visible {
		return 0, false
	}
	i := m.historyStart(n, visible) + row
	if i >= n {
		return 0, false
	}
	return i + 1, true
}

// sanCell renders a move's SAN, highlighting the move shown on the board.
func (m *Model) sanCell(h []chess.MoveRecord, i int) string {
	shown := len(h) - 1
	if m.reviewing {
		shown = m.reviewPly - 1
	}
	if i == shown {
		return m.st.accent.Render(h[i].SAN)
	}
	return textStyle.Render(h[i].SAN)
}

// timeBarWidth is the length, in cells, of the longest move's bar.
const timeBarWidth = 4

var barEighths = []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// timeBar draws d relative to longest as a bar of up to timeBarWidth cells,
// with eighth-block precision.
func timeBar(d, longest time.Duration) string {
	width := timeBarWidth
	if longest <= 0 || d <= 0 {
		return strings.Repeat(" ", width)
	}
	eighths := max(int(math.Round(float64(d)/float64(longest)*float64(width*8))), 1)
	bar := strings.Repeat("█", eighths/8)
	if r := eighths % 8; r > 0 {
		bar += barEighths[r-1]
	}
	return pad(bar, width)
}
