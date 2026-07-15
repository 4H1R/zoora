package imaging

import (
	"bytes"
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
