package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"sort"
	"time"
)

// writeGIF quantizes all frames to one shared 256-color palette (the most
// frequent colors) and stores each frame as just the rectangle that changed.
func writeGIF(path string, frames []frame) {
	counts := map[color.RGBA]int{}
	for _, f := range frames {
		px := f.img.Pix
		for i := 0; i < len(px); i += 4 * 3 { // every third pixel is plenty
			counts[color.RGBA{px[i], px[i+1], px[i+2], 0xFF}]++
		}
	}
	type kv struct {
		c color.RGBA
		n int
	}
	var all []kv
	for c, n := range counts {
		all = append(all, kv{c, n})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].n > all[j].n })
	pal := color.Palette{}
	for i := 0; i < len(all) && i < 256; i++ {
		pal = append(pal, all[i].c)
	}

	cache := map[color.RGBA]uint8{}
	index := func(c color.RGBA) uint8 {
		if i, ok := cache[c]; ok {
			return i
		}
		i := uint8(pal.Index(c))
		cache[c] = i
		return i
	}

	out := &gif.GIF{}
	var prev *image.RGBA
	for _, f := range frames {
		b := f.img.Bounds()
		if prev != nil {
			b = diffBounds(prev, f.img)
			if b.Empty() {
				out.Delay[len(out.Delay)-1] += centis(f.delay)
				continue
			}
		}
		p := image.NewPaletted(b, pal)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				p.SetColorIndex(x, y, index(f.img.RGBAAt(x, y)))
			}
		}
		out.Image = append(out.Image, p)
		out.Delay = append(out.Delay, centis(f.delay))
		out.Disposal = append(out.Disposal, gif.DisposalNone)
		prev = f.img
	}
	out.Config = image.Config{ColorModel: pal, Width: frames[0].img.Bounds().Dx(), Height: frames[0].img.Bounds().Dy()}

	file, err := os.Create(path)
	if err != nil {
		fail(err)
	}
	defer file.Close()
	if err := gif.EncodeAll(file, out); err != nil {
		fail(err)
	}
	fmt.Printf("wrote %s (%d frames)\n", path, len(out.Image))
}

// centis converts a duration to GIF delay units (1/100 s), at least 2.
func centis(d time.Duration) int { return max(int(d/(10*time.Millisecond)), 2) }

// diffBounds returns the rectangle where two frames differ.
func diffBounds(a, b *image.RGBA) image.Rectangle {
	r := image.Rectangle{}
	bb := b.Bounds()
	for y := bb.Min.Y; y < bb.Max.Y; y++ {
		ia, ib := a.PixOffset(bb.Min.X, y), b.PixOffset(bb.Min.X, y)
		rowA, rowB := a.Pix[ia:ia+4*bb.Dx()], b.Pix[ib:ib+4*bb.Dx()]
		for x := range bb.Dx() {
			if rowA[4*x] != rowB[4*x] || rowA[4*x+1] != rowB[4*x+1] || rowA[4*x+2] != rowB[4*x+2] {
				r = r.Union(image.Rect(bb.Min.X+x, y, bb.Min.X+x+1, y+1))
			}
		}
	}
	return r
}
