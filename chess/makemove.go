package chess

// Undo holds what Make cannot recompute, so Unmake can restore it.
type Undo struct {
	captured Piece
	castling CastlingRights
	ep       Square
	halfmove uint16
	hash     uint64
}

// Make plays m, which must be legal in p, and returns the data Unmake needs.
func (p *Position) Make(m Move) Undo {
	u := Undo{captured: NoPiece, castling: p.castling, ep: p.ep, halfmove: p.halfmove, hash: p.hash}
	us, them := p.side, p.side.Other()
	from, to := m.From(), m.To()
	moving := p.board[from]

	if p.ep != NoSquare {
		p.hash ^= zobrist.epFile[p.ep.File()]
		p.ep = NoSquare
	}

	switch {
	case m.IsEnPassant():
		u.captured = p.remove(to ^ 8) // the captured pawn sits behind the target square
	case m.IsCapture():
		u.captured = p.remove(to)
	}
	if u.captured != NoPiece || moving.Type() == Pawn {
		p.halfmove = 0
	} else {
		p.halfmove++
	}

	p.move(from, to)
	switch {
	case m.IsPromotion():
		p.remove(to)
		p.put(MakePiece(us, m.Promotion()), to)
	case m.IsCastle():
		rf, rt := castleRookSquares(to)
		p.move(rf, rt)
	case m.Flag() == FlagDoublePush:
		// Record the en passant square only if an enemy pawn could take, so
		// identical positions hash identically.
		ep := (from + to) / 2
		if pawnAttacks[us][ep]&p.pieces[them][Pawn] != 0 {
			p.ep = ep
			p.hash ^= zobrist.epFile[ep.File()]
		}
	}

	if rights := p.castling &^ (castleClear[from] | castleClear[to]); rights != p.castling {
		p.hash ^= zobrist.castling[p.castling] ^ zobrist.castling[rights]
		p.castling = rights
	}
	if us == Black {
		p.fullmove++
	}
	p.side = them
	p.hash ^= zobrist.side
	return u
}

// Unmake reverts m, which must be the last move made with the given undo.
func (p *Position) Unmake(m Move, u Undo) {
	p.side = p.side.Other()
	us := p.side
	from, to := m.From(), m.To()
	if us == Black {
		p.fullmove--
	}
	switch {
	case m.IsPromotion():
		p.remove(to)
		p.put(MakePiece(us, Pawn), to)
	case m.IsCastle():
		rf, rt := castleRookSquares(to)
		p.move(rt, rf)
	}
	p.move(to, from)
	if u.captured != NoPiece {
		cs := to
		if m.IsEnPassant() {
			cs = to ^ 8
		}
		p.put(u.captured, cs)
	}
	p.castling, p.ep, p.halfmove, p.hash = u.castling, u.ep, u.halfmove, u.hash
}

// put places pc on an empty square s, updating bitboards, mailbox and hash.
func (p *Position) put(pc Piece, s Square) {
	bb := s.Bitboard()
	c := pc.Color()
	p.pieces[c][pc.Type()] |= bb
	p.occ[c] |= bb
	p.all |= bb
	p.board[s] = pc
	p.hash ^= zobrist.piece[pc][s]
}

// remove takes the piece off s and returns it.
func (p *Position) remove(s Square) Piece {
	pc := p.board[s]
	bb := s.Bitboard()
	c := pc.Color()
	p.pieces[c][pc.Type()] &^= bb
	p.occ[c] &^= bb
	p.all &^= bb
	p.board[s] = NoPiece
	p.hash ^= zobrist.piece[pc][s]
	return pc
}

// move moves the piece on from to the empty square to.
func (p *Position) move(from, to Square) {
	pc := p.board[from]
	bb := from.Bitboard() | to.Bitboard()
	c := pc.Color()
	p.pieces[c][pc.Type()] ^= bb
	p.occ[c] ^= bb
	p.all ^= bb
	p.board[from], p.board[to] = NoPiece, pc
	p.hash ^= zobrist.piece[pc][from] ^ zobrist.piece[pc][to]
}
