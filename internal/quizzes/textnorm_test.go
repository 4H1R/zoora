package quizzes

import "testing"

func TestNormalizeText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"english lowercase trim collapse", "  Hello   World ", "hello world"},
		{"arabic yeh folded to persian", "عربي", "عربی"},
		{"arabic kaf folded to persian", "كتاب", "کتاب"},
		{"teh marbuta folded to heh", "مدرسة", "مدرسه"},
		{"persian digits to latin", "۱۲۳", "123"},
		{"arabic-indic digits to latin", "١٢٣", "123"},
		{"diacritics stripped", "مَدْرَسَه", "مدرسه"},
		{"zwnj removed", "می‌روم", "میروم"},
		{"fullwidth via nfkc", "Ｈｅｌｌｏ", "hello"},
		{"arabic presentation form via nfkc", "ﻻ", "لا"},
		{"mixed", "  پاسخ: ۴۲ درجه ", "پاسخ: 42 درجه"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := normalizeText(c.in); got != c.want {
				t.Fatalf("normalizeText(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeCompact(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"spaces stripped", "می روم", "میروم"},
		{"zwnj and spaces equal attached", "می‌روم", "میروم"},
		{"english spaces stripped", "ice  cream", "icecream"},
		{"digits keep order", "1 5", "15"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := normalizeCompact(c.in); got != c.want {
				t.Fatalf("normalizeCompact(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestIsNumericAnswer(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"15", true},
		{"۱۵", true},
		{"1 5", true},
		{"1.5", true},
		{"-3", true},
		{"3/4", true},
		{"42 درجه", false},
		{"abc", false},
		{"", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := isNumericAnswer(c.in); got != c.want {
				t.Fatalf("isNumericAnswer(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
