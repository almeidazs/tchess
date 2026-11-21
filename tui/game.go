package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/clock"
	"github.com/almeidazs/chess/internal/sound"
)

// startGame starts a new game with the given time control.
func (m *Model) startGame(tc clock.TimeControl) tea.Cmd {
	m.screen = screenGame
	m.game = chess.NewGame()
	m.tc = tc
	m.clock = nil
	if !tc.Unlimited() {
		m.clock = clock.New(tc, m.now)
	}
	m.orientation = chess.White
	m.mode = modePlay
	m.cursorCol, m.cursorRow = 4, 6 // e2
	m.showCursor = false
	m.notice = ""
	m.anim = nil
	m.moveTimes = nil
	m.turnStart = m.now()
	m.input, m.inputErr = "", ""
	m.lowWarned = [2]bool{}
	m.clearSelection()
	m.tickGen++
	if m.clock != nil {
		return tick(m.tickGen)
	}
	return nil
}

// toMenu leaves the game for the time-control menu.
func (m *Model) toMenu() {
	m.screen = screenMenu
	m.anim = nil
	m.game = nil
	m.clock = nil
	m.tickGen++
}

// play presses the clock, makes the move and starts the animation. If the
// mover's flag fell before the move was completed, the move does not count.
func (m *Model) play(mv chess.Move) tea.Cmd {
	mover, prevOrient := m.game.Turn(), m.orientation
	before := m.game.Position()
	animated := m.animate && computeLayout(m.width, m.height).pixel
	dropped := m.dropped
	m.dropped = false
	m.mode = modePlay
	m.clearSelection()

	// The opponent's time starts once the board has finished turning.
	var delay time.Duration
	if animated {
		delay = flipDur
		if !dropped {
			delay += slideDur
		}
	}
	if m.clock != nil {
		m.clock.SetStartDelay(delay)
		if m.clock.Press(mover) {
			m.conclude(m.game.FlagFall(mover))
			return nil
		}
	}
	if err := m.game.Play(mv); err != nil {
		m.notice = "Lance ilegal"
		return nil
	}
	now := m.now()
	m.moveTimes = append(m.moveTimes, max(now.Sub(m.turnStart), 0))
	m.turnStart = now.Add(delay)
	if m.game.IsOver() {
		m.finish()
	} else {
		m.orientation = m.game.Turn()
		m.sfx(moveEffect(mv, m.game.InCheck()))
	}
	if animated {
		return m.startAnimation(before, mv, prevOrient, m.orientation, dropped)
	}
	return nil
}

// orderPromotions sorts promotion candidates as the picker shows them: Q R B N.
func orderPromotions(moves []chess.Move) []chess.Move {
	out := make([]chess.Move, 0, 4)
	for _, t := range []chess.PieceType{chess.Queen, chess.Rook, chess.Bishop, chess.Knight} {
		for _, mv := range moves {
			if mv.Promotion() == t {
				out = append(out, mv)
			}
		}
	}
	return out
}

// onTick refreshes the clock and detects flag falls.
func (m *Model) onTick(msg tickMsg) tea.Cmd {
	if msg.gen != m.tickGen || m.screen != screenGame || m.clock == nil || m.game.IsOver() {
		return nil
	}
	if c, flagged := m.clock.Flagged(); flagged {
		m.conclude(m.game.FlagFall(c))
		return nil
	}
	m.checkLowTime()
	return tick(m.tickGen)
}

// conclude finishes the game after an engine call that ends it. The engine
// only refuses when the game is already over, which the modes rule out.
func (m *Model) conclude(err error) {
	if err != nil {
		m.notice = err.Error()
		return
	}
	m.finish()
}

// finish stops the clock and shows the result.
func (m *Model) finish() {
	if m.clock != nil {
		m.clock.Stop()
	}
	m.mode = modeGameOver
	m.clearSelection()
	m.sfx(sound.GameEnd)
}
