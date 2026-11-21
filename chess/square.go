package chess

import (
	"fmt"
)

// Square is a board square, A1 = 0 ... H8 = 63.
type Square uint8

const (
	A1 Square = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
	NoSquare
)

// NewSquare returns the square at file (0-7, a-h) and rank (0-7, 1-8).
func NewSquare(file, rank int) Square { return Square(rank*8 + file) }

// ParseSquare parses algebraic coordinates such as "e4".
func ParseSquare(s string) (Square, error) {
	if len(s) != 2 || s[0] < 'a' || s[0] > 'h' || s[1] < '1' || s[1] > '8' {
		return NoSquare, fmt.Errorf("chess: invalid square %q", s)
	}
	return NewSquare(int(s[0]-'a'), int(s[1]-'1')), nil
}

// File returns 0-7 for files a-h.
func (s Square) File() int { return int(s & 7) }

// Rank returns 0-7 for ranks 1-8.
func (s Square) Rank() int { return int(s >> 3) }

// Bitboard returns a bitboard with only this square set.
func (s Square) Bitboard() Bitboard { return 1 << s }

// IsLight reports whether the square is a light square.
func (s Square) IsLight() bool { return (s.File()+s.Rank())%2 == 1 }

func (s Square) String() string {
	if s >= NoSquare {
		return "-"
	}
	return string([]byte{byte('a' + s.File()), byte('1' + s.Rank())})
}
