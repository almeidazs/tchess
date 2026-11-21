package clock

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in        string
		base, inc time.Duration
	}{
		{"5+3", 5 * time.Minute, 3 * time.Second},
		{"10", 10 * time.Minute, 0},
		{" 0.5+1 ", 30 * time.Second, time.Second},
		{"none", 0, 0},
	}
	for _, c := range cases {
		tc, err := Parse(c.in)
		if err != nil || tc.Base != c.base || tc.Increment != c.inc {
			t.Errorf("Parse(%q) = %+v, %v", c.in, tc, err)
		}
	}
	for _, bad := range []string{"", "abc", "-5", "5+x", "5+-1", "1000"} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) should fail", bad)
		}
	}
	if s := (TimeControl{5 * time.Minute, 3 * time.Second}).String(); s != "5+3" {
		t.Errorf("String() = %q", s)
	}
}
