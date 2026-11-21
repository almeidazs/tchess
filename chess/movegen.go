package chess

// LegalMoves fills ml with every legal move in p.
//
// It generates strictly legal moves directly, without "make and test":
//   - the king may only step on squares the enemy does not attack (computed
//     with the king removed, so it cannot retreat along a checking ray);
//   - in double check only the king moves;
//   - in single check every other piece must capture the checker or block,
//     enforced with a check mask;
//   - absolutely pinned pieces may only move along their pin line;
//   - en passant, the one move that can expose a check along the rank by
//     removing two pieces at once, gets an exact test.
func (p *Position) LegalMoves(ml *MoveList) {
	ml.n = 0
	us, them := p.side, p.side.Other()
	own, enemy, occ := p.occ[us], p.occ[them], p.all
	ksq := p.KingSquare(us)
	tp := &p.pieces[them]

	danger := p.attackedBy(them, occ^ksq.Bitboard())
	ml.addTargets(ksq, kingAttacks[ksq]&^own&^danger, enemy)

	checkers := p.attackersTo(ksq, occ) & enemy
	if checkers.Count() > 1 {
		return
	}
	checkMask := ^Bitboard(0)
	if checkers != 0 {
		checkMask = between[ksq][checkers.LSB()] | checkers
	}

	// Pinned pieces: enemy sliders that would see our king if exactly one of
	// our pieces were removed from the line.
	var pinned Bitboard
	snipers := RookAttacks(ksq, enemy)&(tp[Rook]|tp[Queen]) |
		BishopAttacks(ksq, enemy)&(tp[Bishop]|tp[Queen])
	for snipers != 0 {
		blockers := between[ksq][snipers.PopLSB()] & occ
		if blockers.Count() == 1 && blockers&own != 0 {
			pinned |= blockers
		}
	}

	op := &p.pieces[us]
	for b := op[Knight] &^ pinned; b != 0; { // a pinned knight can never move
		from := b.PopLSB()
		ml.addTargets(from, knightAttacks[from]&^own&checkMask, enemy)
	}
	for b := op[Bishop] | op[Queen]; b != 0; {
		from := b.PopLSB()
		t := BishopAttacks(from, occ) &^ own & checkMask
		if pinned.Has(from) {
			t &= line[ksq][from]
		}
		ml.addTargets(from, t, enemy)
	}
	for b := op[Rook] | op[Queen]; b != 0; {
		from := b.PopLSB()
		t := RookAttacks(from, occ) &^ own & checkMask
		if pinned.Has(from) {
			t &= line[ksq][from]
		}
		ml.addTargets(from, t, enemy)
	}

	p.genPawnMoves(ml, ksq, pinned, checkMask)

	if checkers == 0 {
		p.genCastling(ml, occ, danger)
	}
}

// genPawnMoves adds pushes, double pushes, captures, promotions and en
// passant, respecting pins and the check mask.
func (p *Position) genPawnMoves(ml *MoveList, ksq Square, pinned, checkMask Bitboard) {
	us := p.side
	enemy, occ := p.occ[us.Other()], p.all
	push, startRank := 8, 1
	if us == Black {
		push, startRank = -8, 6
	}
	for b := p.pieces[us][Pawn]; b != 0; {
		from := b.PopLSB()
		mask := checkMask
		if pinned.Has(from) {
			mask &= line[ksq][from]
		}
		if to := Square(int(from) + push); !occ.Has(to) {
			if mask.Has(to) {
				ml.addPawn(from, to, false)
			}
			if to2 := Square(int(to) + push); from.Rank() == startRank && !occ.Has(to2) && mask.Has(to2) {
				ml.add(NewMove(from, to2, FlagDoublePush))
			}
		}
		for caps := pawnAttacks[us][from] & enemy & mask; caps != 0; {
			ml.addPawn(from, caps.PopLSB(), true)
		}
		if p.ep != NoSquare && pawnAttacks[us][from].Has(p.ep) && p.enPassantIsLegal(from, ksq) {
			ml.add(NewMove(from, p.ep, FlagEnPassant))
		}
	}
}

// enPassantIsLegal plays the capture on a scratch occupancy and checks that
// nothing attacks our king afterwards. This covers checks it must resolve,
// pins on the capturing pawn and the horizontal discovered check through
// both pawns.
func (p *Position) enPassantIsLegal(from, ksq Square) bool {
	us, them := p.side, p.side.Other()
	captured := p.ep ^ 8
	occ := (p.all ^ from.Bitboard() ^ captured.Bitboard()) | p.ep.Bitboard()
	tp := &p.pieces[them]
	attackers := RookAttacks(ksq, occ)&(tp[Rook]|tp[Queen]) |
		BishopAttacks(ksq, occ)&(tp[Bishop]|tp[Queen]) |
		knightAttacks[ksq]&tp[Knight] |
		pawnAttacks[us][ksq]&tp[Pawn]&^captured.Bitboard()
	return attackers == 0
}

// genCastling adds castling moves whose path is empty and safe. The caller
// guarantees the king is not in check.
func (p *Position) genCastling(ml *MoveList, occ, danger Bitboard) {
	type castle struct {
		right       CastlingRights
		king, to    Square
		empty, safe Bitboard
		flag        MoveFlag
	}
	var options [2]castle
	if p.side == White {
		options = [2]castle{
			{WhiteKingside, E1, G1, F1.Bitboard() | G1.Bitboard(), F1.Bitboard() | G1.Bitboard(), FlagKingCastle},
			{WhiteQueenside, E1, C1, B1.Bitboard() | C1.Bitboard() | D1.Bitboard(), C1.Bitboard() | D1.Bitboard(), FlagQueenCastle},
		}
	} else {
		options = [2]castle{
			{BlackKingside, E8, G8, F8.Bitboard() | G8.Bitboard(), F8.Bitboard() | G8.Bitboard(), FlagKingCastle},
			{BlackQueenside, E8, C8, B8.Bitboard() | C8.Bitboard() | D8.Bitboard(), C8.Bitboard() | D8.Bitboard(), FlagQueenCastle},
		}
	}
	for _, c := range options {
		if p.castling&c.right != 0 && occ&c.empty == 0 && danger&c.safe == 0 {
			ml.add(NewMove(c.king, c.to, c.flag))
		}
	}
}

// addTargets adds a move from "from" to every target, flagging captures.
func (ml *MoveList) addTargets(from Square, targets, enemy Bitboard) {
	for t := targets & enemy; t != 0; {
		ml.add(NewMove(from, t.PopLSB(), FlagCapture))
	}
	for t := targets &^ enemy; t != 0; {
		ml.add(NewMove(from, t.PopLSB(), FlagQuiet))
	}
}

// addPawn adds a pawn move, expanded into the four promotions on the last rank.
func (ml *MoveList) addPawn(from, to Square, capture bool) {
	var flag MoveFlag
	if capture {
		flag = FlagCapture
	}
	if r := to.Rank(); r == 0 || r == 7 {
		for _, promo := range [...]MoveFlag{FlagPromoQueen, FlagPromoRook, FlagPromoBishop, FlagPromoKnight} {
			ml.add(NewMove(from, to, promo|flag))
		}
		return
	}
	ml.add(NewMove(from, to, flag))
}
