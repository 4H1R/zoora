// Package imaging renders short pieces of text (English or Persian) to PNG
// images, optionally distorted so OCR / multimodal models have a harder time
// lifting the raw string while a human can still read it. It backs the exam
// anti-cheat feature: question and option text is rendered server-side and the
// student is served only the image (the text never reaches the client).
//
// Persian requires complex text shaping (contextual Arabic letter joining) and
// right-to-left layout, which the standard library cannot do; shaping is handled
// by go-text/typesetting's HarfBuzz port. The embedded Vazirmatn font covers
// both Latin and Arabic script.
package imaging

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"sync"
	"unicode"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

//go:embed fonts/Vazirmatn-Regular.ttf
var vazirmatnTTF []byte

// face is parsed once and reused. go-text lazily populates per-face cmap/glyph
// caches during shaping AND rasterization, and those writes are NOT
// goroutine-safe, so all face-touching work is serialized by renderMu. Only the
// cheap shared-nothing steps (obfuscation, PNG encode) run outside the lock, so
// concurrent RenderText callers still overlap there and on their I/O.
var (
	faceOnce sync.Once
	faceInst *font.Face
	faceErr  error
	renderMu sync.Mutex
)

func loadFace() (*font.Face, error) {
	faceOnce.Do(func() {
		faceInst, faceErr = font.ParseTTF(bytes.NewReader(vazirmatnTTF))
	})
	return faceInst, faceErr
}

// NoiseLevel controls how much anti-OCR distortion is applied.
type NoiseLevel int

const (
	NoiseNone   NoiseLevel = 0 // clean render, no distortion
	NoiseLight  NoiseLevel = 1 // subtle wave + speckle
	NoiseMedium NoiseLevel = 2 // wave + speckle + cross-through strokes
	NoiseHigh   NoiseLevel = 3 // + per-glyph jitter/rotation and 2D local warp
)

// Options configures a single render.
type Options struct {
	// FontSizePx is the glyph size in pixels. Zero falls back to DefaultFontSize.
	FontSizePx float64
	// MaxWidthPx is the max image width before text wraps to a new line. Zero
	// falls back to DefaultMaxWidth. A single word wider than the limit still
	// overflows (no mid-word breaking).
	MaxWidthPx float64
	// Noise selects the obfuscation strength.
	Noise NoiseLevel
	// Seed makes the (random) distortion deterministic. The same text + seed
	// always yields byte-identical output, so re-renders are stable and callers
	// can derive the seed from a stable id (e.g. a question/option UUID).
	Seed int64
}

const (
	// DefaultFontSize is used when Options.FontSizePx is zero.
	DefaultFontSize = 72
	// DefaultMaxWidth is used when Options.MaxWidthPx is zero.
	DefaultMaxWidth = 1200
	pad             = 40 // px of padding around the text
)

// RenderText shapes text and returns a PNG. Direction and script are detected
// automatically: Arabic-script-dominant text is laid out right-to-left. Text
// wider than MaxWidthPx wraps onto multiple lines.
func RenderText(text string, opts Options) ([]byte, error) {
	face, err := loadFace()
	if err != nil {
		return nil, fmt.Errorf("loading font: %w", err)
	}
	size := opts.FontSizePx
	if size <= 0 {
		size = DefaultFontSize
	}
	maxWidth := opts.MaxWidthPx
	if maxWidth <= 0 {
		maxWidth = DefaultMaxWidth
	}

	// One rng drives both the per-glyph jitter (during rasterization) and the
	// image-space obfuscation, so the whole render stays deterministic per seed.
	rng := rand.New(rand.NewSource(opts.Seed))

	// Serialize the face-touching render; see renderMu.
	renderMu.Lock()
	img, err := renderText(face, []rune(text), size, maxWidth, int(opts.Noise), rng)
	renderMu.Unlock()
	if err != nil {
		return nil, err
	}

	obfuscate(img, int(opts.Noise), rng)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encoding png: %w", err)
	}
	return buf.Bytes(), nil
}

// isRTL reports whether the text is dominated by Arabic-script (Persian) runes,
// in which case it must be shaped right-to-left. Good enough for exam questions,
// which are single-language.
func isRTL(rs []rune) bool {
	arabic := 0
	for _, r := range rs {
		if unicode.Is(unicode.Arabic, r) {
			arabic++
		}
	}
	return arabic*2 > len(rs)
}

// shapedLine is one wrapped line: its shaped glyphs and pixel advance.
type shapedLine struct {
	out     shaping.Output
	advance float64
}

// renderText wraps the text to maxWidthPx, shapes each line, and rasterizes the
// glyph outlines onto a white RGBA image, returning the result plus padding. At
// NoiseHigh each glyph is given a small deterministic vertical/rotational jitter
// to break the uniform baseline and glyph geometry OCR relies on.
func renderText(face *font.Face, text []rune, sizePx, maxWidthPx float64, strength int, rng *rand.Rand) (*image.RGBA, error) {
	if len(text) == 0 {
		return nil, fmt.Errorf("empty text")
	}
	dir := di.DirectionLTR
	script := language.Latin
	lang := language.NewLanguage("en")
	if isRTL(text) {
		dir = di.DirectionRTL
		script = language.Arabic
		lang = language.NewLanguage("fa")
	}

	shaper := &shaping.HarfbuzzShaper{}
	shape := func(rs []rune) shaping.Output {
		return shaper.Shape(shaping.Input{
			Text:      rs,
			RunStart:  0,
			RunEnd:    len(rs),
			Direction: dir,
			Face:      face,
			Size:      fixed.Int26_6(sizePx * 64),
			Script:    script,
			Language:  lang,
		})
	}

	// Wrap into lines that each fit within the usable width, then shape each.
	usable := maxWidthPx - 2*pad
	lines := make([]shapedLine, 0, 4)
	maxAdvance := 0.0
	for _, lr := range wrapLines(shape, text, usable) {
		out := shape(lr)
		adv := f2f(out.Advance)
		if adv > maxAdvance {
			maxAdvance = adv
		}
		lines = append(lines, shapedLine{out: out, advance: adv})
	}

	// Line metrics are font-consistent across lines; take them from the first.
	ascent := f2f(lines[0].out.LineBounds.Ascent)
	descent := f2f(lines[0].out.LineBounds.Descent) // negative, points down
	lineHeight := ascent - descent
	leading := lineHeight * 0.2

	w := int(math.Ceil(maxAdvance)) + 2*pad
	h := int(math.Ceil(float64(len(lines))*lineHeight+float64(len(lines)-1)*leading)) + 2*pad
	if w < 1 || h < 1 {
		return nil, fmt.Errorf("nothing to render (advance=%.1f)", maxAdvance)
	}

	// Outlines are in font units; scale them to pixels.
	scale := sizePx / float64(face.Upem())

	// Per-glyph jitter magnitudes (0 unless NoiseHigh).
	jitterY, jitterRot := 0.0, 0.0
	if strength >= int(NoiseHigh) {
		jitterY = sizePx * 0.06
		jitterRot = 0.10
	}

	ras := vector.NewRasterizer(w, h)
	for li := range lines {
		ln := lines[li]
		baseline := float64(pad) + ascent + float64(li)*(lineHeight+leading)
		// RTL lines are right-aligned; LTR lines start at the left padding.
		penX := float64(pad)
		if dir == di.DirectionRTL {
			penX = float64(w) - float64(pad) - ln.advance
		}
		for _, g := range ln.out.Glyphs {
			gx := penX + f2f(g.XOffset)
			gy := baseline - f2f(g.YOffset)
			if jitterY != 0 {
				gy += (rng.Float64()*2 - 1) * jitterY
			}
			rot := 0.0
			if jitterRot != 0 {
				rot = (rng.Float64()*2 - 1) * jitterRot
			}
			appendGlyph(ras, face, g.GlyphID, gx, gy, scale, rot)
			penX += f2f(g.XAdvance) //nolint:staticcheck // horizontal-only; XAdvance is the pen advance here
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 0xff // white background
	}
	// Faint decoy words/numbers go down first, so the real text is drawn on top
	// and covers any decoy pixels sharing its space — decoys survive only in the
	// whitespace. This is a MILD deterrent (kills copy-paste, DOM-scrape and weak
	// OCR); it does not beat strong multimodal models, so it is kept legible.
	if strength >= int(NoiseHigh) {
		drawDecoysBehind(img, face, script, lang, dir, w, h, sizePx, scale, rng)
	}
	ras.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{20, 20, 24, 255}), image.Point{})
	return img, nil
}

// wrapLines greedily breaks text into lines no wider than maxAdvance px, using
// the shaper to measure each candidate. Explicit newlines force a break; blank
// lines are preserved. A word wider than maxAdvance is emitted on its own line
// and allowed to overflow (no mid-word breaking). Falls back to one line if
// maxAdvance is non-positive.
func wrapLines(shape func([]rune) shaping.Output, text []rune, maxAdvance float64) [][]rune {
	var lines [][]rune
	for _, para := range splitRunes(text, '\n') {
		if maxAdvance <= 0 {
			lines = append(lines, para)
			continue
		}
		words := splitRunes(para, ' ')
		var cur []rune
		for _, word := range words {
			if len(word) == 0 {
				continue // collapse runs of spaces
			}
			candidate := word
			if len(cur) > 0 {
				candidate = append(append(append([]rune{}, cur...), ' '), word...)
			}
			if len(cur) > 0 && f2f(shape(candidate).Advance) > maxAdvance {
				lines = append(lines, cur)
				cur = append([]rune{}, word...)
			} else {
				cur = candidate
			}
		}
		lines = append(lines, cur) // may be empty for a blank paragraph
	}
	if len(lines) == 0 {
		lines = append(lines, text)
	}
	return lines
}

// splitRunes splits a rune slice on sep, keeping empty segments.
func splitRunes(rs []rune, sep rune) [][]rune {
	var out [][]rune
	start := 0
	for i, r := range rs {
		if r == sep {
			out = append(out, rs[start:i])
			start = i + 1
		}
	}
	return append(out, rs[start:])
}

// appendGlyph translates one glyph's outline into rasterizer path ops, applying
// an optional rotation (radians) about the glyph origin. Font space has Y up;
// image space has Y down, so Y is negated about the baseline.
func appendGlyph(ras *vector.Rasterizer, face *font.Face, gid font.GID, ox, oy, scale, rot float64) {
	data := face.GlyphData(gid)
	outline, ok := data.(font.GlyphOutline)
	if !ok {
		return // bitmap/color/svg glyph — skip
	}
	sin, cos := math.Sin(rot), math.Cos(rot)
	tx := func(p font.SegmentPoint) (float32, float32) {
		dx := float64(p.X) * scale
		dy := -float64(p.Y) * scale
		return float32(ox + dx*cos - dy*sin), float32(oy + dx*sin + dy*cos)
	}
	for _, s := range outline.Segments {
		switch s.Op {
		case ot.SegmentOpMoveTo:
			x, y := tx(s.Args[0])
			ras.MoveTo(x, y)
		case ot.SegmentOpLineTo:
			x, y := tx(s.Args[0])
			ras.LineTo(x, y)
		case ot.SegmentOpQuadTo:
			cx, cy := tx(s.Args[0])
			x, y := tx(s.Args[1])
			ras.QuadTo(cx, cy, x, y)
		case ot.SegmentOpCubeTo:
			c1x, c1y := tx(s.Args[0])
			c2x, c2y := tx(s.Args[1])
			x, y := tx(s.Args[2])
			ras.CubeTo(c1x, c1y, c2x, c2y, x, y)
		}
	}
	ras.ClosePath()
}

func f2f(v fixed.Int26_6) float64 { return float64(v) / 64 }
