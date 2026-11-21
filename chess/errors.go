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
	// ErrBadResult is returned by Record for an outcome and termination
	// that do not go together, or that only the position can produce.
	ErrBadResult = errors.New("chess: result and termination do not match")
	// ErrNothingToUndo is returned by Undo before the first move.
	ErrNothingToUndo = errors.New("chess: no move to undo")
)
