package chess

// MoveList is a fixed-capacity move buffer, so move generation never
// allocates. 256 exceeds the maximum number of legal moves in any position.
type MoveList struct {
	moves [256]Move
	n     int
}

// Len returns the number of moves.
func (ml *MoveList) Len() int { return ml.n }

// At returns the i-th move.
func (ml *MoveList) At(i int) Move { return ml.moves[i] }

// Slice returns the moves, backed by the list (no copy).
func (ml *MoveList) Slice() []Move { return ml.moves[:ml.n] }

// Contains reports whether m is in the list.
func (ml *MoveList) Contains(m Move) bool {
	for _, x := range ml.moves[:ml.n] {
		if x == m {
			return true
		}
	}
	return false
}

// add appends a move.
func (ml *MoveList) add(m Move) { ml.moves[ml.n] = m; ml.n++ }
