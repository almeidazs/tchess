package tui

import (
	"math"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/internal/pixel"
)

const (
	slideDur  = 160 * time.Millisecond // moved piece glides to its square
	flipDur   = 620 * time.Millisecond // board turns 180° to the next player
	frameRate = 16 * time.Millisecond  // ~60 fps
)

// flips reports whether the board turns (false when the game ended).
func (a *animation) flips() bool { return a.from != a.to }

// duration is the total length of the animation.
func (a *animation) duration() time.Duration {
	if a.flips() {
		return slideDur + flipDur
	}
	return slideDur
}

// startAnimation begins the post-move animation and returns its frame
// ticker. A dropped piece is already on its square, so the slide is skipped.
func (m *Model) startAnimation(before chess.Position, mv chess.Move, from, to chess.Color, dropped bool) tea.Cmd {
	if dropped && from == to {
		return nil // nothing left to animate
	}
	a := &animation{start: m.now(), before: before, from: from, to: to}
	if dropped {
		a.start = a.start.Add(-slideDur)
	}
	a.movers = append(a.movers, slide{before.PieceAt(mv.From()), mv.From(), mv.To()})
	if mv.IsCastle() {
		rf, rt := castleRook(mv.To())
		a.movers = append(a.movers, slide{before.PieceAt(rf), rf, rt})
	}
	m.anim = a
	m.animGen++
	return animTick(m.animGen)
}

// castleRook returns the rook's squares for a castle landing on kingTo.
func castleRook(kingTo chess.Square) (from, to chess.Square) {
	switch kingTo {
	case chess.G1:
		return chess.H1, chess.F1
	case chess.C1:
		return chess.A1, chess.D1
	case chess.G8:
		return chess.H8, chess.F8
	}
	return chess.A8, chess.D8
}

// onAnimFrame ends the animation once it has run its course, or asks
// for the next frame.
func (m *Model) onAnimFrame(msg animMsg) tea.Cmd {
	if m.anim == nil || msg.gen != m.animGen {
		return nil
	}
	if m.now().Sub(m.anim.start) >= m.anim.duration() {
		m.anim = nil
		return nil
	}
	return animTick(m.animGen)
}

// viewOrientation is the orientation on screen: during the turn it switches
// halfway through, which is when labels and clocks swap sides.
func (m *Model) viewOrientation() chess.Color {
	if a := m.anim; a != nil && m.now().Sub(a.start) < slideDur+flipDur/2 {
		return a.from
	}
	return m.orientation
}

// easeInOutCubic eases t in [0, 1]: slow start, fast middle, slow end.
func easeInOutCubic(t float64) float64 {
	t = min(max(t, 0), 1)
	if t < 0.5 {
		return 4 * t * t * t
	}
	u := -2*t + 2
	return 1 - u*u*u/2
}

// drawAnimation draws the pieces for the current frame onto a plain board.
func (m *Model) drawAnimation(cv *pixel.Canvas, sw, sh int) {
	a := m.anim
	sprites := spriteSets[sw]
	elapsed := m.now().Sub(a.start)

	if elapsed < slideDur {
		t := easeInOutCubic(float64(elapsed) / float64(slideDur))
		moving := chess.Bitboard(0)
		for _, mv := range a.movers {
			moving |= mv.from.Bitboard()
		}
		for sq := chess.A1; sq < chess.NoSquare; sq++ {
			if pc := a.before.PieceAt(sq); pc != chess.NoPiece && !moving.Has(sq) {
				c, r := cellIn(a.from, sq)
				sprites[pc.Type()].draw(cv, c*sw, r*sh, pc.Color())
			}
		}
		for _, mv := range a.movers {
			c0, r0 := cellIn(a.from, mv.from)
			c1, r1 := cellIn(a.from, mv.to)
			x := float64(c0*sw) + float64((c1-c0)*sw)*t
			y := float64(r0*sh) + float64((r1-r0)*sh)*t
			sprites[mv.piece.Type()].draw(cv, int(math.Round(x)), int(math.Round(y)), mv.piece.Color())
		}
		return
	}

	pos := m.game.Position()
	t := 1.0
	if a.flips() {
		t = easeInOutCubic(float64(elapsed-slideDur) / float64(flipDur))
	}
	theta := math.Pi * t
	// Shrink slightly mid-turn so the corners stay on the board.
	scale := 1 - 0.3*math.Sin(theta)
	sin, cos := math.Sincos(theta)
	halfW, halfH := float64(sw)/2, float64(sh)/2
	centerX, centerY := 4*float64(sw), 4*float64(sh)

	type placed struct {
		pc   chess.Piece
		x, y int
	}
	var pieces []placed
	for sq := chess.A1; sq < chess.NoSquare; sq++ {
		pc := pos.PieceAt(sq)
		if pc == chess.NoPiece {
			continue
		}
		c, r := cellIn(a.from, sq)
		dx := (float64(c*sw) + halfW - centerX) * scale
		dy := (float64(r*sh) + halfH - centerY) * scale
		x := centerX + dx*cos - dy*sin - halfW
		y := centerY + dx*sin + dy*cos - halfH
		pieces = append(pieces, placed{pc, int(math.Round(x)), int(math.Round(y))})
	}
	// Pieces lower on screen are drawn last, as if nearer the viewer.
	sort.SliceStable(pieces, func(i, j int) bool { return pieces[i].y < pieces[j].y })
	for _, p := range pieces {
		sprites[p.pc.Type()].draw(cv, p.x, p.y, p.pc.Color())
	}
}
