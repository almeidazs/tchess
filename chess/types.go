package chess

// Color identifies a side.
type Color uint8

const (
	White Color = iota
	Black
)

// Other returns the opposite color.
func (c Color) Other() Color { return c ^ 1 }

func (c Color) String() string {
	if c == White {
		return "white"
	}
	return "black"
}

// PieceType is a kind of piece, independent of color.
type PieceType uint8

const (
	Pawn PieceType = iota
	Knight
	Bishop
	Rook
	Queen
	King
	NoPieceType
)

var pieceTypeLetters = [...]byte{'P', 'N', 'B', 'R', 'Q', 'K'}

// Letter returns the uppercase SAN letter of the piece type ('P' for pawns).
func (t PieceType) Letter() byte { return pieceTypeLetters[t] }

// Piece is a colored piece. NoPiece marks an empty square.
type Piece uint8

// NoPiece is the value of an empty square.
const NoPiece Piece = 12

const fenPieces = "PNBRQKpnbrqk"

// MakePiece builds a piece from a color and a type.
func MakePiece(c Color, t PieceType) Piece { return Piece(c)*6 + Piece(t) }

// Type returns the piece type. Undefined for NoPiece.
func (p Piece) Type() PieceType { return PieceType(p % 6) }

// Color returns the piece color. Undefined for NoPiece.
func (p Piece) Color() Color { return Color(p / 6) }

func (p Piece) String() string {
	if p >= NoPiece {
		return "."
	}
	return string(fenPieces[p])
}
