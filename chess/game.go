package chess

// MoveRecord is one played move with its notation.
type MoveRecord struct {
	Move     Move
	SAN      string
	Color    Color
	Captured Piece // NoPiece if the move captured nothing
}

// Game tracks a full game: the current position, its legal moves, the move
// history and how the game ended. It applies the FIDE termination rules:
// checkmate, stalemate, dead position, fivefold repetition and the 75-move
// rule end the game automatically; threefold repetition and the 50-move rule
// are draws the player to move may claim.
type Game struct {
	pos         Position
	legal       MoveList
	history     []MoveRecord
	keys        []uint64 // repetition keys of every position, current last
	outcome     Outcome
	termination Termination
}

// NewGame starts a game from the initial position.
func NewGame() *Game {
	g, err := NewGameFromFEN(StartFEN)
	if err != nil {
		panic(err)
	}
	return g
}

// NewGameFromFEN starts a game from an arbitrary position.
func NewGameFromFEN(fen string) (*Game, error) {
	p, err := ParseFEN(fen)
	if err != nil {
		return nil, err
	}
	g := &Game{pos: *p}
	g.refresh()
	return g, nil
}

// Position returns a copy of the current position.
func (g *Game) Position() Position { return g.pos }

// Turn returns the side to move.
func (g *Game) Turn() Color { return g.pos.side }

// InCheck reports whether the side to move is in check.
func (g *Game) InCheck() bool { return g.pos.InCheck() }

// LegalMoves returns the legal moves of the side to move (empty once over).
func (g *Game) LegalMoves() []Move {
	if g.IsOver() {
		return nil
	}
	return append([]Move(nil), g.legal.Slice()...)
}

// MovesFrom returns the legal moves starting on sq.
func (g *Game) MovesFrom(sq Square) []Move {
	if g.IsOver() {
		return nil
	}
	var out []Move
	for _, m := range g.legal.Slice() {
		if m.From() == sq {
			out = append(out, m)
		}
	}
	return out
}

// History returns the moves played so far.
func (g *Game) History() []MoveRecord { return append([]MoveRecord(nil), g.history...) }

// LastMove returns the last move played, if any.
func (g *Game) LastMove() (Move, bool) {
	if len(g.history) == 0 {
		return NullMove, false
	}
	return g.history[len(g.history)-1].Move, true
}

// Captured returns the pieces c has captured, in the order they were taken.
func (g *Game) Captured(c Color) []Piece {
	var out []Piece
	for _, r := range g.history {
		if r.Color == c && r.Captured != NoPiece {
			out = append(out, r.Captured)
		}
	}
	return out
}

// Outcome returns the result, or Ongoing.
func (g *Game) Outcome() Outcome { return g.outcome }

// Termination returns why the game ended, or NoTermination.
func (g *Game) Termination() Termination { return g.termination }

// IsOver reports whether the game has ended.
func (g *Game) IsOver() bool { return g.outcome != Ongoing }

// Play makes a legal move and applies automatic game-ending rules.
func (g *Game) Play(m Move) error {
	if g.IsOver() {
		return ErrGameOver
	}
	if !g.legal.Contains(m) {
		return ErrIllegalMove
	}
	rec := MoveRecord{Move: m, SAN: g.pos.san(m, &g.legal), Color: g.pos.side, Captured: NoPiece}
	switch {
	case m.IsEnPassant():
		rec.Captured = MakePiece(g.pos.side.Other(), Pawn)
	case m.IsCapture():
		rec.Captured = g.pos.board[m.To()]
	}
	g.pos.Make(m)
	g.history = append(g.history, rec)
	g.refresh()
	return nil
}
