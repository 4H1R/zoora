package imaging

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func TestRenderTextProducesValidPNG(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"english", "What is the capital of France?"},
		{"persian", "پایتخت ایران کجاست؟"},
		{"mixed", "Answer: بله"},
		{"digits", "2 + 2 = ?"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := RenderText(tc.text, Options{Noise: NoiseMedium, Seed: 1})
			if err != nil {
				t.Fatalf("RenderText: %v", err)
			}
			img, err := png.Decode(bytes.NewReader(out))
			if err != nil {
				t.Fatalf("decoding png: %v", err)
			}
			if b := img.Bounds(); b.Dx() < 10 || b.Dy() < 10 {
				t.Fatalf("image too small: %v", b)
			}
		})
	}
}

func TestRenderTextDeterministic(t *testing.T) {
	a, err := RenderText("سلام دنیا", Options{Noise: NoiseMedium, Seed: 42})
	if err != nil {
		t.Fatal(err)
	}
	b, err := RenderText("سلام دنیا", Options{Noise: NoiseMedium, Seed: 42})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("same text + seed must be byte-identical")
	}
}

func TestRenderTextEmpty(t *testing.T) {
	if _, err := RenderText("", Options{}); err == nil {
		t.Fatal("expected error for empty text")
	}
}

// TestRenderTextWraps verifies long text wraps instead of producing one giant
// horizontal strip: the wrapped image is taller and no wider than the limit.
func TestRenderTextWraps(t *testing.T) {
	long := "The quick brown fox jumps over the lazy dog and then keeps on running across the entire field without stopping once"
	const maxW = 800

	wrapped, err := RenderText(long, Options{MaxWidthPx: maxW, Seed: 1})
	if err != nil {
		t.Fatalf("RenderText wrapped: %v", err)
	}
	wImg := decode(t, wrapped)
	if w := wImg.Bounds().Dx(); w > maxW {
		t.Fatalf("wrapped image width %d exceeds max %d", w, maxW)
	}

	// A very wide limit keeps it on one line — must be shorter (fewer lines).
	oneLine, err := RenderText(long, Options{MaxWidthPx: 100000, Seed: 1})
	if err != nil {
		t.Fatalf("RenderText one-line: %v", err)
	}
	if wImg.Bounds().Dy() <= decode(t, oneLine).Bounds().Dy() {
		t.Fatal("wrapped image should be taller than the single-line render")
	}
}

func TestRenderTextNewlineBreaks(t *testing.T) {
	out, err := RenderText("line one\nline two", Options{MaxWidthPx: 100000, Seed: 1})
	if err != nil {
		t.Fatalf("RenderText: %v", err)
	}
	// Two forced lines are taller than a single line of the same words.
	single, _ := RenderText("line one line two", Options{MaxWidthPx: 100000, Seed: 1})
	if decode(t, out).Bounds().Dy() <= decode(t, single).Bounds().Dy() {
		t.Fatal("explicit newline should add a line")
	}
}

func TestRenderTextHighNoiseDeterministic(t *testing.T) {
	a, err := RenderText("What is the capital of France?", Options{Noise: NoiseHigh, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	b, err := RenderText("What is the capital of France?", Options{Noise: NoiseHigh, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("high-noise render must be deterministic per seed")
	}
}

func decode(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decoding png: %v", err)
	}
	return img
}
