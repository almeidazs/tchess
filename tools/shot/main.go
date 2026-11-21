// Command shot renders the real tchess interface to images for the README:
// a PNG screenshot of a game in progress and an animated GIF demo. It drives
// the Bubble Tea model directly (window size, keys, clicks, drags) with a
// fake clock, then rasterizes each rendered frame.
//
//	cd tools/shot && go run . -out ../../assets
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const title = "tchess"

func main() {
	out := flag.String("out", "assets", "output directory")
	flag.Parse()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}
	screenshot(filepath.Join(*out, "screenshot.png"))
	demo(filepath.Join(*out, "demo.gif"))
}

// fail prints the error and exits.
func fail(err error) {
	fmt.Fprintln(os.Stderr, "shot:", err)
	os.Exit(1)
}
