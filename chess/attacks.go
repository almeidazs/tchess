package chess

// Precomputed attack tables. Leapers (knight, king, pawn) use plain lookup
// tables; sliders (bishop, rook, queen) use magic bitboards, which turn
// "attacks from square s given occupancy occ" into a multiply, a shift and a
// table lookup.
var (
	knightAttacks [64]Bitboard
	kingAttacks   [64]Bitboard
	pawnAttacks   [2][64]Bitboard

	// between[a][b] holds the squares strictly between a and b when they
	// share a rank, file or diagonal; empty otherwise.
	between [64][64]Bitboard
	// line[a][b] holds the full line through a and b (edge to edge) when
	// they are aligned; empty otherwise.
	line [64][64]Bitboard

	rookMagics   [64]magic
	bishopMagics [64]magic
)

// RookAttacks returns the squares a rook on sq attacks given occupancy occ.
func RookAttacks(sq Square, occ Bitboard) Bitboard { return rookMagics[sq].attacks(occ) }

// BishopAttacks returns the squares a bishop on sq attacks given occupancy occ.
func BishopAttacks(sq Square, occ Bitboard) Bitboard { return bishopMagics[sq].attacks(occ) }

// QueenAttacks returns the squares a queen on sq attacks given occupancy occ.
func QueenAttacks(sq Square, occ Bitboard) Bitboard {
	return RookAttacks(sq, occ) | BishopAttacks(sq, occ)
}

var (
	rookDirs   = [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	bishopDirs = [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
)

func init() {
	initLeapers()
	rng := prng{s: 0x9E3779B97F4A7C15}
	initMagics(&rookMagics, &rookDirs, &rng)
	initMagics(&bishopMagics, &bishopDirs, &rng)
	initLines()
}

// initLeapers fills the knight, king and pawn attack tables.
func initLeapers() {
	knight := [8][2]int{{1, 2}, {2, 1}, {2, -1}, {1, -2}, {-1, -2}, {-2, -1}, {-2, 1}, {-1, 2}}
	king := [8][2]int{{1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}}
	for sq := A1; sq < NoSquare; sq++ {
		f, r := sq.File(), sq.Rank()
		for i := range 8 {
			if nf, nr := f+knight[i][0], r+knight[i][1]; onBoard(nf, nr) {
				knightAttacks[sq] |= NewSquare(nf, nr).Bitboard()
			}
			if nf, nr := f+king[i][0], r+king[i][1]; onBoard(nf, nr) {
				kingAttacks[sq] |= NewSquare(nf, nr).Bitboard()
			}
		}
		b := sq.Bitboard()
		pawnAttacks[White][sq] = (b<<7)&^fileH | (b<<9)&^fileA
		pawnAttacks[Black][sq] = (b>>9)&^fileH | (b>>7)&^fileA
	}
}

// onBoard reports whether file f and rank r are on the board.
func onBoard(f, r int) bool { return f >= 0 && f < 8 && r >= 0 && r < 8 }

// slidingAttacks computes slider attacks the slow way, by walking rays. It is
// only used to fill the magic tables at startup.
func slidingAttacks(sq Square, occ Bitboard, dirs *[4][2]int) Bitboard {
	var att Bitboard
	for _, d := range dirs {
		f, r := sq.File()+d[0], sq.Rank()+d[1]
		for onBoard(f, r) {
			s := NewSquare(f, r)
			att |= s.Bitboard()
			if occ.Has(s) {
				break
			}
			f, r = f+d[0], r+d[1]
		}
	}
	return att
}

// initLines fills the between and line tables from the slider attacks.
func initLines() {
	for a := A1; a < NoSquare; a++ {
		for b := A1; b < NoSquare; b++ {
			if a == b {
				continue
			}
			ab := a.Bitboard() | b.Bitboard()
			switch {
			case BishopAttacks(a, 0).Has(b):
				between[a][b] = BishopAttacks(a, b.Bitboard()) & BishopAttacks(b, a.Bitboard())
				line[a][b] = BishopAttacks(a, 0)&BishopAttacks(b, 0) | ab
			case RookAttacks(a, 0).Has(b):
				between[a][b] = RookAttacks(a, b.Bitboard()) & RookAttacks(b, a.Bitboard())
				line[a][b] = RookAttacks(a, 0)&RookAttacks(b, 0) | ab
			}
		}
	}
}
