package quizzes

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// normalizeText canonicalizes free-text answers for comparison: Unicode NFKC,
// Arabic→Persian letter folding (ي→ی ك→ک ة→ه), Persian/Arabic-Indic digits→ASCII,
// Arabic diacritics, tatweel and zero-width characters stripped, lowercased,
// whitespace collapsed. Used by short-answer grading and descriptive suggestions
// so FA/EN spelling variants of the same answer compare equal.
func normalizeText(s string) string {
	s = norm.NFKC.String(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == 'ي' || r == 'ى': // Arabic yeh / alef maksura
			r = 'ی'
		case r == 'ك': // Arabic kaf
			r = 'ک'
		case r == 'ة': // the marbuta
			r = 'ه'
		case r >= '۰' && r <= '۹': // Extended (Persian) digits
			r = '0' + (r - '۰')
		case r >= '٠' && r <= '٩': // Arabic-Indic digits
			r = '0' + (r - '٠')
		case r >= 0x064B && r <= 0x065F: // Arabic diacritics (fatha, kasra, shadda, ...)
			continue
		case r == 0x0670: // superscript alef
			continue
		case r == 0x0640: // tatweel (kashida)
			continue
		case r >= 0x200B && r <= 0x200F: // zero-width chars incl. ZWNJ/ZWJ
			continue
		case r == 0xFEFF: // BOM
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// normalizeCompact is normalizeText with all spaces removed, so "می روم",
// "می‌روم" and "میروم" compare equal. Used as the second, spacing-insensitive
// matching pass.
func normalizeCompact(s string) string {
	return strings.ReplaceAll(normalizeText(s), " ", "")
}

// isNumericAnswer reports whether a normalized answer is purely numeric
// (digits plus separators). Numeric answers skip the spacing-insensitive pass:
// "1 5" must not match "15".
func isNumericAnswer(s string) bool {
	s = normalizeText(s)
	hasDigit := false
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			hasDigit = true
		case r == '.' || r == ',' || r == '/' || r == '-' || r == '%' || r == ' ':
		default:
			return false
		}
	}
	return hasDigit
}
