package chess

import (
	"math/bits"
)

type magic struct {
	mask  Bitboard
	mul   uint64
	shift uint8
	table []Bitboard
}

// attacks looks up the attack set for the given occupancy.
func (m *magic) attacks(occ Bitboard) Bitboard {
	return m.table[(uint64(occ&m.mask)*m.mul)>>m.shift]
}

// initMagics finds a magic multiplier for every square by trial: a candidate
// is accepted when it maps every relevant occupancy to a table slot without
// destructive collisions.
func initMagics(magics *[64]magic, dirs *[4][2]int, rng *prng) {
	var (
		occs, refs [4096]Bitboard
		epoch      [4096]int
		attempt    int
	)
	for sq := A1; sq < NoSquare; sq++ {
		edges := ((rank1 | rank8) &^ (rank1 << (8 * sq.Rank()))) |
			((fileA | fileH) &^ (fileA << sq.File()))
		mask := slidingAttacks(sq, 0, dirs) &^ edges
		nbits := mask.Count()
		shift := uint8(64 - nbits)

		// Enumerate every subset of mask (Carry-Rippler trick).
		n := 0
		for sub := Bitboard(0); ; {
			occs[n], refs[n] = sub, slidingAttacks(sq, sub, dirs)
			n++
			if sub = (sub - mask) & mask; sub == 0 {
				break
			}
		}

		table := make([]Bitboard, 1<<nbits)
		for {
			mul := rng.sparse()
			if bits.OnesCount64((uint64(mask)*mul)>>56) < 6 {
				continue
			}
			attempt++
			ok := true
			for i := range n {
				idx := (uint64(occs[i]) * mul) >> shift
				if epoch[idx] < attempt {
					epoch[idx], table[idx] = attempt, refs[i]
				} else if table[idx] != refs[i] {
					ok = false
					break
				}
			}
			if ok {
				magics[sq] = magic{mask: mask, mul: mul, shift: shift, table: table}
				break
			}
		}
	}
}

// prng is a xorshift64* generator; deterministic so startup is reproducible.
type prng struct{ s uint64 }

// next returns the next pseudo-random number.
func (p *prng) next() uint64 {
	p.s ^= p.s >> 12
	p.s ^= p.s << 25
	p.s ^= p.s >> 27
	return p.s * 2685821657736338717
}

// sparse returns a number with few bits set, which makes good magic candidates.
func (p *prng) sparse() uint64 {
	a, b, c := p.next(), p.next(), p.next()
	return a & b & c
}
