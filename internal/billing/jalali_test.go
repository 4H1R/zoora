package billing

import (
	"testing"
	"time"
)

func TestToJalaliYear(t *testing.T) {
	// 2026-07-08 Gregorian = 1405-04-17 Jalali.
	jy, jm, jd := toJalali(2026, 7, 8)
	if jy != 1405 || jm != 4 || jd != 17 {
		t.Fatalf("got %d-%02d-%02d, want 1405-04-17", jy, jm, jd)
	}
	if got := jalaliYearPrefix(time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)); got != "1405" {
		t.Errorf("prefix = %s, want 1405", got)
	}
}
