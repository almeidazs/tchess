package chess

// Move is a move packed in 16 bits: from (6) | to (6) | flag (4).
type Move uint16

// MoveFlag classifies a move. Bit 2 marks captures, bit 3 promotions and the
// low two bits of a promotion select the piece.
type MoveFlag uint8

const (
	FlagQuiet       MoveFlag = 0
	FlagDoublePush  MoveFlag = 1
	FlagKingCastle  MoveFlag = 2
	FlagQueenCastle MoveFlag = 3
	FlagCapture     MoveFlag = 4
	FlagEnPassant   MoveFlag = 5

	FlagPromoKnight MoveFlag = 8
	FlagPromoBishop MoveFlag = 9
	FlagPromoRook   MoveFlag = 10
	FlagPromoQueen  MoveFlag = 11

	FlagPromoCaptureKnight MoveFlag = 12
	FlagPromoCaptureBishop MoveFlag = 13
	FlagPromoCaptureRook   MoveFlag = 14
	FlagPromoCaptureQueen  MoveFlag = 15
)

// NullMove is the zero move; it is never legal.
const NullMove Move = 0

// NewMove packs a move.
func NewMove(from, to Square, flag MoveFlag) Move {
	return Move(from) | Move(to)<<6 | Move(flag)<<12
}

// From returns the origin square.
func (m Move) From() Square { return Square(m & 0x3F) }

// To returns the destination square.
func (m Move) To() Square { return Square(m >> 6 & 0x3F) }

// Flag returns the move kind.
func (m Move) Flag() MoveFlag { return MoveFlag(m >> 12) }

// IsCapture reports captures, including en passant and capturing promotions.
func (m Move) IsCapture() bool { return m.Flag()&4 != 0 }

// IsPromotion reports pawn promotions.
func (m Move) IsPromotion() bool { return m.Flag()&8 != 0 }

// IsEnPassant reports en passant captures.
func (m Move) IsEnPassant() bool { return m.Flag() == FlagEnPassant }

// IsCastle reports castling, on either side.
func (m Move) IsCastle() bool {
	f := m.Flag()
	return f == FlagKingCastle || f == FlagQueenCastle
}

// Promotion returns the promoted piece type, or NoPieceType.
func (m Move) Promotion() PieceType {
	if !m.IsPromotion() {
		return NoPieceType
	}
	return Knight + PieceType(m.Flag()&3)
}

// UCI returns the move in coordinate notation, e.g. "e2e4" or "e7e8q".
func (m Move) UCI() string {
	s := m.From().String() + m.To().String()
	if m.IsPromotion() {
		s += string(m.Promotion().Letter() + ('a' - 'A'))
	}
	return s
}

func (m Move) String() string { return m.UCI() }
