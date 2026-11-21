// Package ucitest provides a small fake UCI engine for tests: it answers
// the handshake, declares Stockfish-like strength options and always plays
// the first legal move, reporting a configurable score. Tests run it by
// re-executing their own binary (see Main).
package ucitest

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/almeidazs/tchess/chess"
)

// EnvVar, when set in the environment, makes Main serve as the engine.
const EnvVar = "TCHESS_FAKE_UCI"

// LogVar names a file where the engine appends every command it receives.
const LogVar = "TCHESS_FAKE_UCI_LOG"

// ScoreVar sets the score the engine reports, in centipawns (default 25).
const ScoreVar = "TCHESS_FAKE_UCI_SCORE"

// Main is meant for TestMain: in the child process started by a test it
// serves the fake engine on stdin/stdout and exits; otherwise it sets
// EnvVar so children become engines, and returns.
func Main() {
	if os.Getenv(EnvVar) != "" {
		serveFromEnv()
		os.Exit(0)
	}
	os.Setenv(EnvVar, "1")
}

// serveFromEnv serves the engine as the environment configures it.
func serveFromEnv() {
	log := io.Discard
	if path := os.Getenv(LogVar); path != "" {
		if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
			defer f.Close()
			log = f
		}
	}
	score := 25
	if v, err := strconv.Atoi(os.Getenv(ScoreVar)); err == nil {
		score = v
	}
	Serve(os.Stdin, os.Stdout, log, score)
}

// Serve runs the fake engine until "quit" or the end of input.
func Serve(in io.Reader, out io.Writer, log io.Writer, score int) {
	w := bufio.NewWriter(out)
	say := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
		w.Flush()
	}
	var pos *chess.Position
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		fmt.Fprintln(log, line)
		switch {
		case line == "uci":
			say("id name FakeFish 1")
			say("id author tchess tests")
			say("option name Hash type spin default 16 min 1 max 1024")
			say("option name Skill Level type spin default 20 min 0 max 20")
			say("option name UCI_LimitStrength type check default false")
			say("option name UCI_Elo type spin default 1320 min 1320 max 3190")
			say("uciok")
		case line == "isready":
			say("readyok")
		case strings.HasPrefix(line, "position "):
			pos = parsePosition(line)
		case strings.HasPrefix(line, "go"):
			best := "(none)"
			if pos != nil {
				var ml chess.MoveList
				pos.LegalMoves(&ml)
				if moves := ml.Slice(); len(moves) > 0 {
					best = moves[0].UCI()
				}
			}
			say("info depth 1 score cp %d nodes 20 pv %s", score, best)
			say("bestmove %s", best)
		case line == "quit":
			return
		}
	}
}

// parsePosition reads "position fen <fen> [moves ...]" or "position
// startpos [moves ...]".
func parsePosition(line string) *chess.Position {
	rest := strings.TrimPrefix(line, "position ")
	fen := chess.StartFEN
	if strings.HasPrefix(rest, "fen ") {
		rest = strings.TrimPrefix(rest, "fen ")
		fen, rest, _ = strings.Cut(rest, " moves ")
		rest = " moves " + rest
	}
	pos, err := chess.ParseFEN(strings.TrimSpace(fen))
	if err != nil {
		return nil
	}
	if _, moves, ok := strings.Cut(rest, "moves "); ok {
		for _, mv := range strings.Fields(moves) {
			m, err := pos.ParseUCI(mv)
			if err != nil {
				return nil
			}
			pos.Make(m)
		}
	}
	return pos
}
