package main

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// window wraps a terminal image in a macOS-style window. With shadow the
// result has a transparent margin and a soft drop shadow (PNG); without it
// the window fills the image (GIF).
func (fs *fontSet) window(content *image.RGBA, title string, shadow bool) *image.RGBA {
	const radius = 12
	pad := fs.cellW * 2
	bar := fs.cellH + fs.cellW
	margin := 0
	if shadow {
		margin = 56
	}
	cw, ch := content.Bounds().Dx(), content.Bounds().Dy()
	ww, wh := cw+2*pad, ch+bar+pad
	img := image.NewRGBA(image.Rect(0, 0, ww+2*margin, wh+2*margin))

	// Signed distance to the window's rounded rectangle (negative inside).
	dist := func(x, y float64) float64 {
		cx := math.Max(math.Abs(x-float64(margin)-float64(ww)/2)-float64(ww)/2+radius, 0)
		cy := math.Max(math.Abs(y-float64(margin)-float64(wh)/2)-float64(wh)/2+radius, 0)
		return math.Hypot(cx, cy) - radius
	}
	titleBg := color.RGBA{0x16, 0x1B, 0x22, 0xFF}
	border := color.RGBA{0x30, 0x36, 0x3D, 0xFF}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			d := dist(float64(x)+0.5, float64(y)+0.5)
			switch {
			case d > 0 && shadow:
				ds := dist(float64(x)+0.5, float64(y)+0.5-14) // shadow offset down
				a := 0.0
				if ds > 0 {
					a = 0.42 * math.Exp(-(ds*ds)/(2*20*20))
				} else {
					a = 0.42
				}
				img.SetRGBA(x, y, color.RGBA{0, 0, 0, uint8(a * 255)})
			case d > 0:
				img.SetRGBA(x, y, termBg)
			case d > -1:
				img.SetRGBA(x, y, border)
			case y-margin < bar:
				img.SetRGBA(x, y, titleBg)
			default:
				img.SetRGBA(x, y, termBg)
			}
		}
	}
	fillRect(img, margin+1, margin+bar, margin+ww-1, margin+bar+1, border)

	// Traffic lights.
	lights := []color.RGBA{{0xFF, 0x5F, 0x57, 0xFF}, {0xFE, 0xBC, 0x2E, 0xFF}, {0x28, 0xC8, 0x40, 0xFF}}
	lr := float64(fs.cellW) * 0.55
	for i, c := range lights {
		cx := float64(margin) + float64(pad) + float64(i)*lr*3.4
		cy := float64(margin) + float64(bar)/2
		for y := int(cy - lr - 1); y <= int(cy+lr+1); y++ {
			for x := int(cx - lr - 1); x <= int(cx+lr+1); x++ {
				if a := lr + 0.5 - math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy); a > 0 {
					img.SetRGBA(x, y, mix(titleBg, c, math.Min(a, 1)))
				}
			}
		}
	}

	// Title.
	f := fs.faceFor('a', false)
	tw := font.MeasureString(f, title)
	m := f.Metrics()
	d := font.Drawer{Dst: img, Src: image.NewUniform(color.RGBA{0x8B, 0x94, 0x9E, 0xFF}), Face: f,
		Dot: fixed.Point26_6{
			X: fixed.I(margin) + (fixed.I(ww)-tw)/2,
			Y: fixed.I(margin) + (fixed.I(bar)-(m.Ascent+m.Descent))/2 + m.Ascent,
		}}
	d.DrawString(title)

	ox, oy := margin+pad, margin+bar+pad/2
	for y := range ch {
		for x := range cw {
			img.SetRGBA(ox+x, oy+y, content.RGBAAt(x, y))
		}
	}
	return img
}
