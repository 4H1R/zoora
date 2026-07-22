package imaging

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// obfuscate applies image-space distortion. It is script-agnostic (works on the
// rendered pixels, not the glyphs), so it degrades OCR equally for English and
// Persian while leaving the text human-legible.
func obfuscate(img *image.RGBA, strength int, rng *rand.Rand) {
	if strength <= 0 {
		return
	}
	if strength >= int(NoiseHigh) {
		// A smooth 2D displacement field bends letters non-uniformly in both
		// axes — far harder for OCR segmentation than a single column sine,
		// while staying legible. Replaces the plain wave at this strength.
		localWarp(img, strength, rng)
	} else {
		waveWarp(img, strength, rng)
	}
	speckle(img, strength, rng)
	if strength >= int(NoiseMedium) {
		strokes(img, rng)
	}
}

// localWarp resamples the image through a smooth low-frequency displacement
// field built from a handful of superposed sines with random phases. Each pixel
// is pulled from a nearby source, so straight baselines and consistent glyph
// shapes are broken locally without shredding legibility.
func localWarp(img *image.RGBA, strength int, rng *rand.Rand) {
	b := img.Bounds()
	src := cloneRGBA(img)
	fill(img, color.RGBA{0xff, 0xff, 0xff, 0xff})
	amp := 2.0 + float64(strength)
	// Two independent wave sets per axis for a non-repeating field.
	type wave struct{ fx, fy, phase, weight float64 }
	mk := func() wave {
		return wave{
			fx:     (0.4 + rng.Float64()) / 90.0,
			fy:     (0.4 + rng.Float64()) / 90.0,
			phase:  rng.Float64() * 2 * math.Pi,
			weight: 0.5 + rng.Float64(),
		}
	}
	dxW := []wave{mk(), mk()}
	dyW := []wave{mk(), mk()}
	disp := func(ws []wave, x, y int) float64 {
		var v float64
		for _, w := range ws {
			v += w.weight * math.Sin(2*math.Pi*(float64(x)*w.fx+float64(y)*w.fy)+w.phase)
		}
		return v / float64(len(ws)) * amp
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			sx := x + int(disp(dxW, x, y))
			sy := y + int(disp(dyW, x, y))
			if sx < b.Min.X || sx >= b.Max.X || sy < b.Min.Y || sy >= b.Max.Y {
				continue
			}
			img.SetRGBA(x, y, src.RGBAAt(sx, sy))
		}
	}
}

// waveWarp displaces each column vertically along a sine, breaking the straight
// baselines and consistent glyph geometry OCR relies on.
func waveWarp(img *image.RGBA, strength int, rng *rand.Rand) {
	b := img.Bounds()
	amp := float64(strength) * 3.0
	period := 60.0 + rng.Float64()*40.0
	phase := rng.Float64() * 2 * math.Pi
	src := cloneRGBA(img)
	fill(img, color.RGBA{0xff, 0xff, 0xff, 0xff})
	for x := b.Min.X; x < b.Max.X; x++ {
		dy := int(amp * math.Sin(2*math.Pi*float64(x)/period+phase))
		for y := b.Min.Y; y < b.Max.Y; y++ {
			sy := y - dy
			if sy < b.Min.Y || sy >= b.Max.Y {
				continue
			}
			img.SetRGBA(x, y, src.RGBAAt(x, sy))
		}
	}
}

// speckle sprinkles random dark/light dots to add texture that confuses
// binarization and edge detection.
func speckle(img *image.RGBA, strength int, rng *rand.Rand) {
	b := img.Bounds()
	n := strength * (b.Dx() * b.Dy()) / 120
	for range n {
		x := b.Min.X + rng.Intn(b.Dx())
		y := b.Min.Y + rng.Intn(b.Dy())
		g := uint8(rng.Intn(90))
		if rng.Intn(2) == 0 {
			g = uint8(200 + rng.Intn(55))
		}
		img.SetRGBA(x, y, color.RGBA{g, g, g, 0xff})
	}
}

// strokes draws a few faint sinusoidal lines across the whole image — the
// CAPTCHA cross-through that ties glyphs together and defeats per-character
// segmentation.
func strokes(img *image.RGBA, rng *rand.Rand) {
	b := img.Bounds()
	for range 3 {
		amp := 4.0 + rng.Float64()*8.0
		period := 80.0 + rng.Float64()*80.0
		phase := rng.Float64() * 2 * math.Pi
		base := b.Min.Y + rng.Intn(b.Dy())
		col := color.RGBA{80, 80, 90, 255}
		for x := b.Min.X; x < b.Max.X; x++ {
			y := base + int(amp*math.Sin(2*math.Pi*float64(x)/period+phase))
			for dy := -1; dy <= 1; dy++ {
				yy := y + dy
				if yy >= b.Min.Y && yy < b.Max.Y {
					img.SetRGBA(x, yy, col)
				}
			}
		}
	}
}

func cloneRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

func fill(img *image.RGBA, c color.RGBA) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}
