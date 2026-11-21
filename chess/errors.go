package chess

import "errors"

// Errors returned by Game.
var (
	// ErrGameOver is returned when acting on a game that has ended.
	ErrGameOver = errors.New("chess: the game is over")
	// ErrIllegalMove is returned by Play for a move not legal in the position.
	ErrIllegalMove = errors.New("chess: illegal move")
	// ErrNoClaim is returned by ClaimDraw when no draw can be claimed.
	ErrNoClaim = errors.New("chess: no draw can be claimed")
)
