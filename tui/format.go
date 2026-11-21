package tui

import (
	"fmt"
	"time"
)

// formatClock shows a clock: h:mm:ss, m:ss, or s.d under ten seconds.
func formatClock(d time.Duration) string {
	switch {
	case d >= time.Hour:
		return fmt.Sprintf("%d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
	case d < 10*time.Second:
		return fmt.Sprintf("0:%02d.%d", int(d.Seconds()), int(d.Milliseconds()/100)%10)
	}
	return fmt.Sprintf("%d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}

// formatMoveTime shows a move's thinking time: "1.4s" or "1:05".
func formatMoveTime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}
