package billing

import (
	"strconv"
	"time"
)

// jalaliYearPrefix returns the Jalali (Solar Hijri) year as a string, e.g.
// "1405", for invoice numbering. Uses the standard algorithm; good enough for a
// year prefix. Phase 5 adds full Jalali date formatting for the PDF.
func jalaliYearPrefix(t time.Time) string {
	jy, _, _ := toJalali(t.Year(), int(t.Month()), t.Day())
	return strconv.Itoa(jy)
}

// toJalali converts a Gregorian date to Jalali (algorithm by Kazimierz Borkowski).
func toJalali(gy, gm, gd int) (jy, jm, jd int) {
	gDaysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	jDaysInMonth := []int{31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29}
	gy2 := gy - 1600
	gm2 := gm - 1
	gd2 := gd - 1
	gDayNo := 365*gy2 + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400
	for i := range gm2 {
		gDayNo += gDaysInMonth[i]
	}
	if gm2 > 1 && ((gy%4 == 0 && gy%100 != 0) || gy%400 == 0) {
		gDayNo++
	}
	gDayNo += gd2
	jDayNo := gDayNo - 79
	jNp := jDayNo / 12053
	jDayNo %= 12053
	jy = 979 + 33*jNp + 4*(jDayNo/1461)
	jDayNo %= 1461
	if jDayNo >= 366 {
		jy += (jDayNo - 1) / 365
		jDayNo = (jDayNo - 1) % 365
	}
	var i int
	for i = 0; i < 11 && jDayNo >= jDaysInMonth[i]; i++ {
		jDayNo -= jDaysInMonth[i]
	}
	jm = i + 1
	jd = jDayNo + 1
	return jy, jm, jd
}
