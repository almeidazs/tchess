package tui

import (
	"math"

	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
)

// The evaluation bar sits in the middle column of the gap between the board
// and the panel. White's share grows from White's side of the board, and
// each terminal row holds two half-block pixels.

var (
	evalWhite = lipgloss.Color("#EDE7DA")
	evalBlack = lipgloss.Color("#57545F")
)

// whiteShare turns an evaluation into White's share of the bar, from 0 to
// 1, with the winning-chances curve Lichess uses: +1 pawn is about 59%.
func (e evalScore) whiteShare() float64 {
	switch {
	case e.mate > 0:
		return 1
	case e.mate < 0:
		return 0
	case e.over:
		return 0.5
	}
	chances := 2/(1+math.Exp(-0.00368208*float64(e.cp))) - 1
	return (chances + 1) / 2
}

// evalBar renders the bar for rows terminal rows, top to bottom, or nil
// when there is no evaluation to show.
func (m *Model) evalBar(rows int) []string {
	if !m.evalVisible() {
		return nil
	}
	e, ok := m.evals[m.shownPly()]
	if !ok {
		return nil
	}
	pixels := 2 * rows
	white := int(math.Round(e.whiteShare() * float64(pixels)))
	whiteBelow := m.viewOrientation() == chess.White
	isWhite := func(p int) bool {
		if whiteBelow {
			return p >= pixels-white
		}
		return p < white
	}
	colorOf := func(p int) lipgloss.Color {
		if isWhite(p) {
			return evalWhite
		}
		return evalBlack
	}
	out := make([]string, rows)
	for r := range out {
		out[r] = lipgloss.NewStyle().Foreground(colorOf(2 * r)).Background(colorOf(2*r + 1)).Render("▀")
	}
	return out
}
