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
	waveWarp(img, strength, rng)
	speckle(img, strength, rng)
	if strength >= int(NoiseMedium) {
		strokes(img, rng)
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
	for i := 0; i < n; i++ {
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
	for i := 0; i < 3; i++ {
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
