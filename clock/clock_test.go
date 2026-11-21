package clock

import (
	"testing"
	"time"

	"github.com/almeidazs/tchess/chess"
)

type fakeTime struct{ t time.Time }

func (f *fakeTime) now() time.Time { return f.t }

func (f *fakeTime) advance(d time.Duration) { f.t = f.t.Add(d) }

func TestClockFlow(t *testing.T) {
	ft := &fakeTime{t: time.Unix(0, 0)}
	cl := New(TimeControl{Base: time.Minute, Increment: 2 * time.Second}, ft.now)

	ft.advance(10 * time.Second) // white thinks before the clock starts: free
	if cl.Press(chess.White) {
		t.Fatal("unexpected flag")
	}
	if got := cl.Remaining(chess.White); got != time.Minute {
		t.Fatalf("white charged before first press: %v", got)
	}

	ft.advance(15 * time.Second)
	if got := cl.Remaining(chess.Black); got != 45*time.Second {
		t.Fatalf("black running: %v", got)
	}
	cl.Press(chess.Black)
	if got := cl.Remaining(chess.Black); got != 47*time.Second {
		t.Fatalf("black after increment: %v", got)
	}

	ft.advance(5 * time.Second)
	cl.Press(chess.White)
	if got := cl.Remaining(chess.White); got != 57*time.Second {
		t.Fatalf("white after move: %v", got)
	}

	ft.advance(time.Minute)
	if c, ok := cl.Flagged(); !ok || c != chess.Black {
		t.Fatalf("black should have flagged: %v %v", c, ok)
	}
	if !cl.Press(chess.Black) {
		t.Fatal("press after flag must report it")
	}
	if cl.Remaining(chess.Black) != 0 || cl.Running() {
		t.Fatal("flagged clock should stop at zero")
	}
}

func TestStop(t *testing.T) {
	ft := &fakeTime{t: time.Unix(0, 0)}
	cl := New(TimeControl{Base: time.Minute}, ft.now)
	cl.Press(chess.White)
	ft.advance(20 * time.Second)
	cl.Stop()
	ft.advance(time.Hour)
	if got := cl.Remaining(chess.Black); got != 40*time.Second {
		t.Fatalf("stopped clock kept running: %v", got)
	}
}

func TestStartDelay(t *testing.T) {
	ft := &fakeTime{t: time.Unix(0, 0)}
	cl := New(TimeControl{Base: time.Minute}, ft.now)
	cl.SetStartDelay(time.Second)
	cl.Press(chess.White)
	ft.advance(700 * time.Millisecond) // still within the grace period
	if got := cl.Remaining(chess.Black); got != time.Minute {
		t.Fatalf("time ran during the delay: %v", got)
	}
	ft.advance(1300 * time.Millisecond) // 1s of delay + 1s of thinking
	if got := cl.Remaining(chess.Black); got != 59*time.Second {
		t.Fatalf("after delay: %v", got)
	}
	cl.Press(chess.Black)
	if got := cl.Remaining(chess.Black); got != 59*time.Second {
		t.Fatalf("after press: %v", got)
	}
}

func TestRestore(t *testing.T) {
	now := time.Unix(0, 0)
	cl := New(TimeControl{Base: time.Minute, Increment: 2 * time.Second}, func() time.Time { return now })
	cl.Restore([2]time.Duration{30 * time.Second, 40 * time.Second}, chess.Black, true)
	now = now.Add(5 * time.Second)
	if got := cl.Remaining(chess.Black); got != 35*time.Second {
		t.Fatalf("black %v, want 35s", got)
	}
	if got := cl.Remaining(chess.White); got != 30*time.Second {
		t.Fatalf("white %v, want 30s", got)
	}

	// Stopped: nothing runs until the side to move presses.
	cl.Restore([2]time.Duration{30 * time.Second, 40 * time.Second}, chess.White, false)
	now = now.Add(time.Hour)
	if cl.Running() || cl.Remaining(chess.White) != 30*time.Second {
		t.Fatal("a stopped clock should not run")
	}
	cl.Press(chess.White)
	now = now.Add(time.Second)
	if cl.Remaining(chess.White) != 30*time.Second || cl.Remaining(chess.Black) != 39*time.Second {
		t.Fatalf("after press: %v / %v", cl.Remaining(chess.White), cl.Remaining(chess.Black))
	}
}
