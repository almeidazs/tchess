package uci

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Errors returned by Engine.
var (
	// ErrClosed is returned when the engine process is gone.
	ErrClosed = errors.New("uci: engine closed")
	// ErrTimeout is returned when the engine does not answer in time.
	ErrTimeout = errors.New("uci: engine did not answer in time")
)

// handshakeTimeout bounds how long Start waits for "uciok" and "readyok".
const handshakeTimeout = 10 * time.Second

// Option is an option the engine declared ("option name ... type ...").
type Option struct {
	Name     string
	Type     string // check, spin, combo, button or string
	Default  string
	Min, Max int // spin only
}

// Engine is a running UCI engine. Its methods must not be called
// concurrently; Search can be canceled through its context.
type Engine struct {
	// Name is what the engine calls itself ("id name"), or the program's
	// file name when it does not say.
	Name    string
	Options map[string]Option // by lowercase name

	cmd   *exec.Cmd
	in    io.WriteCloser
	lines chan string // stdout, line by line; closed when the process ends

	closeOnce sync.Once
}

// Start runs the engine at path and completes the UCI handshake.
func Start(path string, args ...string) (*Engine, error) {
	cmd := exec.Command(path, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("uci: %w", err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("uci: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("uci: %w", err)
	}
	e := &Engine{Options: map[string]Option{}, cmd: cmd, in: in, lines: make(chan string, 256)}
	go func() {
		sc := bufio.NewScanner(out)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			e.lines <- strings.TrimSpace(sc.Text())
		}
		close(e.lines)
	}()

	if err := e.send("uci"); err != nil {
		e.Close()
		return nil, err
	}
	deadline := time.After(handshakeTimeout)
	for done := false; !done; {
		select {
		case line, ok := <-e.lines:
			if !ok {
				e.Close()
				return nil, fmt.Errorf("uci: %s exited during the handshake", path)
			}
			switch {
			case line == "uciok":
				done = true
			case strings.HasPrefix(line, "id name "):
				e.Name = strings.TrimPrefix(line, "id name ")
			case strings.HasPrefix(line, "option "):
				if o, ok := parseOption(line); ok {
					e.Options[strings.ToLower(o.Name)] = o
				}
			}
		case <-deadline:
			e.Close()
			return nil, fmt.Errorf("uci: %s is not a UCI engine: %w", path, ErrTimeout)
		}
	}
	if e.Name == "" {
		e.Name = baseName(path)
	}
	if err := e.sync(); err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}

// baseName is the file name of path without its extension.
func baseName(path string) string {
	name := path[strings.LastIndexAny(path, `/\`)+1:]
	if i := strings.LastIndexByte(name, '.'); i > 0 {
		name = name[:i]
	}
	return name
}

// parseOption reads an "option name <n> type <t> [default <d>] [min <a>]
// [max <b>] [var <v>]..." line. Names may contain spaces.
func parseOption(line string) (Option, bool) {
	keywords := map[string]bool{"name": true, "type": true, "default": true, "min": true, "max": true, "var": true}
	fields := strings.Fields(line)[1:]
	values := map[string][]string{}
	key := ""
	for _, f := range fields {
		if keywords[f] {
			key = f
			if _, seen := values[key]; !seen {
				values[key] = nil
			}
			continue
		}
		if key != "" {
			values[key] = append(values[key], f)
		}
	}
	o := Option{
		Name:    strings.Join(values["name"], " "),
		Type:    strings.Join(values["type"], " "),
		Default: strings.Join(values["default"], " "),
	}
	o.Min, _ = strconv.Atoi(strings.Join(values["min"], ""))
	o.Max, _ = strconv.Atoi(strings.Join(values["max"], ""))
	return o, o.Name != ""
}

// Option returns the named option, ignoring case.
func (e *Engine) Option(name string) (Option, bool) {
	o, ok := e.Options[strings.ToLower(name)]
	return o, ok
}

// SetOption sets an option. Unknown options are ignored by engines, so the
// caller should check Option first when it matters.
func (e *Engine) SetOption(name, value string) error {
	return e.send("setoption name " + name + " value " + value)
}

// NewGame tells the engine a new game starts and waits until it is ready.
func (e *Engine) NewGame() error {
	if err := e.send("ucinewgame"); err != nil {
		return err
	}
	return e.sync()
}

// sync sends "isready" and waits for "readyok".
func (e *Engine) sync() error {
	if err := e.send("isready"); err != nil {
		return err
	}
	deadline := time.After(handshakeTimeout)
	for {
		select {
		case line, ok := <-e.lines:
			if !ok {
				return ErrClosed
			}
			if line == "readyok" {
				return nil
			}
		case <-deadline:
			return ErrTimeout
		}
	}
}

// send writes one command.
func (e *Engine) send(cmd string) error {
	if _, err := io.WriteString(e.in, cmd+"\n"); err != nil {
		return fmt.Errorf("%w: %w", ErrClosed, err)
	}
	return nil
}

// Limits bounds a search. Zero fields are left out; with none set the
// engine searches until the context is canceled.
type Limits struct {
	WTime, BTime time.Duration // clocks
	WInc, BInc   time.Duration // increments
	MoveTime     time.Duration // exact time for this move
	Depth        int
}

// Score is an evaluation from the point of view of the side to move.
type Score struct {
	CP   int  // centipawns, when Mate is 0
	Mate int  // moves to mate: positive when the side to move mates
	Set  bool // the engine reported a score
}

// Result is what a search found.
type Result struct {
	BestMove string // coordinate notation, e.g. "e2e4"; "" or "(none)" without moves
	Score    Score
	Depth    int
	PV       []string
}

// Search sets the position (a FEN plus the moves played from it, in
// coordinate notation), searches within limits and returns the engine's
// choice. Canceling ctx stops the search early; the engine still answers
// with its best move so far.
func (e *Engine) Search(ctx context.Context, fen string, moves []string, l Limits) (Result, error) {
	pos := "position fen " + fen
	if len(moves) > 0 {
		pos += " moves " + strings.Join(moves, " ")
	}
	if err := e.send(pos); err != nil {
		return Result{}, err
	}
	goCmd := "go"
	add := func(name string, d time.Duration) {
		if d > 0 {
			goCmd += fmt.Sprintf(" %s %d", name, d.Milliseconds())
		}
	}
	add("wtime", l.WTime)
	add("btime", l.BTime)
	add("winc", l.WInc)
	add("binc", l.BInc)
	add("movetime", l.MoveTime)
	if l.Depth > 0 {
		goCmd += " depth " + strconv.Itoa(l.Depth)
	}
	if goCmd == "go" {
		goCmd = "go infinite"
	}
	if err := e.send(goCmd); err != nil {
		return Result{}, err
	}

	var res Result
	stopped := false
	done := ctx.Done()
	for {
		select {
		case line, ok := <-e.lines:
			if !ok {
				return res, ErrClosed
			}
			switch {
			case strings.HasPrefix(line, "info "):
				parseInfo(line, &res)
			case strings.HasPrefix(line, "bestmove"):
				if f := strings.Fields(line); len(f) > 1 {
					res.BestMove = f[1]
				}
				return res, ctx.Err()
			}
		case <-done:
			if !stopped {
				stopped = true
				done = nil // wait for bestmove now
				if err := e.send("stop"); err != nil {
					return res, err
				}
			}
		}
	}
}

// parseInfo folds an "info" line into res: depth, score and principal
// variation. Lines about the current move only (no score) are skipped.
func parseInfo(line string, res *Result) {
	f := strings.Fields(line)
	var depth int
	var score Score
	var pv []string
	for i := 1; i < len(f); i++ {
		switch f[i] {
		case "depth":
			if i+1 < len(f) {
				depth, _ = strconv.Atoi(f[i+1])
				i++
			}
		case "score":
			if i+2 < len(f) {
				v, err := strconv.Atoi(f[i+2])
				if err == nil {
					switch f[i+1] {
					case "cp":
						score = Score{CP: v, Set: true}
					case "mate":
						score = Score{Mate: v, Set: true}
					}
				}
				i += 2
			}
		case "pv":
			pv = append([]string(nil), f[i+1:]...)
			i = len(f)
		}
	}
	if !score.Set {
		return
	}
	// Bounds ("lowerbound"/"upperbound") are close enough for display.
	res.Score, res.Depth = score, depth
	if len(pv) > 0 {
		res.PV = pv
	}
}

// Close asks the engine to quit and makes sure the process ends.
func (e *Engine) Close() error {
	e.closeOnce.Do(func() {
		_ = e.send("quit")
		_ = e.in.Close()
		done := make(chan struct{})
		go func() {
			_ = e.cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = e.cmd.Process.Kill()
			<-done
		}
		for line := range e.lines { // drain so the reader goroutine ends
			_ = line
		}
	})
	return nil
}
