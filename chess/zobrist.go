package chess

// Zobrist keys for incremental position hashing, used to detect repetitions.
var zobrist struct {
	piece    [12][64]uint64
	castling [16]uint64
	epFile   [8]uint64
	side     uint64
}

func init() {
	s := uint64(0x2545F4914F6CDD1D)
	next := func() uint64 { // splitmix64
		s += 0x9E3779B97F4A7C15
		z := s
		z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
		z = (z ^ (z >> 27)) * 0x94D049BB133111EB
		return z ^ (z >> 31)
	}
	for p := range zobrist.piece {
		for sq := range zobrist.piece[p] {
			zobrist.piece[p][sq] = next()
		}
	}
	for i := range zobrist.castling {
		zobrist.castling[i] = next()
	}
	for i := range zobrist.epFile {
		zobrist.epFile[i] = next()
	}
	zobrist.side = next()
}
