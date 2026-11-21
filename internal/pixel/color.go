package pixel

import (
	"fmt"
	"strings"
)

// RGB is a 24-bit color.
type RGB struct{ R, G, B uint8 }

// Hex parses "#RRGGBB" (the "#" is optional). It panics on malformed input,
// since colors are meant to be literals in source code.
func Hex(s string) RGB {
	var c RGB
	if _, err := fmt.Sscanf(strings.TrimPrefix(s, "#"), "%02x%02x%02x", &c.R, &c.G, &c.B); err != nil {
		panic("pixel: bad color " + s)
	}
	return c
}

// Hex formats the color as "#RRGGBB".
func (c RGB) Hex() string { return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B) }

// Blend mixes a toward b by t, clamped to [0, 1]: 0 gives a, 1 gives b.
func Blend(a, b RGB, t float64) RGB {
	t = min(max(t, 0), 1)
	mix := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return RGB{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B)}
}
