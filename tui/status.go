package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
)

// headerView is the title line: name, time control and move number.
func (m *Model) headerView(l layout) string {
	tc := "sem relógio"
	if !m.tc.Unlimited() {
		tc = m.tc.String()
	}
	left := m.st.title.Render("♞ Xadrez") + dimStyle.Render("  ·  "+tc)
	right := dimStyle.Render(fmt.Sprintf("lance %d", m.moveNumber()))
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
	yn := m.st.key.Render("[y]") + textStyle.Render(" sim   ") + m.st.key.Render("[n]") + textStyle.Render(" não")

	if m.reviewing && (m.mode == modePlay || m.mode == modeGameOver) {
		return m.reviewStatus()
	}
	switch m.mode {
	case modeUndoOffer:
		return m.st.accent.Render(m.undoStatus()) + yn
	case modeInput:
		return m.st.key.Render(": ") + textStyle.Render(m.input) + m.st.accent.Render("▌") + "  " + m.inputHint()
	case modePromotion:
		return textStyle.Render("Promoção: escolha a peça (") + m.st.key.Render("↑↓ enter") +
			textStyle.Render(" ou clique · ") + m.st.key.Render("esc") + textStyle.Render(" cancela)")
	case modeConfirmResign:
		return dangerStyle.Render("As "+turn+" desistem?  ") + yn
	case modeDrawOffer:
		return m.st.accent.Render("As "+turn+" propõem empate. "+other+", aceitam?  ") + yn
	case modeConfirmQuit:
		return m.st.accent.Render("Sair da partida?  ") + yn
	case modeGameOver:
		s := m.st.accent.Render(outcomeHeadline(m.game.Outcome())) +
			dimStyle.Render(" · "+outcomeScore(m.game.Outcome())+" · "+terminationText(m.game.Termination()))
		if m.notice != "" {
			s += dimStyle.Render("  · " + m.notice)
		}
		return s
	}

	s := textStyle.Render("Vez das ") + m.st.accent.Render(turn)
	if m.game.InCheck() {
		s += dangerStyle.Render("  · Xeque!")
	}
	if claim := m.game.ClaimableDraw(); claim != chess.NoTermination {
		s += textStyle.Render("  · ") + m.st.key.Render("d") +
			textStyle.Render(": reivindicar empate ("+terminationText(claim)+")")
	}
	if m.notice != "" {
		s += dimStyle.Render("  · " + m.notice)
	}
	return s
}

// reviewStatus says which move the board shows.
func (m *Model) reviewStatus() string {
	s := m.st.accent.Render("Revendo") + textStyle.Render(": ")
	if m.reviewPly == 0 {
		s += textStyle.Render("posição inicial")
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
	s += dimStyle.Render(fmt.Sprintf("  (%d de %d)", m.reviewPly, m.game.Ply()))
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
		return m.helpLine([][2]string{{step, "lance anterior/seguinte"}, {"home end", "início/fim"}, {"esc", "voltar ao fim"}})
	}
	switch m.mode {
	case modeGameOver:
		return m.helpLine([][2]string{{"←→", "rever lances"}, {"esc", "resultado"}, {"n", "nova partida"}, {"e", "exportar"}, {"?", "ajuda"}, {"q", "sair"}})
	case modeInput:
		return m.helpLine([][2]string{{"enter", "jogar"}, {"esc", "cancelar"}})
	case modePlay:
		return m.helpLine([][2]string{{"←↑↓→ enter", "mover"}, {":", "digitar"}, {"u", "voltar"},
			{"[ ]", "rever"}, {"d", "empate"}, {"r", "desistir"}, {"?", "ajuda"}, {"q", "sair"}})
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
	keys := m.st.key.Background(panelBg).Render("n") + bg.Render(" ") + dimStyle.Background(panelBg).Render("nova partida") +
		dimStyle.Background(panelBg).Render("  ·  ") +
		m.st.key.Background(panelBg).Render("q") + bg.Render(" ") + dimStyle.Background(panelBg).Render("sair")
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
