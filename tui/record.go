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
		{Name: "Event", Value: "Partida casual"},
		{Name: "Site", Value: "tchess"},
		{Name: "Date", Value: m.now().Format("2006.01.02")},
		{Name: "Round", Value: "-"},
		{Name: "White", Value: colorName(chess.White)},
		{Name: "Black", Value: colorName(chess.Black)},
		{Name: "Result", Value: "*"},
		{Name: "TimeControl", Value: pgnTimeControl(m.tc)},
	}
	if t := m.game.Termination(); t != chess.NoTermination {
		tags = append(tags, chess.Tag{Name: "Termination", Value: terminationText(t)})
	}
	comments := make([]string, m.game.Ply())
	for i := range comments {
		var parts []string
		if i < len(m.moveClocks) {
			parts = append(parts, "[%clk "+formatPGNClock(m.moveClocks[i].after)+"]")
		}
		if i < len(m.moveTimes) {
			parts = append(parts, "[%emt "+formatPGNClock(m.moveTimes[i])+"]")
		}
		comments[i] = strings.Join(parts, " ")
	}
	return m.game.PGN(tags, comments)
}

// loadRecord opens a game read from PGN. A finished game opens on its
// result; an unfinished one can be played on. withClock restores the time
// control and both clocks from the record (resuming); otherwise the game
// continues without a clock (importing). Either way the clock stays
// stopped until the next move.
func (m *Model) loadRecord(pg *chess.PGNGame, withClock bool) tea.Cmd {
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
	m.moveClocks, m.moveTimes = nil, nil
	m.savedPath = ""
	m.cursorCol, m.cursorRow, m.showCursor = 4, 6, false
	m.notice, m.input, m.inputErr = "", "", ""
	m.anim, m.reviewing, m.help = nil, false, false
	m.lowWarned = [2]bool{}
	m.clearSelection()

	n := pg.Game.Ply()
	for i := 0; i < n && i < len(pg.Comments); i++ {
		d, _ := commentDuration(pg.Comments[i], "%emt")
		m.moveTimes = append(m.moveTimes, d)
	}
	if tc, ok := parsePGNTimeControl(pg.Tag("TimeControl")); ok && withClock && !tc.Unlimited() {
		m.tc = tc
		m.clock = clock.New(tc, m.now)
		remaining := [2]time.Duration{tc.Base, tc.Base}
		for i := 0; i < n && i < len(pg.Comments); i++ {
			d, ok := commentDuration(pg.Comments[i], "%clk")
			if !ok {
				break
			}
			// The record only has the time after each move; assume the
			// usual increment was earned, except on the first move.
			pr := moveClock{before: d, after: d}
			if i > 0 {
				pr.before, pr.running = max(d-tc.Increment, 0), true
			}
			m.moveClocks = append(m.moveClocks, pr)
			remaining[pg.Game.History()[i].Color] = d
		}
		m.clock.Restore(remaining, pg.Game.Turn(), false) // runs from the next move
	}
	m.orientation = m.facing()
	m.turnStart = m.now()
	m.tickGen++

	var cmds []tea.Cmd
	m.resultHidden = false
	if pg.Game.IsOver() {
		m.mode = modeGameOver
		m.resultHidden = true // the player opened it to look at the board
	} else {
		m.resume = nil
		cmds = append(cmds, m.files.saveResume(m.gamePGN()))
	}
	if m.clock != nil && !pg.Game.IsOver() {
		cmds = append(cmds, tick(m.tickGen))
	}
	return tea.Batch(cmds...)
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
