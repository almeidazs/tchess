package chess

import (
	"sort"
)

// Perft counts the leaf nodes of the legal move tree to the given depth. It
// is the standard correctness test for move generators: the counts for
// well-known positions are published and any rules bug changes them.
func (p *Position) Perft(depth int) uint64 {
	if depth == 0 {
		return 1
	}
	var ml MoveList
	p.LegalMoves(&ml)
	if depth == 1 {
		return uint64(ml.n) // bulk counting: no need to make the last ply
	}
	var nodes uint64
	for _, m := range ml.moves[:ml.n] {
		u := p.Make(m)
		nodes += p.Perft(depth - 1)
		p.Unmake(m, u)
	}
	return nodes
}

// PerftEntry is the node count under one root move.
type PerftEntry struct {
	Move  Move
	Nodes uint64
}

// Divide returns the perft count split by root move, sorted by UCI string.
func (p *Position) Divide(depth int) []PerftEntry {
	if depth < 1 {
		return nil
	}
	var ml MoveList
	p.LegalMoves(&ml)
	out := make([]PerftEntry, 0, ml.n)
	for _, m := range ml.Slice() {
		u := p.Make(m)
		out = append(out, PerftEntry{m, p.Perft(depth - 1)})
		p.Unmake(m, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Move.UCI() < out[j].Move.UCI() })
	return out
}
