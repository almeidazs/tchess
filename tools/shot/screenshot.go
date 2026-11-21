package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/clock"
)

// screenshot renders the README picture: a Møller Attack position with the
// f3 knight selected.
func screenshot(path string) {
	tc := clock.TimeControl{Base: 5 * time.Minute, Increment: 3 * time.Second}
	s := newSession(126, 47, &tc)
	moves := []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Bc5", "c3", "Nf6", "d4", "exd4", "cxd4", "Bb4+",
		"Nc3", "Nxe4", "O-O", "Bxc3", "d5", "Bf6", "Re1", "Ne7", "Rxe4", "d6", "Bg5", "Bxg5"}
	thinks := []float64{1.1, 0.9, 1.4, 2.2, 1.8, 3.1, 4.6, 2.4, 6.2, 3.3, 1.2, 8.7,
		11.4, 5.1, 3.9, 14.2, 21.6, 9.8, 4.4, 7.5, 2.9, 12.3, 18.1, 4.2}
	for i, san := range moves {
		s.play(san, time.Duration(thinks[i]*float64(time.Second)))
	}
	// White to move: lift the f3 knight to show its moves (Nxg5 captures).
	s.now = s.now.Add(6 * time.Second)
	s.click(chess.F3)

	fs := newFontSet(12)
	img := fs.window(fs.renderGrid(parseANSI(s.view(), s.w, s.h)), title, true)
	writePNG(path, img)
}

// writePNG saves an image as an optimized PNG.
func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		fail(err)
	}
	defer f.Close()
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(f, img); err != nil {
		fail(err)
	}
	fmt.Println("wrote", path)
}
