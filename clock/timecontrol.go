package clock

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeControl is a base time plus a per-move increment. The zero value means
// "no clock".
type TimeControl struct {
	Base      time.Duration
	Increment time.Duration
}

// Parse reads "minutes+seconds" ("5+3", "10", "0.5+1") or "none".
func Parse(s string) (TimeControl, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "none" || s == "-" || s == "0" {
		return TimeControl{}, nil
	}
	baseStr, incStr, hasInc := strings.Cut(s, "+")
	base, err := strconv.ParseFloat(baseStr, 64)
	if err != nil || base <= 0 || base > 600 {
		return TimeControl{}, fmt.Errorf("tempo base inválido %q (minutos entre 0 e 600)", baseStr)
	}
	var inc float64
	if hasInc {
		inc, err = strconv.ParseFloat(incStr, 64)
		if err != nil || inc < 0 || inc > 600 {
			return TimeControl{}, fmt.Errorf("incremento inválido %q (segundos entre 0 e 600)", incStr)
		}
	}
	return TimeControl{
		Base:      time.Duration(base * float64(time.Minute)),
		Increment: time.Duration(inc * float64(time.Second)),
	}, nil
}

// Unlimited reports whether the time control means playing without a clock.
func (tc TimeControl) Unlimited() bool { return tc.Base <= 0 }

// String formats as "minutes+seconds", e.g. "5+3".
func (tc TimeControl) String() string {
	if tc.Unlimited() {
		return "∞"
	}
	return strconv.FormatFloat(tc.Base.Minutes(), 'f', -1, 64) + "+" +
		strconv.FormatFloat(tc.Increment.Seconds(), 'f', -1, 64)
}
