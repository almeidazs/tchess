package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/internal/pixel"
)

// glyphBoard draws the compact board: one Unicode glyph per square.
func (m *Model) glyphBoard(l layout) []string {
	th := m.theme()
	bs := m.boardState(l)
	pointCol, pointRow := -1, -1
	if bs.drag.active && bs.drag.x >= 0 && bs.drag.y >= 0 {
		pointCol, pointRow = bs.drag.x/l.cellW, bs.drag.y/l.cellH
	}
	lines := make([]string, 0, 8*l.cellH)
	for row := range 8 {
		var sb strings.Builder
		for col := range 8 {
			lk := m.squareLook(&bs, col, row)
			bg := lk.bg
			switch {
			case lk.cursor:
				bg = shade(lk.base == th.Light, th.CursorLight, th.CursorDark)
			case lk.capture:
				bg = shade(lk.base == th.Light, th.CaptureLight, th.CaptureDark)
			}
			content, fg := " ", pieceFg
			switch {
			case col == pointCol && row == pointRow:
				content = glyph(bs.drag.piece)
			case lk.ghost:
			case lk.piece != chess.NoPiece:
				content = glyph(lk.piece)
			case lk.dot:
				content, fg = "•", lipgloss.Color(pixel.Blend(bg, th.Mark, 0.7).Hex())
			}
			left := (l.cellW - 1) / 2
			cell := strings.Repeat(" ", left) + content + strings.Repeat(" ", l.cellW-1-left)
			sb.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(bg.Hex())).Foreground(fg).Bold(true).Render(cell))
		}
		lines = append(lines, sb.String())
	}
	return lines
}
