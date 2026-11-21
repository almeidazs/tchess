package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
)

// Games are written as standard PGN. Each move carries the mover's clock
// after the move ([%clk h:mm:ss.s]) and the time spent on it
// ([%emt h:mm:ss.s]), the comment commands Lichess and other programs read.
// The same format keeps the unfinished game for resuming.

// gamePGN renders the current game as PGN.
func (m *Model) gamePGN() string {
	tags := []chess.Tag{
		{Name: "Event", Value: "Casual game"},
		{Name: "Site", Value: "tchess"},
		{Name: "Date", Value: m.now().Format("2006.01.02")},
		{Name: "Round", Value: "-"},
		{Name: "White", Value: m.pgnPlayer(chess.White)},
		{Name: "Black", Value: m.pgnPlayer(chess.Black)},
		{Name: "Result", Value: "*"},
		{Name: "TimeControl", Value: pgnTimeControl(m.tc)},
	}
	tags = append(tags, m.opponentTags()...)
	if t := m.game.Termination(); t != chess.NoTermination {
		tags = append(tags, chess.Tag{Name: "Termination", Value: terminationText(t)})
	}
	comments := make([]string, m.game.Ply())
	for i := range comments {
		var parts []string
		if j := i - m.clockFrom; j >= 0 && j < len(m.moveClocks) {
			parts = append(parts, "[%clk "+formatPGNClock(m.moveClocks[j].after)+"]")
		}
		if i < len(m.moveTimes) {
			parts = append(parts, "[%emt "+formatPGNClock(m.moveTimes[i])+"]")
		}
		comments[i] = strings.Join(parts, " ")
	}
	return m.game.PGN(tags, comments)
}

// Tags tchess adds to remember a game against the computer, so it can be
// resumed against the same opponent.
const (
	tagOpponent = "TchessOpponent" // "computer"
	tagElo      = "TchessElo"      // the computer's strength: an Elo or "max"
	tagHuman    = "TchessHuman"    // the human's side: "white" or "black"
	tagAssist   = "TchessAssist"   // help used: "takebacks", "hints"
)

// pgnPlayer names a side in the PGN: the engine and its strength for the
// computer, "Player" for the human, the color between two people.
func (m *Model) pgnPlayer(c chess.Color) string {
	switch {
	case !m.vs.computer:
		return colorName(c)
	case c == m.vs.human:
		return "Player"
	}
	return m.computerName() + " (" + levelName(m.vs.elo) + ")"
}

// opponentTags records the computer opponent and the help the human used.
func (m *Model) opponentTags() []chess.Tag {
	if !m.vs.computer {
		return nil
	}
	elo := "max"
	if m.vs.elo > 0 {
		elo = strconv.Itoa(m.vs.elo)
	}
	tags := []chess.Tag{
		{Name: tagOpponent, Value: "computer"},
		{Name: tagElo, Value: elo},
		{Name: tagHuman, Value: strings.ToLower(colorName(m.vs.human))},
	}
	var assist []string
	if m.assistTakeback {
		assist = append(assist, "takebacks")
	}
	if m.assistHint {
		assist = append(assist, "hints")
	}
	if len(assist) > 0 {
		tags = append(tags, chess.Tag{Name: tagAssist, Value: strings.Join(assist, ", ")})
	}
	return tags
}

// recordOpponent reads back the opponent of a game tchess saved. Without
// an engine, a game against the computer goes on between two players.
func (m *Model) recordOpponent(pg *chess.PGNGame) {
	m.vs = opponent{}
	assist := pg.Tag(tagAssist)
	m.assistTakeback = strings.Contains(assist, "takebacks")
	m.assistHint = strings.Contains(assist, "hints")
	if pg.Tag(tagOpponent) != "computer" || pg.Game.IsOver() {
		return
	}
	if !m.engine.available() {
		m.notice = "No engine found: playing two-player"
		return
	}
	m.vs = opponent{computer: true, human: chess.White}
	if pg.Tag(tagHuman) == "black" {
		m.vs.human = chess.Black
	}
	if elo, ok := ParseLevel(pg.Tag(tagElo)); ok {
		m.vs.elo = elo
	}
}

// clockSource is where the clocks of a game opened from PGN come from.
type clockSource uint8

const (
	clocksOff        clockSource = iota // play on without a clock
	clocksFromRecord                    // the TimeControl tag and the [%clk] times
	clocksFresh                         // a new time control, both clocks full
)

// recordClocks reads the record's time control and the mover's clock after
// each move, as far as the [%clk] comments go. ok is false when the record
// has no usable clock.
func recordClocks(pg *chess.PGNGame) (tc clock.TimeControl, clocks []moveClock, remaining [2]time.Duration, ok bool) {
	tc, ok = parsePGNTimeControl(pg.Tag("TimeControl"))
	if !ok || tc.Unlimited() {
		return clock.TimeControl{}, nil, remaining, false
	}
	remaining = [2]time.Duration{tc.Base, tc.Base}
	history := pg.Game.History()
	for i := 0; i < len(history) && i < len(pg.Comments); i++ {
		d, found := commentDuration(pg.Comments[i], "%clk")
		if !found {
			break
		}
		// The record only has the time after each move; assume the usual
		// increment was earned, except on the first move.
		mc := moveClock{before: d, after: d}
		if i > 0 {
			mc.before, mc.running = max(d-tc.Increment, 0), true
		}
		clocks = append(clocks, mc)
		remaining[history[i].Color] = d
	}
	return tc, clocks, remaining, len(clocks) > 0
}

// loadRecord opens a game read from PGN. A finished game opens on its
// final position, ready to be reviewed. An unfinished one is played on
// with the clocks src gives (tc is the time control for clocksFresh); the
// side to move's time starts running at once.
func (m *Model) loadRecord(pg *chess.PGNGame, src clockSource, tc clock.TimeControl) tea.Cmd {
	m.screen = screenGame
	m.game = pg.Game
	m.mode = modePlay
	if pg.Result != chess.Ongoing && !pg.Game.IsOver() {
		// The moves leave the game open; the record says how it ended.
		if pg.Game.Record(pg.Result, terminationFromText(pg.Tag("Termination"))) != nil {
			_ = pg.Game.Record(pg.Result, chess.RecordedResult) // accepts any finished result
		}
	}
	m.tc, m.clock = clock.TimeControl{}, nil
	m.moveClocks, m.clockFrom, m.moveTimes = nil, 0, nil
	m.savedPath = ""
	m.cursorCol, m.cursorRow, m.showCursor = 4, 6, false
	m.notice, m.input, m.inputErr = "", "", ""
	m.anim, m.reviewing, m.help = nil, false, false
	m.lowWarned = [2]bool{}
	m.stopComputer()
	m.engineGame++
	m.resetAnalysis()
	m.recordOpponent(pg)
	m.clearSelection()

	n := pg.Game.Ply()
	for i := 0; i < n && i < len(pg.Comments); i++ {
		d, _ := commentDuration(pg.Comments[i], "%emt")
		m.moveTimes = append(m.moveTimes, d)
	}
	over := pg.Game.IsOver()
	if !over {
		var remaining [2]time.Duration
		switch src {
		case clocksFromRecord:
			if t, clocks, rem, ok := recordClocks(pg); ok {
				m.tc, m.moveClocks, remaining = t, clocks, rem
				if len(clocks) != n { // gaps: keep the times, but not as take-back points
					m.moveClocks, m.clockFrom = nil, n
				}
			}
		case clocksFresh:
			m.tc, m.clockFrom = tc, n // earlier moves have no clock to restore
			remaining = [2]time.Duration{tc.Base, tc.Base}
		}
		if !m.tc.Unlimited() {
			m.clock = clock.New(m.tc, m.now)
			// Before White's first move nothing runs, as in a new game.
			m.clock.Restore(remaining, pg.Game.Turn(), n > 0)
		}
	}
	m.orientation = m.facing()
	m.turnStart = m.now()
	m.tickGen++

	var cmds []tea.Cmd
	m.resultHidden = over // an opened finished game shows its board
	if over {
		m.mode = modeGameOver
	} else {
		m.resume = nil
		cmds = append(cmds, m.files.saveResume(m.gamePGN()))
	}
	if m.clock != nil {
		cmds = append(cmds, tick(m.tickGen))
	}
	return tea.Batch(append(cmds, m.maybeComputer())...)
}

// pgnTimeControl writes a time control as the PGN tag does: "300+3", or
// "-" without a clock.
func pgnTimeControl(tc clock.TimeControl) string {
	if tc.Unlimited() {
		return "-"
	}
	return strconv.FormatFloat(tc.Base.Seconds(), 'f', -1, 64) + "+" +
		strconv.FormatFloat(tc.Increment.Seconds(), 'f', -1, 64)
}

// parsePGNTimeControl reads a TimeControl tag ("300+3", "600", "-").
func parsePGNTimeControl(s string) (clock.TimeControl, bool) {
	if s == "-" {
		return clock.TimeControl{}, true
	}
	base, inc, hasInc := strings.Cut(s, "+")
	b, err := strconv.ParseFloat(base, 64)
	if err != nil || b < 0 {
		return clock.TimeControl{}, false
	}
	tc := clock.TimeControl{Base: time.Duration(b * float64(time.Second))}
	if hasInc {
		i, err := strconv.ParseFloat(inc, 64)
		if err != nil || i < 0 {
			return clock.TimeControl{}, false
		}
		tc.Increment = time.Duration(i * float64(time.Second))
	}
	return tc, true
}

// formatPGNClock writes a duration as h:mm:ss.s.
func formatPGNClock(d time.Duration) string {
	d = max(d, 0).Round(100 * time.Millisecond)
	h := int(d / time.Hour)
	m := int(d/time.Minute) % 60
	s := float64(d%time.Minute) / float64(time.Second)
	return fmt.Sprintf("%d:%02d:%04.1f", h, m, s)
}

// commentDuration finds a [%cmd h:mm:ss(.s)] command in a move comment.
func commentDuration(comment, cmd string) (time.Duration, bool) {
	i := strings.Index(comment, "["+cmd+" ")
	if i < 0 {
		return 0, false
	}
	rest := comment[i+len(cmd)+2:]
	end := strings.IndexByte(rest, ']')
	if end < 0 {
		return 0, false
	}
	fields := strings.Split(strings.TrimSpace(rest[:end]), ":")
	var total float64
	for _, f := range fields {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil || v < 0 {
			return 0, false
		}
		total = total*60 + v
	}
	return time.Duration(total * float64(time.Second)), true
}
