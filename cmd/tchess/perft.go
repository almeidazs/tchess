package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/almeidazs/tchess/chess"
)

// runPerft implements "chess perft <depth> [fen]": node counts per root
// move, total, time and speed.
func runPerft(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("uso: chess perft <profundidade> [fen]")
	}
	depth, err := strconv.Atoi(args[0])
	if err != nil || depth < 1 {
		return fmt.Errorf("profundidade inválida %q", args[0])
	}
	fen := chess.StartFEN
	if len(args) > 1 {
		fen = strings.Join(args[1:], " ")
	}
	pos, err := chess.ParseFEN(fen)
	if err != nil {
		return err
	}

	start := time.Now()
	var total uint64
	for _, e := range pos.Divide(depth) {
		fmt.Printf("%-6s %d\n", e.Move.UCI(), e.Nodes)
		total += e.Nodes
	}
	elapsed := time.Since(start)
	fmt.Printf("\nnós: %d\ntempo: %s\nnós/s: %.0f\n", total, elapsed.Round(time.Millisecond), float64(total)/elapsed.Seconds())
	return nil
}
