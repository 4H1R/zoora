package imaging

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

// Decoy words in the two scripts we render. They are deliberately plausible
// exam-answer vocabulary in the SAME language as the question, so an OCR /
// multimodal model cannot tell a decoy from the real text by language alone
// (a foreign-script watermark would be trivial to filter out).
var (
	decoyWordsFa = []string{
		"پاسخ", "سوال", "درست", "غلط", "گزینه", "عدد", "مقدار", "معادله",
		"جواب", "شکل", "نمودار", "جدول", "فرمول", "نتیجه", "تقریبا", "بیشتر",
		"کمتر", "برابر", "شامل", "زیرا", "بنابراین", "همیشه", "هرگز", "شاید",
		"مثبت", "منفی", "متغیر", "ثابت", "واحد", "درصد",
	}
	decoyWordsEn = []string{
		"answer", "question", "true", "false", "option", "value", "equals",
		"result", "figure", "table", "formula", "always", "never", "maybe",
		"because", "therefore", "positive", "negative", "constant", "unit",
		"percent", "approx", "greater", "less", "sum", "total", "select",
		"correct", "wrong", "none",
	}
)

// drawDecoysBehind paints a field of scattered decoy words and numbers BEHIND
// the real text (drawn before it). The real near-black glyphs are drawn on top
// and cover any overlap, so decoys survive only in the whitespace, staying
// human-legible. This is a MILD deterrent — it degrades copy-paste, DOM-scrape
// and weak OCR, but does NOT beat strong multimodal models, so it deliberately
// stays faint rather than shredding legibility for a fight it can't win.
func drawDecoysBehind(img *image.RGBA, face *font.Face, script language.Script, lang language.Language, dir di.Direction, w, h int, sizePx, scale float64, rng *rand.Rand) {
	// Roughly one decoy per ~20k px², several at minimum, capped so dense wrapped
	// blocks do not turn into mush.
	n := min(w*h/20000+6, 80)
	placeDecoys(img, face, script, lang, dir, w, h, sizePx, scale, n, 0.30, 0.40, 150, 70, rng)
}

// placeDecoys shapes and stamps n decoy tokens at random positions, sizes,
// rotations and gray tones. sizeLo/sizeSpan are fractions of the real font size;
// toneLo/toneSpan bound the gray value (higher = lighter). It is face-touching
// (shapes glyphs), so it must run under renderMu like the rest of renderText.
func placeDecoys(img *image.RGBA, face *font.Face, script language.Script, lang language.Language, dir di.Direction, w, h int, sizePx, scale float64, n int, sizeLo, sizeSpan float64, toneLo, toneSpan int, rng *rand.Rand) {
	words := decoyWordsEn
	if dir == di.DirectionRTL {
		words = decoyWordsFa
	}

	shaper := &shaping.HarfbuzzShaper{}

	for range n {
		token := decoyToken(words, rng)
		dsize := sizePx * (sizeLo + rng.Float64()*sizeSpan)
		dscale := scale * (dsize / sizePx)

		out := shaper.Shape(shaping.Input{
			Text:      token,
			RunStart:  0,
			RunEnd:    len(token),
			Direction: dir,
			Face:      face,
			Size:      fixed.Int26_6(dsize * 64),
			Script:    script,
			Language:  lang,
		})

		ax := rng.Float64() * float64(w)
		ay := rng.Float64() * float64(h)
		rot := (rng.Float64()*2 - 1) * 0.6 // ±0.6 rad
		sin, cos := math.Sin(rot), math.Cos(rot)

		ras := vector.NewRasterizer(w, h)
		penX := 0.0
		for _, g := range out.Glyphs {
			// Rotate the whole word rigidly: place each glyph origin along the
			// unrotated advance, then rotate that offset about the word anchor
			// (appendGlyph also rotates the outline by the same angle).
			lx := penX + f2f(g.XOffset)
			ly := -f2f(g.YOffset)
			gx := ax + lx*cos - ly*sin
			gy := ay + lx*sin + ly*cos
			appendGlyph(ras, face, g.GlyphID, gx, gy, dscale, rot)
			penX += f2f(g.XAdvance) //nolint:staticcheck // horizontal-only; XAdvance is the pen advance here
		}

		// Per-decoy gray tone; varying it per word means no single brightness cut
		// removes every decoy at once.
		v := uint8(toneLo + rng.Intn(toneSpan))
		b := v
		if int(v)+6 <= 255 {
			b = v + 6
		}
		ras.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{v, v, b, 0xff}), image.Point{})
	}
}

// decoyToken returns either a random decoy word or a random 2–4 digit number
// (ASCII or Persian digits, chosen at random). Numbers dominate (~55%): a stray
// decoy figure makes any AI-extracted numeric answer untrustworthy, so they are
// the highest-value distractor.
func decoyToken(words []string, rng *rand.Rand) []rune {
	if rng.Intn(20) < 11 {
		return decoyNumber(rng)
	}
	return []rune(words[rng.Intn(len(words))])
}

func decoyNumber(rng *rand.Rand) []rune {
	ndigits := 2 + rng.Intn(3) // 2–4 digits
	persian := rng.Intn(2) == 0
	out := make([]rune, ndigits)
	for i := range out {
		d := rune('0' + rng.Intn(10))
		if persian {
			d = '۰' + rune(rng.Intn(10)) // U+06F0..U+06F9
		}
		out[i] = d
	}
	return out
}
