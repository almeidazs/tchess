package chess

// refresh regenerates legal moves for the current position, records its
// repetition key and checks the automatic termination rules.
func (g *Game) refresh() {
	g.pos.LegalMoves(&g.legal)
	g.keys = append(g.keys, g.repetitionKey())

	switch {
	case g.legal.n == 0 && g.pos.InCheck():
		g.end(winner(g.pos.side.Other()), Checkmate)
	case g.legal.n == 0:
		g.end(Draw, Stalemate)
	case g.pos.IsInsufficientMaterial():
		g.end(Draw, InsufficientMaterial)
	case g.RepetitionCount() >= 5:
		g.end(Draw, FivefoldRepetition)
	case g.pos.halfmove >= 150:
		g.end(Draw, SeventyFiveMoveRule)
	}
}

// repetitionKey identifies the position for repetition purposes. Under FIDE
// rules two positions differ if en passant was possible in one and not the
// other; the hash already includes the en passant file whenever an enemy
// pawn is adjacent, so it is removed here when no legal capture exists
// (e.g. the capturing pawn is pinned).
func (g *Game) repetitionKey() uint64 {
	key := g.pos.hash
	if g.pos.ep == NoSquare {
		return key
	}
	for _, m := range g.legal.Slice() {
		if m.IsEnPassant() {
			return key
		}
	}
	return key ^ zobrist.epFile[g.pos.ep.File()]
}

// RepetitionCount returns how many times the current position has occurred,
// counting only positions since the last capture or pawn move (earlier ones
// cannot be identical).
func (g *Game) RepetitionCount() int {
	n := len(g.keys) - 1
	cur, count := g.keys[n], 1
	for i := n - 2; i >= 0 && n-i <= int(g.pos.halfmove); i -= 2 {
		if g.keys[i] == cur {
			count++
		}
	}
	return count
}

// ClaimableDraw returns the draw the side to move may claim right now, or
// NoTermination.
func (g *Game) ClaimableDraw() Termination {
	switch {
	case g.IsOver():
		return NoTermination
	case g.RepetitionCount() >= 3:
		return ThreefoldRepetition
	case g.pos.halfmove >= 100:
		return FiftyMoveRule
	}
	return NoTermination
}

// ClaimDraw ends the game with the currently claimable draw.
func (g *Game) ClaimDraw() error {
	t := g.ClaimableDraw()
	if t == NoTermination {
		return ErrNoClaim
	}
	g.end(Draw, t)
	return nil
}

// AgreeDraw ends the game by mutual agreement.
func (g *Game) AgreeDraw() error {
	if g.IsOver() {
		return ErrGameOver
	}
	g.end(Draw, DrawAgreement)
	return nil
}

// Resign ends the game with c resigning.
func (g *Game) Resign(c Color) error {
	if g.IsOver() {
		return ErrGameOver
	}
	g.end(winner(c.Other()), Resignation)
	return nil
}

// FlagFall ends the game because c ran out of time. The opponent wins
// unless they lack mating material, in which case it is a draw.
func (g *Game) FlagFall(c Color) error {
	if g.IsOver() {
		return ErrGameOver
	}
	if !g.pos.HasMatingMaterial(c.Other()) {
		g.end(Draw, TimeoutVsInsufficientMaterial)
	} else {
		g.end(winner(c.Other()), Timeout)
	}
	return nil
}

// Record ends the game with a result the position does not reach on its
// own, as read from a game record. t says why: DrawAgreement,
// ThreefoldRepetition or FiftyMoveRule for a draw, Resignation for a win,
// Timeout or TimeoutVsInsufficientMaterial, or RecordedResult when the
// record does not say. o must agree with t.
func (g *Game) Record(o Outcome, t Termination) error {
	if g.IsOver() {
		return ErrGameOver
	}
	ok := false
	switch t {
	case DrawAgreement, ThreefoldRepetition, FiftyMoveRule, TimeoutVsInsufficientMaterial:
		ok = o == Draw
	case Resignation, Timeout:
		ok = o == WhiteWins || o == BlackWins
	case RecordedResult:
		ok = o != Ongoing
	}
	if !ok {
		return ErrBadResult
	}
	g.end(o, t)
	return nil
}

// end records the result.
func (g *Game) end(o Outcome, t Termination) {
	g.outcome, g.termination = o, t
}
