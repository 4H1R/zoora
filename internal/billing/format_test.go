package billing

import (
	"testing"
	"time"
)

func mustDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestFormatToman(t *testing.T) {
	// 15,000,000 Rial = 1,500,000 Toman, grouped, Persian digits.
	if got := formatTomanFa(15_000_000); got != "۱٬۵۰۰٬۰۰۰" {
		t.Errorf("formatTomanFa = %q, want ۱٬۵۰۰٬۰۰۰", got)
	}
}

func TestFormatJalaliDate(t *testing.T) {
	// 2026-07-08 => 1405/04/17 in Persian digits.
	got := formatJalaliFa(mustDate("2026-07-08"))
	if got != "۱۴۰۵/۰۴/۱۷" {
		t.Errorf("formatJalaliFa = %q, want ۱۴۰۵/۰۴/۱۷", got)
	}
}

func TestToPersianDigits(t *testing.T) {
	if got := toPersianDigits("1405-000123"); got != "۱۴۰۵-۰۰۰۱۲۳" {
		t.Errorf("toPersianDigits = %q", got)
	}
}
