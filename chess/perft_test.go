package chess

import (
	"testing"
)

// Reference counts from https://www.chessprogramming.org/Perft_Results.
var perftSuite = []struct {
	name   string
	fen    string
	counts []uint64 // counts[i] = perft(i+1)
	short  int      // depth used with -short
}{
	{"initial", StartFEN, []uint64{20, 400, 8902, 197281, 4865609}, 4},
	{"kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		[]uint64{48, 2039, 97862, 4085603}, 3},
	{"position3", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		[]uint64{14, 191, 2812, 43238, 674624}, 4},
	{"position4", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		[]uint64{6, 264, 9467, 422333}, 3},
	{"position4-mirrored", "r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1",
		[]uint64{6, 264, 9467, 422333}, 3},
	{"position5", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		[]uint64{44, 1486, 62379, 2103487}, 3},
	{"position6", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		[]uint64{46, 2079, 89890, 3894594}, 3},
}

func TestPerft(t *testing.T) {
	for _, tc := range perftSuite {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, err := ParseFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			before := *p
			maxDepth := len(tc.counts)
			if testing.Short() {
				maxDepth = tc.short
			}
			for d := 1; d <= maxDepth; d++ {
				if got := p.Perft(d); got != tc.counts[d-1] {
					t.Fatalf("perft(%d) = %d, want %d", d, got, tc.counts[d-1])
				}
			}
			if *p != before {
				t.Fatal("perft did not restore the position (make/unmake mismatch)")
			}
		})
	}
}

// TestMakeUnmakeHash walks the tree and checks that the incremental hash
// always equals a full recomputation and that unmake restores everything.
func TestMakeUnmakeHash(t *testing.T) {
	for _, tc := range perftSuite {
		p, err := ParseFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		var walk func(depth int)
		walk = func(depth int) {
			if p.hash != p.computeHash() {
				t.Fatalf("%s: incremental hash diverged at %s", tc.name, p.FEN())
			}
			if depth == 0 {
				return
			}
			var ml MoveList
			p.LegalMoves(&ml)
			for _, m := range ml.Slice() {
				before := *p
				u := p.Make(m)
				walk(depth - 1)
				p.Unmake(m, u)
				if *p != before {
					t.Fatalf("%s: unmake %s did not restore %s", tc.name, m, before.FEN())
				}
			}
		}
		walk(3)
	}
}

func TestDivideSumsToPerft(t *testing.T) {
	p := StartingPosition()
	var sum uint64
	for _, e := range p.Divide(3) {
		sum += e.Nodes
	}
	if sum != 8902 {
		t.Fatalf("divide(3) sums to %d, want 8902", sum)
	}
}
