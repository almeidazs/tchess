package pixel

import (
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func TestHexRoundTrip(t *testing.T) {
	c := Hex("#F0D9B5")
	if c != (RGB{0xF0, 0xD9, 0xB5}) || c.Hex() != "#F0D9B5" {
		t.Fatalf("got %v / %s", c, c.Hex())
	}
	if Hex("1a2b3c") != (RGB{0x1A, 0x2B, 0x3C}) {
		t.Fatal("the # is optional")
	}
}

func TestBlend(t *testing.T) {
	a, b := RGB{0, 0, 0}, RGB{200, 100, 50}
	if Blend(a, b, 0) != a || Blend(a, b, 1) != b || Blend(a, b, 7) != b || Blend(a, b, -1) != a {
		t.Fatal("blend endpoints / clamping")
	}
	if got := Blend(a, b, 0.5); got != (RGB{100, 50, 25}) {
		t.Fatalf("midpoint %v", got)
	}
}

func TestCanvasClipping(t *testing.T) {
	c := NewCanvas(4, 4)
	c.Fill(-2, -2, 3, 3, RGB{9, 9, 9}) // only (0,0) is on the canvas
	c.Set(10, 10, RGB{1, 1, 1})        // ignored
	if c.At(0, 0) != (RGB{9, 9, 9}) || c.At(1, 1) != (RGB{}) {
		t.Fatal("fill should clip to the canvas")
	}
	if c.Width() != 4 || c.Height() != 4 || c.Contains(4, 0) || !c.Contains(3, 3) {
		t.Fatal("bounds")
	}
}

func TestLinesUseHalfBlocks(t *testing.T) {
	red, blue := RGB{255, 0, 0}, RGB{0, 0, 255}
	c := NewCanvas(2, 2)
	c.Set(0, 0, red) // column 0: red over blue -> "▀"
	c.Set(0, 1, blue)
	c.Set(1, 0, blue) // column 1: blue over blue -> " "
	c.Set(1, 1, blue)

	lines := c.Lines(termenv.TrueColor)
	if len(lines) != 1 {
		t.Fatalf("2 pixel rows should print as 1 line, got %d", len(lines))
	}
	l := lines[0]
	for _, want := range []string{"38;2;255;0;0m▀", "48;2;0;0;255m", "▀ "} {
		if !strings.Contains(l, want) {
			t.Errorf("line %q missing %q", l, want)
		}
	}
	if strings.Count(l, "48;2;0;0;255") != 1 {
		t.Error("an unchanged background must not be re-emitted")
	}
	if plain := c.Lines(termenv.Ascii)[0]; !strings.HasPrefix(plain, "▀ ") {
		t.Errorf("ascii profile should emit no colors, got %q", plain)
	}
}
