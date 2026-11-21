package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/sound"
)

// startGame starts a new game with the given time control.
func (m *Model) startGame(tc clock.TimeControl) tea.Cmd {
	m.screen = screenGame
	m.game = chess.NewGame()
	m.stopComputer()
	m.notice = ""
	m.vs = m.resolveOpponent()
	m.engineGame++
	m.resetAnalysis()
	m.assistTakeback, m.assistHint = false, false
	m.tc = tc
	m.clock = nil
	if !tc.Unlimited() {
		m.clock = clock.New(tc, m.now)
	}
	m.orientation = m.facing()
	m.mode = modePlay
	m.cursorCol, m.cursorRow = 4, 6 // e2
	m.showCursor = false
	m.anim = nil
	m.moveTimes = nil
	m.turnStart = m.now()
	m.input, m.inputErr = "", ""
	m.lowWarned = [2]bool{}
	m.moveClocks, m.clockFrom = nil, 0
	m.savedPath = ""
	m.reviewing, m.help = false, false
	m.clearSelection()
	m.tickGen++
	m.resume = nil // a new game replaces the saved one
	cmds := []tea.Cmd{m.files.clearResume()}
	if m.clock != nil {
		cmds = append(cmds, tick(m.tickGen))
	}
	return tea.Batch(append(cmds, m.maybeComputer())...)
}

// toMenu leaves the game for the time-control menu.
func (m *Model) toMenu() {
	m.stopComputer()
	m.screen = screenMenu
	m.anim = nil
	m.game = nil
	m.clock = nil
	m.reviewing = false
	m.tickGen++
	m.menu.restore(m.menuEntries(), m.cfg.TimeControl)
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
	m.hintSet = false

	// The opponent's time starts once the animation is over: the slide
	// (skipped for a dropped piece) and the turn of the board, if it turns.
	var delay time.Duration
	if animated {
		if m.flip && !m.vs.computer {
			delay = flipDur
		}
		if !dropped {
			delay += slideDur
		}
	}
	var pr moveClock
	if m.clock != nil {
		pr.before, pr.running = m.clock.Remaining(mover), m.clock.Running()
		m.clock.SetStartDelay(delay)
		if m.clock.Press(mover) {
			return m.conclude(m.game.FlagFall(mover))
		}
	}
	if err := m.game.Play(mv); err != nil {
		m.notice = "Illegal move"
		return nil
	}
	now := m.now()
	m.moveTimes = append(m.moveTimes, max(now.Sub(m.turnStart), 0))
	m.turnStart = now.Add(delay)
	if m.clock != nil {
		pr.after = m.clock.Remaining(mover)
		m.moveClocks = append(m.moveClocks, pr)
	}
	var save tea.Cmd
	if m.game.IsOver() {
		save = m.finish()
	} else {
		m.orientation = m.facing()
		m.sfx(moveEffect(mv, m.game.InCheck()))
		save = m.files.saveResume(m.gamePGN())
	}
	next := m.maybeComputer()
	if animated {
		return tea.Batch(save, m.startAnimation(before, mv, prevOrient, m.orientation, dropped), next)
	}
	return tea.Batch(save, next)
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
		return m.conclude(m.game.FlagFall(c))
	}
	m.checkLowTime()
	return tick(m.tickGen)
}

// conclude finishes the game after an engine call that ends it. The engine
// only refuses when the game is already over, which the modes rule out.
func (m *Model) conclude(err error) tea.Cmd {
	if err != nil {
		m.notice = err.Error()
		return nil
	}
	return m.finish()
}

// finish stops the clock and shows the result. The game moves from the
// resume file, since a finished game has nothing to resume, to the games
// folder, so every finished game can be opened again.
func (m *Model) finish() tea.Cmd {
	if m.clock != nil {
		m.clock.Stop()
	}
	m.stopComputer()
	m.mode = modeGameOver
	m.reviewing = false
	m.resultHidden = false
	m.notice = ""
	m.clearSelection()
	m.sfx(sound.GameEnd)
	return tea.Batch(m.files.clearResume(), m.saveFinished())
}
