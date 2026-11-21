package chess

// CastlingRights is a set of the four castling permissions.
type CastlingRights uint8

const (
	WhiteKingside CastlingRights = 1 << iota
	WhiteQueenside
	BlackKingside
	BlackQueenside

	NoCastling  CastlingRights = 0
	AllCastling                = WhiteKingside | WhiteQueenside | BlackKingside | BlackQueenside
)

// castleClear[sq] lists the rights lost when a move starts or ends on sq.
var castleClear = func() (t [64]CastlingRights) {
	t[E1], t[A1], t[H1] = WhiteKingside|WhiteQueenside, WhiteQueenside, WhiteKingside
	t[E8], t[A8], t[H8] = BlackKingside|BlackQueenside, BlackQueenside, BlackKingside
	return t
}()

// castleRookSquares returns the rook's from/to squares for a castle whose king
// lands on kingTo.
func castleRookSquares(kingTo Square) (from, to Square) {
	switch kingTo {
	case G1:
		return H1, F1
	case C1:
		return A1, D1
	case G8:
		return H8, F8
	default: // C8
		return A8, D8
	}
}
