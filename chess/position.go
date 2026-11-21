package chess

// Position is a chess position. Pieces are stored both as bitboards (for move
// generation) and as a mailbox (for O(1) "what is on this square").
//
// The zero value is not a valid position; use ParseFEN or StartingPosition.
// Position is a plain value: copying it yields an independent position.
type Position struct {
	pieces   [2][6]Bitboard
	occ      [2]Bitboard
	all      Bitboard
	board    [64]Piece
	side     Color
	castling CastlingRights
	ep       Square // only set when an en passant capture is pseudo-legal
	halfmove uint16
	fullmove uint16
	hash     uint64
}

// StartingPosition returns the standard initial position.
func StartingPosition() *Position {
	p, err := ParseFEN(StartFEN)
	if err != nil {
		panic(err)
	}
	return p
}

// SideToMove returns the color to move.
func (p *Position) SideToMove() Color { return p.side }

// PieceAt returns the piece on s, or NoPiece.
func (p *Position) PieceAt(s Square) Piece { return p.board[s] }

// Castling returns the castling rights still available.
func (p *Position) Castling() CastlingRights { return p.castling }

// EnPassant returns the en passant target square, or NoSquare.
func (p *Position) EnPassant() Square { return p.ep }

// HalfmoveClock returns the plies since the last capture or pawn move.
func (p *Position) HalfmoveClock() int { return int(p.halfmove) }

// FullmoveNumber returns the move number, starting at 1.
func (p *Position) FullmoveNumber() int { return int(p.fullmove) }

// Hash returns the Zobrist hash of the position.
func (p *Position) Hash() uint64 { return p.hash }

// Occupied returns the squares occupied by c.
func (p *Position) Occupied(c Color) Bitboard { return p.occ[c] }

// Pieces returns the squares holding c's pieces of type t.
func (p *Position) Pieces(c Color, t PieceType) Bitboard { return p.pieces[c][t] }

// KingSquare returns the square of c's king.
func (p *Position) KingSquare(c Color) Square { return p.pieces[c][King].LSB() }

// InCheck reports whether the side to move is in check.
func (p *Position) InCheck() bool {
	return p.attackersTo(p.KingSquare(p.side), p.all)&p.occ[p.side.Other()] != 0
}

// attackersTo returns every piece of either color attacking sq given occ.
func (p *Position) attackersTo(sq Square, occ Bitboard) Bitboard {
	w, b := &p.pieces[White], &p.pieces[Black]
	diag := w[Bishop] | w[Queen] | b[Bishop] | b[Queen]
	orth := w[Rook] | w[Queen] | b[Rook] | b[Queen]
	return pawnAttacks[White][sq]&b[Pawn] |
		pawnAttacks[Black][sq]&w[Pawn] |
		knightAttacks[sq]&(w[Knight]|b[Knight]) |
		kingAttacks[sq]&(w[King]|b[King]) |
		BishopAttacks(sq, occ)&diag |
		RookAttacks(sq, occ)&orth
}

// attackedBy returns every square attacked by color c given occupancy occ.
func (p *Position) attackedBy(c Color, occ Bitboard) Bitboard {
	pc := &p.pieces[c]
	var a Bitboard
	if c == White {
		a = (pc[Pawn]<<7)&^fileH | (pc[Pawn]<<9)&^fileA
	} else {
		a = (pc[Pawn]>>9)&^fileH | (pc[Pawn]>>7)&^fileA
	}
	for b := pc[Knight]; b != 0; {
		a |= knightAttacks[b.PopLSB()]
	}
	for b := pc[Bishop] | pc[Queen]; b != 0; {
		a |= BishopAttacks(b.PopLSB(), occ)
	}
	for b := pc[Rook] | pc[Queen]; b != 0; {
		a |= RookAttacks(b.PopLSB(), occ)
	}
	return a | kingAttacks[pc[King].LSB()]
}

// computeHash recomputes the Zobrist hash from scratch.
func (p *Position) computeHash() uint64 {
	var h uint64
	for s := A1; s < NoSquare; s++ {
		if pc := p.board[s]; pc != NoPiece {
			h ^= zobrist.piece[pc][s]
		}
	}
	h ^= zobrist.castling[p.castling]
	if p.ep != NoSquare {
		h ^= zobrist.epFile[p.ep.File()]
	}
	if p.side == Black {
		h ^= zobrist.side
	}
	return h
}
