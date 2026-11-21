package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
)

// headerView is the title line: name, time control and move number.
func (m *Model) headerView(l layout) string {
	tc := "no clock"
	if !m.tc.Unlimited() {
		tc = m.tc.String()
	}
	left := m.st.title.Render("♞ tchess") + dimStyle.Render("  ·  "+tc)
	right := dimStyle.Render(fmt.Sprintf("move %d", m.moveNumber()))
	return left + strings.Repeat(" ", max(l.totalW-lipgloss.Width(left)-lipgloss.Width(right), 1)) + right
}

// moveNumber returns the current full-move number.
func (m *Model) moveNumber() int {
	pos := m.game.Position()
	return pos.FullmoveNumber()
}

// statusView describes the turn, checks and claimable draws, or the
// question being asked in the current mode.
func (m *Model) statusView() string {
	turn := colorName(m.game.Turn())
	other := colorName(m.game.Turn().Other())
	yn := m.st.key.Render("[y]") + textStyle.Render(" yes   ") + m.st.key.Render("[n]") + textStyle.Render(" no")

	if m.reviewing && (m.mode == modePlay || m.mode == modeGameOver) {
		return m.reviewStatus()
	}
	switch m.mode {
	case modeUndoOffer:
		return m.st.accent.Render(m.undoStatus()) + yn
	case modeInput:
		return m.st.key.Render(": ") + textStyle.Render(m.input) + m.st.accent.Render("▌") + "  " + m.inputHint()
	case modePromotion:
		return textStyle.Render("Promotion: choose a piece (") + m.st.key.Render("↑↓ enter") +
			textStyle.Render(" or click · ") + m.st.key.Render("esc") + textStyle.Render(" cancels)")
	case modeConfirmResign:
		return dangerStyle.Render(turn+" resigns?  ") + yn
	case modeDrawOffer:
		return m.st.accent.Render(turn+" offers a draw. "+other+", accept?  ") + yn
	case modeConfirmQuit:
		return m.st.accent.Render("Quit the game?  ") + yn
	case modeGameOver:
		s := m.st.accent.Render(outcomeHeadline(m.game.Outcome())) +
			dimStyle.Render(" · "+outcomeScore(m.game.Outcome())+" · "+terminationText(m.game.Termination()))
		if m.notice != "" {
			s += dimStyle.Render("  · " + m.notice)
		}
		return s
	}

	s := m.st.accent.Render(turn) + textStyle.Render(" to move")
	if m.game.InCheck() {
		s += dangerStyle.Render("  · Check!")
	}
	if claim := m.game.ClaimableDraw(); claim != chess.NoTermination {
		s += textStyle.Render("  · ") + m.st.key.Render("d") +
			textStyle.Render(": claim a draw ("+terminationText(claim)+")")
	}
	if m.notice != "" {
		s += dimStyle.Render("  · " + m.notice)
	}
	return s
}

// reviewStatus says which move the board shows.
func (m *Model) reviewStatus() string {
	s := m.st.accent.Render("Reviewing") + textStyle.Render(": ")
	if m.reviewPly == 0 {
		s += textStyle.Render("starting position")
	} else {
		start := m.game.StartPosition()
		idx := m.reviewPly - 1 // half-moves since the start's white move
		if start.SideToMove() == chess.Black {
			idx++
		}
		num := fmt.Sprintf("%d.", start.FullmoveNumber()+idx/2)
		if idx%2 == 1 {
			num += ".."
		}
		s += textStyle.Render(num + m.game.History()[m.reviewPly-1].SAN)
	}
	s += dimStyle.Render(fmt.Sprintf("  (%d of %d)", m.reviewPly, m.game.Ply()))
	if m.notice != "" {
		s += dimStyle.Render("  · " + m.notice)
	}
	return s
}

// helpView lists the keys available in the current mode.
func (m *Model) helpView() string {
	if m.reviewing && (m.mode == modePlay || m.mode == modeGameOver) {
		step := "[ ]"
		if m.mode == modeGameOver {
			step = "←→ [ ]"
		}
		return m.helpLine([][2]string{{step, "previous/next move"}, {"home end", "start/end"}, {"esc", "back to the end"}})
	}
	switch m.mode {
	case modeGameOver:
		return m.helpLine([][2]string{{"←→", "review"}, {"esc", "result"}, {"n", "new game"}, {"e", "export"}, {"?", "help"}, {"q", "quit"}})
	case modeInput:
		return m.helpLine([][2]string{{"enter", "play"}, {"esc", "cancel"}})
	case modePlay:
		return m.helpLine([][2]string{{"←↑↓→ enter", "move"}, {":", "type"}, {"u", "take back"},
			{"[ ]", "review"}, {"d", "draw"}, {"r", "resign"}, {"?", "help"}, {"q", "quit"}})
	}
	return ""
}

// resultBox is the game-over card drawn over the board.
func (m *Model) resultBox() string {
	o := m.game.Outcome()
	headline := outcomeHeadline(o)
	color := m.st.accentColor
	if o == chess.Draw {
		color = success
	}
	bg := lipgloss.NewStyle().Background(panelBg)
	keys := m.st.key.Background(panelBg).Render("n") + bg.Render(" ") + dimStyle.Background(panelBg).Render("new game") +
		dimStyle.Background(panelBg).Render("  ·  ") +
		m.st.key.Background(panelBg).Render("q") + bg.Render(" ") + dimStyle.Background(panelBg).Render("quit")
	body := strings.Join([]string{
		cardLine(lipgloss.NewStyle().Bold(true).Foreground(color).Background(panelBg).Render(headline), 28, lipgloss.Center),
		cardLine(textStyle.Background(panelBg).Render(outcomeScore(o)+" · "+terminationText(m.game.Termination())), 28, lipgloss.Center),
		cardLine("", 28, lipgloss.Center),
		cardLine(keys, 28, lipgloss.Center),
	}, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Background(panelBg).
		Padding(1, 2).
		Render(body)
}
