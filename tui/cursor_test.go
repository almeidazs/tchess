package tui

import (
	"testing"

	"github.com/almeidazs/tchess/chess"
)

func TestCursorStopsOnMovablePieces(t *testing.T) {
	m := newGame(t, 160, 50, "none")
	key(m, "up") // reveal on e2
	if cursorSquare(m) != chess.E2 {
		t.Fatalf("cursor at %v", cursorSquare(m))
	}
	key(m, "up") // nothing movable above e2: stay
	if cursorSquare(m) != chess.E2 {
		t.Fatalf("cursor should not stop on empty squares, at %v", cursorSquare(m))
	}
	key(m, "right")
	if cursorSquare(m) != chess.F2 {
		t.Fatalf("right from e2: %v", cursorSquare(m))
	}
	key(m, "down") // g1 knight is the only movable piece on the back rank side
	if got := cursorSquare(m); got != chess.G1 && got != chess.B1 {
		t.Fatalf("down from f2 should reach a knight, at %v", got)
	}
}

func TestCursorJumpsBetweenTargets(t *testing.T) {
	m := newGame(t, 160, 50, "none")
	click(m, chess.G1) // select the knight with the mouse...
	m.cursorCol, m.cursorRow = m.cellOf(chess.G1)
	m.showCursor = true // ...then continue with the keyboard
	key(m, "left")
	if cursorSquare(m) != chess.F3 {
		t.Fatalf("left from g1 should reach f3, at %v", cursorSquare(m))
	}
	key(m, "right")
	if cursorSquare(m) != chess.H3 {
		t.Fatalf("right from f3 should reach h3, at %v", cursorSquare(m))
	}
	key(m, "right") // nothing further right: stay on h3
	if cursorSquare(m) != chess.H3 {
		t.Fatalf("cursor left the targets: %v", cursorSquare(m))
	}
	key(m, "down") // back to the piece itself
	if cursorSquare(m) != chess.G1 {
		t.Fatalf("down from h3 should return to g1, at %v", cursorSquare(m))
	}
	key(m, "up")
	key(m, "enter")
	if h := m.game.History(); len(h) != 1 || (h[0].SAN != "Nf3" && h[0].SAN != "Nh3") {
		t.Fatalf("history %v", h)
	}
}

func TestCursorSnapsAfterFlip(t *testing.T) {
	m := newGame(t, 160, 50, "none")
	key(m, "up")
	key(m, "enter") // select e2
	key(m, "up")
	key(m, "up")
	key(m, "enter") // e4; board turns to black
	key(m, "up")    // the cursor cell now shows an empty square: snap to a black piece
	pos := m.game.Position()
	if pc := pos.PieceAt(cursorSquare(m)); pc == chess.NoPiece || pc.Color() != chess.Black {
		t.Fatalf("cursor should snap to a movable black piece, at %v", cursorSquare(m))
	}
}
