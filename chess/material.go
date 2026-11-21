package chess

// pieceValues are the conventional material values (king excluded).
var pieceValues = [6]int{Pawn: 1, Knight: 3, Bishop: 3, Rook: 5, Queen: 9}

// Value returns the conventional material value of a piece type.
func (t PieceType) Value() int {
	if t >= King {
		return 0
	}
	return pieceValues[t]
}

// Material returns the total material value of c's pieces on the board.
func (p *Position) Material(c Color) int {
	total := 0
	for t := Pawn; t < King; t++ {
		total += p.pieces[c][t].Count() * pieceValues[t]
	}
	return total
}

// IsInsufficientMaterial reports a dead position that no sequence of legal
// moves can turn into checkmate: K vs K, K+minor vs K, and positions where
// the only remaining pieces are bishops all on squares of one color.
func (p *Position) IsInsufficientMaterial() bool {
	w, b := &p.pieces[White], &p.pieces[Black]
	if w[Pawn]|b[Pawn]|w[Rook]|b[Rook]|w[Queen]|b[Queen] != 0 {
		return false
	}
	knights := w[Knight] | b[Knight]
	bishops := w[Bishop] | b[Bishop]
	if knights != 0 {
		return knights.Count() == 1 && bishops == 0
	}
	return bishops&lightSquares == 0 || bishops&darkSquares == 0
}

// HasMatingMaterial reports whether color c has more than a lone king or a
// king and a single minor piece. Used to decide a flag fall: if the opponent
// of the player who ran out of time cannot mate, the game is drawn.
func (p *Position) HasMatingMaterial(c Color) bool {
	pc := &p.pieces[c]
	if pc[Pawn]|pc[Rook]|pc[Queen] != 0 {
		return true
	}
	return (pc[Knight] | pc[Bishop]).Count() >= 2
}
