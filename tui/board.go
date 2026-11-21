package tui

import (
	"strings"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/internal/pixel"
)

// boardState gathers the per-frame data the square renderers need.
func (m *Model) boardState(l layout) boardState {
	bs := boardState{pos: m.shownPosition(), checkSq: chess.NoSquare, pickCol: -1, drag: m.dragView(l)}
	bs.last, bs.hasLast = m.shownLastMove()
	if bs.pos.InCheck() {
		bs.checkSq = bs.pos.KingSquare(bs.pos.SideToMove())
	}
	for _, mv := range m.targets {
		bs.targets |= mv.To().Bitboard()
	}
	if m.mode == modePromotion {
		bs.pickCol, bs.pickRow0 = m.cellOf(m.promo[0].To())
	}
	if m.mode == modeInput {
		var targets chess.Bitboard
		bs.marked, targets = m.inputMarks()
		bs.targets |= targets
	}
	return bs
}

// squareLook decides how the square at screen cell (col, row) is drawn.
func (m *Model) squareLook(bs *boardState, col, row int) squareLook {
	th := m.theme()
	sq := m.squareAt(col, row)
	light := sq.IsLight()
	lk := squareLook{base: shade(light, th.Light, th.Dark), piece: bs.pos.PieceAt(sq)}
	lk.bg = lk.base

	if pickIdx := abs(row - bs.pickRow0); col == bs.pickCol && pickIdx < len(m.promo) {
		lk.bg = pickerRGB
		if pickIdx == m.promoIdx {
			lk.bg = th.Accent
		}
		lk.piece = chess.MakePiece(bs.pos.SideToMove(), m.promo[pickIdx].Promotion())
		return lk
	}

	switch {
	case sq == bs.checkSq:
		lk.check = true
		lk.bg = checkRGB
	case sq == m.selected || bs.marked.Has(sq):
		lk.bg = shade(light, th.SelLight, th.SelDark)
	case bs.hasLast && (sq == bs.last.From() || sq == bs.last.To()):
		lk.bg = shade(light, th.LastLight, th.LastDark)
	}
	if sq == bs.drag.hover {
		lk.bg = pixel.Blend(lk.bg, hoverRGB, 0.4)
	}
	if bs.targets.Has(sq) {
		lk.dot = lk.piece == chess.NoPiece
		lk.capture = !lk.dot
	}
	lk.ghost = bs.drag.active && sq == bs.drag.from
	lk.cursor = m.showCursor && m.mode == modePlay && !m.reviewing && col == m.cursorCol && row == m.cursorRow
	return lk
}

// boardView renders the board with rank and file labels for the
// orientation currently on screen.
func (m *Model) boardView(l layout) []string {
	var squares []string
	if l.pixel {
		squares = m.pixelBoard(l)
	} else {
		squares = m.glyphBoard(l)
	}
	orient := m.viewOrientation()
	out := make([]string, 0, l.boardH())
	mid := l.cellH / 2
	for i, line := range squares {
		label := strings.Repeat(" ", rankLabelW)
		if i%l.cellH == mid {
			rank := 8 - i/l.cellH
			if orient == chess.Black {
				rank = i/l.cellH + 1
			}
			label = dimStyle.Render(" " + string(rune('0'+rank)) + " ")
		}
		out = append(out, label+line)
	}
	var files strings.Builder
	files.WriteString(strings.Repeat(" ", rankLabelW))
	for col := range 8 {
		f := 'a' + rune(col)
		if orient == chess.Black {
			f = 'h' - rune(col)
		}
		files.WriteString(dimStyle.Render(center(string(f), l.cellW)))
	}
	return append(out, files.String())
}
