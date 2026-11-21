package chess

// Outcome is the result of a game.
type Outcome uint8

const (
	Ongoing Outcome = iota
	WhiteWins
	BlackWins
	Draw
)

func (o Outcome) String() string {
	switch o {
	case WhiteWins:
		return "1-0"
	case BlackWins:
		return "0-1"
	case Draw:
		return "1/2-1/2"
	}
	return "*"
}

// winner returns the outcome where c wins.
func winner(c Color) Outcome {
	if c == White {
		return WhiteWins
	}
	return BlackWins
}

// Termination explains why a game ended.
type Termination uint8

const (
	NoTermination Termination = iota
	Checkmate
	Stalemate
	InsufficientMaterial
	FivefoldRepetition  // automatic
	SeventyFiveMoveRule // automatic
	ThreefoldRepetition // claimed
	FiftyMoveRule       // claimed
	DrawAgreement
	Resignation
	Timeout
	TimeoutVsInsufficientMaterial
)
