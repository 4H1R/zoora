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

// face is parsed once and reused; go-text faces are safe for concurrent shaping.
var (
	faceOnce sync.Once
	faceInst *font.Face
	faceErr  error
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
)

// Options configures a single render.
type Options struct {
	// FontSizePx is the glyph size in pixels. Zero falls back to DefaultFontSize.
	FontSizePx float64
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
	pad             = 40 // px of padding around the text
)

// RenderText shapes text and returns a PNG. Direction and script are detected
// automatically: Arabic-script-dominant text is laid out right-to-left.
func RenderText(text string, opts Options) ([]byte, error) {
	face, err := loadFace()
	if err != nil {
		return nil, fmt.Errorf("loading font: %w", err)
	}
	size := opts.FontSizePx
	if size <= 0 {
		size = DefaultFontSize
	}

	img, err := renderText(face, []rune(text), size)
	if err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(opts.Seed))
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

// renderText shapes the text and rasterizes its glyph outlines onto a white
// RGBA image, returning the tightly-cropped result plus padding.
func renderText(face *font.Face, text []rune, sizePx float64) (*image.RGBA, error) {
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
	out := shaper.Shape(shaping.Input{
		Text:      text,
		RunStart:  0,
		RunEnd:    len(text),
		Direction: dir,
		Face:      face,
		Size:      fixed.Int26_6(sizePx * 64),
		Script:    script,
		Language:  lang,
	})

	// Metrics come back in pixels (fixed 26.6) at the requested size.
	ascent := f2f(out.LineBounds.Ascent)
	descent := f2f(out.LineBounds.Descent) // negative, points down
	advance := f2f(out.Advance)

	w := int(math.Ceil(advance)) + 2*pad
	h := int(math.Ceil(ascent-descent)) + 2*pad
	if w < 1 || h < 1 {
		return nil, fmt.Errorf("nothing to render (advance=%.1f)", advance)
	}

	// Outlines are in font units; scale them to pixels.
	scale := sizePx / float64(face.Upem())
	baseline := float64(pad) + ascent

	// Accumulate every glyph contour into one rasterizer, then draw once.
	ras := vector.NewRasterizer(w, h)
	penX := float64(pad)
	for _, g := range out.Glyphs {
		gx := penX + f2f(g.XOffset)
		gy := baseline - f2f(g.YOffset)
		appendGlyph(ras, face, g.GlyphID, gx, gy, scale)
		penX += f2f(g.XAdvance) //nolint:staticcheck // horizontal-only; XAdvance is the pen advance here
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 0xff // white background
	}
	ras.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{20, 20, 24, 255}), image.Point{})
	return img, nil
}

// appendGlyph translates one glyph's outline into rasterizer path ops. Font
// space has Y up; image space has Y down, so Y is negated about the baseline.
func appendGlyph(ras *vector.Rasterizer, face *font.Face, gid font.GID, ox, oy, scale float64) {
	data := face.GlyphData(gid)
	outline, ok := data.(font.GlyphOutline)
	if !ok {
		return // bitmap/color/svg glyph — skip
	}
	tx := func(p font.SegmentPoint) (float32, float32) {
		return float32(ox + float64(p.X)*scale), float32(oy - float64(p.Y)*scale)
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
