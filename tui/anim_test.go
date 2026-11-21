package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
)

func TestMoveAnimation(t *testing.T) {
	m := New(Options{})
	m.Update(tea.WindowSizeMsg{Width: 160, Height: 50})
	m.startGame(clock.TimeControl{Base: 5 * time.Minute})
	start := time.Now()
	m.now = func() time.Time { return start }

	click(m, chess.E2)
	click(m, chess.E4)
	if m.anim == nil || !m.anim.flips() {
		t.Fatal("a move should start the slide + turn animation")
	}
	if m.orientation != chess.Black || m.viewOrientation() != chess.White {
		t.Fatal("logic flips at once; the screen keeps the old side until mid-turn")
	}
	if got := m.clock.Remaining(chess.Black); got != 5*time.Minute {
		t.Fatalf("black's time must not run during the animation: %v", got)
	}
	for _, at := range []time.Duration{0, slideDur / 2, slideDur + flipDur/3, slideDur + flipDur/2 + time.Millisecond} {
		m.now = func() time.Time { return start.Add(at) }
		if out := m.View(); out == "" {
			t.Fatalf("empty frame at %v", at)
		}
	}
	if m.viewOrientation() != chess.Black {
		t.Fatal("after mid-turn the screen should face black")
	}

	m.now = func() time.Time { return start.Add(slideDur + flipDur) }
	if cmd := m.onAnimFrame(animMsg{m.animGen}); cmd != nil || m.anim != nil {
		t.Fatal("animation should end after its duration")
	}

	// Any input skips a running animation.
	click(m, chess.E7)
	click(m, chess.E5)
	if m.anim == nil {
		t.Fatal("expected a second animation")
	}
	key(m, "left")
	if m.anim != nil {
		t.Fatal("a key press should skip the animation")
	}
}

func TestDropSkipsSlideButStillTurns(t *testing.T) {
	m := New(Options{})
	m.Update(tea.WindowSizeMsg{Width: 160, Height: 50})
	m.startGame(clock.TimeControl{})
	start := time.Now()
	m.now = func() time.Time { return start }
	drag(m, chess.E2, chess.E3, chess.E4)
	if m.anim == nil || m.now().Sub(m.anim.start) != slideDur {
		t.Fatal("a drop should start the turn animation directly, without the slide")
	}
}

func TestNoAnimationOnGlyphBoard(t *testing.T) {
	m := New(Options{})
	m.Update(tea.WindowSizeMsg{Width: 70, Height: 20}) // only the glyph board fits
	m.startGame(clock.TimeControl{})
	click(m, chess.E2)
	click(m, chess.E4)
	if m.anim != nil {
		t.Fatal("the compact glyph board has no animation")
	}
}
