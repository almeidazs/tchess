package clock

import (
	"time"

	"github.com/almeidazs/tchess/chess"
)

// Clock tracks both players' remaining time. Only one side's time runs at a
// time; pressing the clock after a move stops it, adds the increment and
// starts the opponent's. The time source is injectable for tests.
type Clock struct {
	now       func() time.Time
	remaining [2]time.Duration
	increment time.Duration
	active    chess.Color
	running   bool
	since     time.Time     // when the active side's time starts running
	delay     time.Duration // grace period before the next side's time runs
}

// New creates a stopped clock. now may be nil to use time.Now.
func New(tc TimeControl, now func() time.Time) *Clock {
	if now == nil {
		now = time.Now
	}
	return &Clock{
		now:       now,
		remaining: [2]time.Duration{tc.Base, tc.Base},
		increment: tc.Increment,
	}
}

// SetStartDelay sets a grace period applied each time a side's time starts,
// e.g. while the interface animates the previous move.
func (cl *Clock) SetStartDelay(d time.Duration) { cl.delay = max(d, 0) }

// elapsed is how long the active side's time has been running.
func (cl *Clock) elapsed(now time.Time) time.Duration { return max(now.Sub(cl.since), 0) }

// Remaining returns c's time left, including the time currently elapsing.
func (cl *Clock) Remaining(c chess.Color) time.Duration {
	r := cl.remaining[c]
	if cl.running && cl.active == c {
		r -= cl.elapsed(cl.now())
	}
	return max(r, 0)
}

// Running reports whether a side's time is running.
func (cl *Clock) Running() bool { return cl.running }

// Active returns whose time is running (meaningful only while Running).
func (cl *Clock) Active() chess.Color { return cl.active }

// Press is called when c completes a move. If c's time had already run out
// it returns true and stops the clock without applying the move's
// increment. The first press (White's first move) just starts Black's time,
// so neither side is charged before the game gets going.
func (cl *Clock) Press(c chess.Color) (flagged bool) {
	now := cl.now()
	if cl.running && cl.active == c {
		cl.remaining[c] -= cl.elapsed(now)
		if cl.remaining[c] <= 0 {
			cl.remaining[c] = 0
			cl.running = false
			return true
		}
		cl.remaining[c] += cl.increment
	}
	cl.active, cl.running, cl.since = c.Other(), true, now.Add(cl.delay)
	return false
}

// Flagged reports whether the running side has run out of time.
func (cl *Clock) Flagged() (chess.Color, bool) {
	if cl.running && cl.Remaining(cl.active) <= 0 {
		return cl.active, true
	}
	return chess.White, false
}

// Stop freezes both times.
func (cl *Clock) Stop() {
	if cl.running {
		cl.remaining[cl.active] = cl.Remaining(cl.active)
		cl.running = false
	}
}

// Restore sets both players' remaining time and whose time runs, as when
// taking back a move or resuming a saved game. A running clock counts from
// now, after the start delay; a stopped one starts on the next Press.
func (cl *Clock) Restore(remaining [2]time.Duration, active chess.Color, running bool) {
	cl.remaining = remaining
	cl.active, cl.running = active, running
	cl.since = cl.now().Add(cl.delay)
}
