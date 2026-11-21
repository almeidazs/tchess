package tui

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/pixel"
)

// TestSpritePreview writes a PNG of every sprite set on a board, for
// reviewing the pixel art. Run with CHESS_PREVIEW=out.png.
func TestSpritePreview(t *testing.T) {
	out := os.Getenv("CHESS_PREVIEW")
	if out == "" {
		t.Skip("set CHESS_PREVIEW=path.png to render")
	}
	const scale = 9 // one canvas pixel ≈ one 9x9 px half cell
	sizes := [][2]int{{11, 10}, {9, 8}, {7, 6}}
	pos := chess.StartingPosition()

	var boards []*pixel.Canvas
	width, height := 0, 0
	for _, sz := range sizes {
		sw, sh := sz[0], sz[1]
		cv := pixel.NewCanvas(8*sw, 8*sh)
		for row := range 8 {
			for col := range 8 {
				sq := chess.NewSquare(col, 7-row)
				base := themes[0].Light
				if !sq.IsLight() {
					base = themes[0].Dark
				}
				cv.Fill(col*sw, row*sh, sw, sh, base)
				if pc := pos.PieceAt(sq); pc != chess.NoPiece {
					spriteSets[sw][pc.Type()].draw(cv, col*sw, row*sh, pc.Color())
				}
			}
		}
		boards = append(boards, cv)
		width += cv.Width()*scale + 20
		height = max(height, cv.Height()*scale)
	}

	// Animation frames on the largest board: e2-e4 slide, then the turn.
	m := New(Options{})
	m.Update(tea.WindowSizeMsg{Width: 160, Height: 50})
	m.startGame(clock.TimeControl{})
	start := time.Unix(0, 0)
	m.now = func() time.Time { return start }
	mv, _ := func() (chess.Move, error) { p := m.game.Position(); return p.ParseUCI("e2e4") }()
	m.play(mv)
	l := computeLayout(m.width, m.height)
	for _, at := range []time.Duration{slideDur / 2, slideDur + flipDur/4, slideDur + flipDur/2, slideDur + 3*flipDur/4} {
		m.now = func() time.Time { return start.Add(at) }
		cv := m.pixelCanvas(l)
		boards = append(boards, cv)
		width += cv.Width()*scale + 20
	}

	// A drag in progress: knight lifted from g1, hovering over f3.
	m.anim = nil
	m.startGame(clock.TimeControl{})
	mouseAt(m, chess.G1, tea.MouseActionPress)
	mouseAt(m, chess.F3, tea.MouseActionMotion)
	cv := m.pixelCanvas(l)
	boards = append(boards, cv)
	width += cv.Width()*scale + 20

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	ox := 0
	for _, cv := range boards {
		for y := range cv.Height() * scale {
			for x := range cv.Width() * scale {
				c := cv.At(x/scale, y/scale)
				img.Set(ox+x, y, color.RGBA{c.R, c.G, c.B, 255})
			}
		}
		ox += cv.Width()*scale + 20
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
