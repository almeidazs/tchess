package tui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/uci"
)

// The engine runs in the background, one search at a time. Scheduling a
// search cancels the one in progress, and every result carries the number
// of the search that produced it, so results of canceled searches are
// recognized and dropped.

// jobKind is what a search is for.
type jobKind uint8

const (
	jobMove jobKind = iota // the computer's move
	jobHint                // a hint for the human
	jobDraw                // deciding on a draw offer
	jobEval                // evaluating a reviewed position
)

// Thinking times of the searches that are not the computer's own moves.
const (
	hintTime = time.Second
	drawTime = 300 * time.Millisecond
	evalTime = 300 * time.Millisecond
)

// searchJob describes one search.
type searchJob struct {
	kind  jobKind
	ply   int      // plies played in the searched position
	fen   string   // the game's starting position
	moves []string // moves from it, in coordinate notation
	game  int      // game number: a new one sends "ucinewgame"

	elo      int        // jobMove: strength (0: full); other jobs: always full
	clock    uci.Limits // jobMove: the clocks, when the game has them
	clocked  bool
	moveTime time.Duration // other jobs: how long to think
}

// engineMsg is the outcome of a search.
type engineMsg struct {
	job  int
	kind jobKind
	ply  int
	res  uci.Result
	name string // the engine's name, once known
	err  error
}

// engineCtl owns the engine process.
type engineCtl struct {
	path string // "" when no engine was found

	mu       sync.Mutex // held while a search uses the engine
	eng      *uci.Engine
	applied  string // strength options last sent
	lastGame int

	job    int // number of the latest search (update loop only)
	cancel context.CancelFunc
}

// available reports whether there is an engine to run.
func (c *engineCtl) available() bool { return c.path != "" }

// displayName guesses a name from the program before the engine says its
// own: "stockfish" and "stockfish-linux-x86-64" become "Stockfish".
func (c *engineCtl) displayName() string {
	name := filepath.Base(c.path)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if i := strings.IndexAny(name, "-_"); i > 0 {
		name = name[:i]
	}
	if name == "" || name == "." {
		return "Computer"
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// stop cancels the search in progress, if any.
func (c *engineCtl) stop() {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}

// run schedules a search and returns the command that performs it.
func (c *engineCtl) run(j searchJob) tea.Cmd {
	c.stop()
	c.job++
	id := c.job
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	return func() tea.Msg {
		defer cancel()
		msg := engineMsg{job: id, kind: j.kind, ply: j.ply}
		msg.res, msg.name, msg.err = c.search(ctx, j)
		return msg
	}
}

// search runs j on the engine, starting it when needed.
func (c *engineCtl) search(ctx context.Context, j searchJob) (uci.Result, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ctx.Err() != nil {
		return uci.Result{}, "", ctx.Err()
	}
	if c.eng == nil {
		e, err := uci.Start(c.path)
		if err != nil {
			return uci.Result{}, "", err
		}
		c.eng, c.applied, c.lastGame = e, "", 0
	}
	res, err := c.searchOn(ctx, j)
	if errors.Is(err, uci.ErrClosed) { // the engine died: start a fresh one next time
		c.eng.Close()
		c.eng = nil
	}
	name := ""
	if c.eng != nil {
		name = c.eng.Name
	}
	return res, name, err
}

// searchOn configures the engine for j and searches.
func (c *engineCtl) searchOn(ctx context.Context, j searchJob) (uci.Result, error) {
	elo := 0
	if j.kind == jobMove {
		elo = j.elo
	}
	st := strengthFor(c.eng.Options, elo)
	if k := st.key(); k != c.applied {
		for _, o := range st.options {
			if err := c.eng.SetOption(o[0], o[1]); err != nil {
				return uci.Result{}, err
			}
		}
		c.applied = k
	}
	if j.game != c.lastGame {
		if err := c.eng.NewGame(); err != nil {
			return uci.Result{}, err
		}
		c.lastGame = j.game
	}

	limits := uci.Limits{MoveTime: j.moveTime}
	if j.kind == jobMove {
		limits = uci.Limits{Depth: st.depth}
		switch {
		case j.clocked:
			limits.WTime, limits.BTime = j.clock.WTime, j.clock.BTime
			limits.WInc, limits.BInc = j.clock.WInc, j.clock.BInc
			if st.think > 0 { // weaker levels answer within their thinking time
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, st.think)
				defer cancel()
			}
		case st.think > 0:
			limits.MoveTime = st.think
		default:
			limits.MoveTime = fullStrengthMoveTime
		}
	}
	res, err := c.eng.Search(ctx, j.fen, j.moves, limits)
	if errors.Is(err, context.DeadlineExceeded) {
		err = nil // thinking time used up: the best move so far is the answer
	}
	return res, err
}

// close stops any search and ends the engine process.
func (c *engineCtl) close() {
	c.stop()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.eng != nil {
		c.eng.Close()
		c.eng = nil
	}
}
