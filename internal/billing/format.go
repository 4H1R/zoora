package billing

import (
	"strconv"
	"strings"
	"time"
)

var persianDigits = []rune{'۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'}

// toPersianDigits rewrites ASCII digits (0-9) as their Persian counterparts,
// leaving all other runes untouched.
func toPersianDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(persianDigits[r-'0'])
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// groupThousands inserts the Persian thousands separator (٬) every 3 digits.
func groupThousands(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	out := strings.Join(parts, "٬")
	if neg {
		out = "-" + out
	}
	return out
}

// formatTomanFa converts Rial minor units to Toman (÷10), groups thousands, and
// renders the result with Persian digits.
func formatTomanFa(rial int64) string {
	toman := rial / 10
	return toPersianDigits(groupThousands(toman))
}

// formatJalaliFa renders a Gregorian instant as a Jalali date (YYYY/MM/DD) in
// Persian digits.
func formatJalaliFa(t time.Time) string {
	jy, jm, jd := toJalali(t.Year(), int(t.Month()), t.Day())
	s := strconv.Itoa(jy) + "/" + pad2(jm) + "/" + pad2(jd)
	return toPersianDigits(s)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
