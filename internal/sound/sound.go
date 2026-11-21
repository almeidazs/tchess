package sound

// Effect identifies a sound effect.
type Effect uint8

// The game's sound effects.
const (
	Move    Effect = iota // a quiet move
	Capture               // a piece was taken
	Check                 // the move gives check
	Castle                // castling
	Promote               // a pawn was promoted
	GameEnd               // the game is over
	LowTime               // a clock dropped below the warning threshold
	Error                 // a typed move was rejected
	numEffects
)

// Effects lists every effect, in declaration order.
func Effects() []Effect {
	out := make([]Effect, numEffects)
	for i := range out {
		out[i] = Effect(i)
	}
	return out
}

// String is the effect's name, also used for its file name.
func (e Effect) String() string {
	switch e {
	case Move:
		return "move"
	case Capture:
		return "capture"
	case Check:
		return "check"
	case Castle:
		return "castle"
	case Promote:
		return "promote"
	case GameEnd:
		return "gameend"
	case LowTime:
		return "lowtime"
	case Error:
		return "error"
	}
	return "unknown"
}

// Player plays sound effects. Play must not block.
type Player interface {
	Play(Effect)
}

// Nop is a Player that stays silent.
type Nop struct{}

// Play does nothing.
func (Nop) Play(Effect) {}
