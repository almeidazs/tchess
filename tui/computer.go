package tui

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/uci"
)

// opponentChoice is what the player picked in the menu for the next games.
type opponentChoice struct {
	computer bool
	elo      int    // computer strength, 0 for full
	color    string // the human's side: "white", "black" or "random"
}

// opponent is who plays against the human in the current game.
type opponent struct {
	computer bool
	elo      int
	human    chess.Color // the human's side against the computer
}

// evalScore is an evaluation from White's point of view.
type evalScore struct {
	cp   int // centipawns, when mate is 0
	mate int // moves to mate: positive when White mates
	over bool
}

// resolveOpponent turns the menu choice into the opponent of a new game.
// Without an engine, a computer opponent falls back to two players.
func (m *Model) resolveOpponent() opponent {
	if !m.choice.computer {
		return opponent{}
	}
	if !m.engine.available() {
		m.notice = "No engine found: playing two-player"
		return opponent{}
	}
	vs := opponent{computer: true, elo: m.choice.elo, human: chess.White}
	switch m.choice.color {
	case "black":
		vs.human = chess.Black
	case "random":
		if rand.IntN(2) == 1 {
			vs.human = chess.Black
		}
	}
	return vs
}

// humanTurn reports whether the human may move now.
func (m *Model) humanTurn() bool {
	return !m.vs.computer || m.game.Turn() == m.vs.human
}

// computerName is how the computer is called on screen.
func (m *Model) computerName() string {
	if m.engineName != "" {
		return m.engineName
	}
	return m.engine.displayName()
}

// playerLabel names a side in the clock panel.
func (m *Model) playerLabel(c chess.Color) string {
	if !m.vs.computer {
		return colorName(c)
	}
	if c == m.vs.human {
		return "You"
	}
	return m.computerName()
}

// gameMoves lists the moves played so far in coordinate notation.
func gameMoves(g *chess.Game, ply int) []string {
	h := g.History()
	ply = min(ply, len(h))
	out := make([]string, ply)
	for i := range out {
		out[i] = h[i].Move.UCI()
	}
	return out
}

// job prepares a search of the position after ply plies.
func (m *Model) job(kind jobKind, ply int) searchJob {
	start := m.game.StartPosition()
	return searchJob{kind: kind, ply: ply, fen: start.FEN(), moves: gameMoves(m.game, ply), game: m.engineGame}
}

// maybeComputer starts the computer's search when it is its turn.
func (m *Model) maybeComputer() tea.Cmd {
	if !m.vs.computer || m.game == nil || m.game.IsOver() || m.humanTurn() || m.thinking {
		return nil
	}
	j := m.job(jobMove, m.game.Ply())
	j.elo = m.vs.elo
	if m.clock != nil {
		j.clocked = true
		j.clock = uci.Limits{
			WTime: m.clock.Remaining(chess.White), BTime: m.clock.Remaining(chess.Black),
			WInc: m.tc.Increment, BInc: m.tc.Increment,
		}
	}
	m.thinking = true
	m.evalPending = -1
	return m.engine.run(j)
}

// stopComputer cancels whatever the engine is doing for this game.
func (m *Model) stopComputer() {
	m.engine.stop()
	m.thinking = false
	m.evalPending = -1
}

// onEngine applies a search result.
func (m *Model) onEngine(msg engineMsg) tea.Cmd {
	if msg.name != "" {
		m.engineName = msg.name
	}
	if msg.job != m.engine.job || m.screen != screenGame || m.game == nil {
		return nil // a canceled or outdated search
	}
	if msg.kind == jobMove {
		m.thinking = false
	}
	if msg.kind == jobEval {
		m.evalPending = -1
	}
	if errors.Is(msg.err, context.Canceled) {
		return nil
	}
	if msg.err != nil {
		return m.engineFailed(msg.err)
	}
	switch msg.kind {
	case jobMove:
		return m.computerMoves(msg)
	case jobHint:
		return m.showHint(msg)
	case jobDraw:
		return m.answerDraw(msg)
	case jobEval:
		m.storeEval(msg.ply, msg.res.Score)
	}
	return nil
}

// engineFailed reports an engine that cannot play. A game against the
// computer goes on between two players.
func (m *Model) engineFailed(err error) tea.Cmd {
	m.thinking = false
	if m.vs.computer && !m.game.IsOver() {
		m.vs = opponent{}
		m.orientation = m.facing()
		m.say("Engine error (" + err.Error() + "): playing two-player")
		return nil
	}
	m.say("Engine error: " + err.Error())
	return nil
}

// computerMoves plays the computer's move.
func (m *Model) computerMoves(msg engineMsg) tea.Cmd {
	if m.game.IsOver() || msg.ply != m.game.Ply() || m.humanTurn() {
		return nil
	}
	pos := m.game.Position()
	mv, err := pos.ParseUCI(msg.res.BestMove)
	if err != nil {
		return m.engineFailed(fmt.Errorf("illegal move %q", msg.res.BestMove))
	}
	keep := m.mode // a question the human has open stays open
	cmd := m.play(mv)
	if !m.game.IsOver() && (keep == modeConfirmQuit || keep == modeConfirmResign) {
		m.mode = keep
	}
	return cmd
}

// computerUndo takes the game back to target plies at once: against the
// computer there is nobody to ask.
func (m *Model) computerUndo(target int) tea.Cmd {
	if target < 0 || target >= m.game.Ply() {
		m.notice = "No move to take back"
		return nil
	}
	m.stopComputer()
	m.undoTarget = target
	cmd := m.undo()
	m.assistTakeback = true
	return tea.Batch(cmd, m.maybeComputer())
}

// undoAgainstComputer works out how far u goes: the human's last move,
// and the computer's reply to it when there is one.
func (m *Model) undoAgainstComputer() tea.Cmd {
	target := m.game.Ply() - 2
	if !m.humanTurn() { // the computer is still thinking about its reply
		target = m.game.Ply() - 1
	}
	return m.computerUndo(target)
}

// requestHint asks the engine for the best move.
func (m *Model) requestHint() tea.Cmd {
	if !m.vs.computer || m.game.IsOver() || m.reviewing || !m.humanTurn() {
		return nil
	}
	m.notice = "Looking for a hint…"
	j := m.job(jobHint, m.game.Ply())
	j.moveTime = hintTime
	return m.engine.run(j)
}

// showHint highlights the suggested move.
func (m *Model) showHint(msg engineMsg) tea.Cmd {
	if msg.ply != m.game.Ply() || m.game.IsOver() {
		return nil
	}
	pos := m.game.Position()
	mv, err := pos.ParseUCI(msg.res.BestMove)
	if err != nil {
		return nil
	}
	m.hint, m.hintPly, m.hintSet = mv, msg.ply, true
	m.assistHint = true
	m.notice = "Hint: " + pos.SAN(mv)
	return m.files.saveResume(m.gamePGN())
}

// offerDrawToComputer asks the engine whether to accept a draw.
func (m *Model) offerDrawToComputer() tea.Cmd {
	m.notice = "The computer is considering your offer…"
	j := m.job(jobDraw, m.game.Ply())
	j.moveTime = drawTime
	return m.engine.run(j)
}

// answerDraw accepts the draw when the computer is not better. The score
// is from the side to move's view, which is the human's.
func (m *Model) answerDraw(msg engineMsg) tea.Cmd {
	if msg.ply != m.game.Ply() || m.game.IsOver() {
		return nil
	}
	s := msg.res.Score
	accept := s.Set && (s.Mate > 0 || (s.Mate == 0 && s.CP >= 0))
	if !accept {
		m.notice = "The computer declines the draw"
		return nil
	}
	m.notice = ""
	return m.conclude(m.game.AgreeDraw())
}

// evalVisible reports whether the evaluation is shown: while reviewing
// and after the game, never while it is being played.
func (m *Model) evalVisible() bool {
	return m.engine.available() && m.screen == screenGame && m.game != nil &&
		(m.reviewing || m.mode == modeGameOver)
}

// shownPly is the number of plies in the position on the board.
func (m *Model) shownPly() int {
	if m.reviewing {
		return m.reviewPly
	}
	return m.game.Ply()
}

// autoEval asks for the evaluation of the shown position when it is
// visible and not known yet. It never interrupts the computer's search.
func (m *Model) autoEval() tea.Cmd {
	if !m.evalVisible() || m.thinking {
		return nil
	}
	ply := m.shownPly()
	if _, ok := m.evals[ply]; ok || m.evalPending == ply {
		return nil
	}
	pos := m.game.PositionAt(ply)
	var legal chess.MoveList
	pos.LegalMoves(&legal)
	if legal.Len() == 0 { // the game is decided on the board
		e := evalScore{over: true}
		if pos.InCheck() {
			e.mate = 1
			if pos.SideToMove() == chess.White {
				e.mate = -1
			}
		}
		m.evals[ply] = e
		return nil
	}
	m.evalPending = ply
	j := m.job(jobEval, ply)
	j.moveTime = evalTime
	return m.engine.run(j)
}

// storeEval keeps an evaluation, turned to White's point of view.
func (m *Model) storeEval(ply int, s uci.Score) {
	if !s.Set {
		return
	}
	e := evalScore{cp: s.CP, mate: s.Mate}
	if pos := m.game.PositionAt(ply); pos.SideToMove() == chess.Black {
		e.cp, e.mate = -e.cp, -e.mate
	}
	m.evals[ply] = e
}

// text is the evaluation as chess players write it: +1.2, -0.4, M3,
// -M2; "#" when the game is decided on the board.
func (e evalScore) text() string {
	switch {
	case e.over && e.mate == 0:
		return "0.0"
	case e.over:
		return map[bool]string{true: "1-0", false: "0-1"}[e.mate > 0]
	case e.mate > 0:
		return fmt.Sprintf("M%d", e.mate)
	case e.mate < 0:
		return fmt.Sprintf("-M%d", -e.mate)
	}
	return fmt.Sprintf("%+.1f", float64(e.cp)/100)
}

// resetAnalysis forgets evaluations and hints, as when the game changes.
func (m *Model) resetAnalysis() {
	m.evals = map[int]evalScore{}
	m.evalPending = -1
	m.hintSet = false
}
