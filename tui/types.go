package tui

import (
	"time"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/internal/pixel"
)

// This file holds the interface's state types. Their behavior lives in the
// file of each feature (layout.go, board.go, mouse.go, anim.go, menu.go…).

// ---------------------------------------------------------------------------
// Screens and modes
// ---------------------------------------------------------------------------

// screen is the top-level page being shown.
type screen uint8

const (
	screenMenu screen = iota
	screenGame
)

// mode is the interaction state while on the game screen.
type mode uint8

const (
	modePlay          mode = iota // waiting for a move
	modeInput                     // typing a move after ':'
	modePromotion                 // choosing the promotion piece
	modeConfirmResign             // "resign? y/n"
	modeDrawOffer                 // the opponent answers a draw offer
	modeConfirmQuit               // "quit? y/n"
	modeGameOver                  // result shown, waiting for new game / quit
)

// ---------------------------------------------------------------------------
// Geometry
// ---------------------------------------------------------------------------

// layout places every element of the game screen. It is a pure function of
// the terminal size, so mouse hit-testing uses exactly what View drew.
type layout struct {
	cellW, cellH     int
	pixel            bool // sprite board (false: glyph board)
	totalW, totalH   int
	originX, originY int // top-left corner of the whole game screen
	boardX, boardY   int // top-left corner of the top-left board cell
}

// cell is a board square's position on screen: column and row from the
// top-left, independent of orientation.
type cell struct{ col, row int }

// ---------------------------------------------------------------------------
// Board rendering
// ---------------------------------------------------------------------------

// boardState is what every square needs to know, computed once per frame.
type boardState struct {
	pos      chess.Position
	last     chess.Move
	hasLast  bool
	checkSq  chess.Square
	targets  chess.Bitboard
	marked   chess.Bitboard // pieces the typed move may refer to
	pickCol  int            // promotion picker column, -1 when closed
	pickRow0 int
	drag     dragView
}

// squareLook is everything needed to draw one square, shared by the pixel
// and the glyph renderers.
type squareLook struct {
	base    pixel.RGB   // checkerboard color
	bg      pixel.RGB   // background after highlights
	check   bool        // king in check: red glow
	piece   chess.Piece // NoPiece for none
	dot     bool        // legal destination, empty square
	capture bool        // legal destination with a capture
	cursor  bool        // keyboard cursor
	ghost   bool        // origin of the dragged piece
}

// ---------------------------------------------------------------------------
// Mouse
// ---------------------------------------------------------------------------

// dragState tracks a piece being dragged with the mouse.
type dragState struct {
	from        chess.Square
	x, y        int  // pointer, in terminal cells
	moved       bool // pointer has left the origin square
	wasSelected bool // the piece was already selected when pressed
}

// dragView describes the drag for the renderers: the piece under the
// pointer, the origin square to show as a ghost and the hovered target.
type dragView struct {
	active bool
	piece  chess.Piece
	from   chess.Square
	hover  chess.Square // legal target under the pointer, or NoSquare
	x, y   int          // pointer relative to the board's top-left cell
}

// ---------------------------------------------------------------------------
// Animation
// ---------------------------------------------------------------------------

// animation plays after a move on the sprite board: first the moved
// piece(s) slide, then, if the turn passes, every piece sweeps around the
// board center until the board faces the other player. Squares need no
// animation: the checkerboard is symmetric under a 180° turn.
type animation struct {
	start    time.Time
	before   chess.Position // position before the move (slide phase)
	movers   []slide
	from, to chess.Color // orientation before / after
}

// slide is one piece moving during the slide phase (two when castling).
type slide struct {
	piece    chess.Piece
	from, to chess.Square
}

// ---------------------------------------------------------------------------
// Menu
// ---------------------------------------------------------------------------

// menuItem is a time-control preset.
type menuItem struct {
	label string
	tc    string // clock.Parse input; "" means "custom"
}

// menuState is the menu's cursor and custom time-control input.
type menuState struct {
	cursor  int
	editing bool // typing a custom time control
	input   string
	err     string
	notice  string // feedback for the theme and sound keys
	custom  string // last custom time control, offered again when editing
}
