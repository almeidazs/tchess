package chess

import (
	"math/bits"
	"strings"
)

// Bitboard is a set of squares, one bit per square (bit 0 = A1).
type Bitboard uint64

const (
	fileA Bitboard = 0x0101010101010101
	fileH Bitboard = fileA << 7
	rank1 Bitboard = 0xFF
	rank8 Bitboard = rank1 << 56

	lightSquares Bitboard = 0x55AA55AA55AA55AA
	darkSquares           = ^lightSquares
)

// Has reports whether s is in the set.
func (b Bitboard) Has(s Square) bool { return b&(1<<s) != 0 }

// Count returns the number of squares in the set.
func (b Bitboard) Count() int { return bits.OnesCount64(uint64(b)) }

// LSB returns the lowest square in the set. Undefined when empty.
func (b Bitboard) LSB() Square { return Square(bits.TrailingZeros64(uint64(b))) }

// PopLSB removes and returns the lowest square in the set.
func (b *Bitboard) PopLSB() Square {
	s := b.LSB()
	*b &= *b - 1
	return s
}

func (b Bitboard) String() string {
	var sb strings.Builder
	for r := 7; r >= 0; r-- {
		for f := range 8 {
			if b.Has(NewSquare(f, r)) {
				sb.WriteByte('1')
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
