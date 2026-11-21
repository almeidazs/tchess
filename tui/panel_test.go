package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
)

func TestMoveTimesInHistory(t *testing.T) {
	m := newGame(t, 120, 40, "none")
	start := time.Unix(1000, 0)
	m.now = func() time.Time { return start }
	m.turnStart = start

	click(m, chess.E2)
	m.now = func() time.Time { return start.Add(2500 * time.Millisecond) }
	click(m, chess.E4)
	click(m, chess.E7)
	m.now = func() time.Time { return start.Add(3*time.Second + 300*time.Millisecond) }
	click(m, chess.E5)

	if len(m.moveTimes) != 2 || m.moveTimes[0] != 2500*time.Millisecond || m.moveTimes[1] != 800*time.Millisecond {
		t.Fatalf("move times %v", m.moveTimes)
	}
	out := ansi.Strip(m.View())
	for _, want := range []string{"2.5s", "0.8s", "████"} {
		if !strings.Contains(out, want) {
			t.Errorf("history missing %q:\n%s", want, out)
		}
	}
}

func TestMoveTimeExcludesAnimation(t *testing.T) {
	m := New(Options{Config: config.Default()})
	m.Update(tea.WindowSizeMsg{Width: 160, Height: 50})
	start := time.Unix(1000, 0)
	m.now = func() time.Time { return start }
	m.startGame(clock.TimeControl{})
	click(m, chess.E2)
	click(m, chess.E4) // animated: black starts thinking after the turn
	m.now = func() time.Time { return start.Add(slideDur + flipDur + time.Second) }
	click(m, chess.E7)
	click(m, chess.E5)
	if got := m.moveTimes[1]; got != time.Second {
		t.Fatalf("black's move time should exclude the animation, got %v", got)
	}
}

func TestFormatMoveTime(t *testing.T) {
	cases := map[time.Duration]string{
		1400 * time.Millisecond:  "1.4s",
		25600 * time.Millisecond: "25.6s",
		65 * time.Second:         "1:05",
	}
	for d, want := range cases {
		if got := formatMoveTime(d); got != want {
			t.Errorf("formatMoveTime(%v) = %q, want %q", d, got, want)
		}
	}
	if timeBar(0, time.Second) != "    " || timeBar(time.Second, time.Second) != "████" {
		t.Error("timeBar extremes")
	}
	if got := timeBar(time.Millisecond, time.Hour); got != "▏   " {
		t.Errorf("tiny times still get a sliver, got %q", got)
	}
}

func TestCapturedPiecesShown(t *testing.T) {
	m := newGame(t, 120, 40, "none")
	for _, mv := range [][2]chess.Square{{chess.E2, chess.E4}, {chess.D7, chess.D5}, {chess.E4, chess.D5}} {
		click(m, mv[0])
		click(m, mv[1])
	}
	line := ansi.Strip(m.capturedLine(chess.White))
	if line != "♟ +1" {
		t.Fatalf("white's captures line %q, want \"♟ +1\"", line)
	}
	if ansi.Strip(m.capturedLine(chess.Black)) != "" {
		t.Fatal("black has captured nothing and is behind")
	}
	if !strings.Contains(ansi.Strip(m.View()), "♟ +1") {
		t.Fatal("captured pieces should be visible in the panel")
	}
}
