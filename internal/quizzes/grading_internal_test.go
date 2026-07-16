package quizzes

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

func TestGradeShortAnswer(t *testing.T) {
	opts := func(value float64, v string) []domain.QuestionOption {
		return []domain.QuestionOption{{ID: "a1", Value: v, Score: value}}
	}
	cases := []struct {
		name    string
		options []domain.QuestionOption
		value   string
		want    float64
	}{
		{"exact match", opts(2, "photosynthesis"), "Photosynthesis", 2},
		{"whitespace and case", opts(2, "ice cream"), "  ICE   Cream ", 2},
		{"arabic yeh vs persian yeh", opts(3, "علی"), "علي", 3},
		{"arabic kaf", opts(3, "کتاب"), "كتاب", 3},
		{"persian digits vs latin", opts(1, "۱۵"), "15", 1},
		{"diacritics ignored", opts(2, "مدرسه"), "مَدْرَسَة", 2},
		{"zwnj vs attached", opts(2, "می‌روم"), "میروم", 2},
		{"zwnj vs spaced (pass 2)", opts(2, "می‌روم"), "می روم", 2},
		{"spaced vs attached english (pass 2)", opts(1, "ice cream"), "icecream", 1},
		{"numeric guard: spaced digits stay distinct", opts(1, "15"), "1 5", 0},
		{"numeric exact still works", opts(1, "15"), "۱۵", 1},
		{"wrong answer", opts(2, "photosynthesis"), "respiration", 0},
		{"zero-score option never matches", []domain.QuestionOption{{ID: "a1", Value: "x", Score: 0}}, "x", 0},
		{"empty student answer", opts(2, "photosynthesis"), "", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := gradeShortAnswer(c.options, c.value); got != c.want {
				t.Fatalf("gradeShortAnswer = %v, want %v", got, c.want)
			}
		})
	}
}

func TestGradeChoiceWithNegative(t *testing.T) {
	opts := []domain.QuestionOption{
		{ID: "c1", Value: "correct", Score: 1},
		{ID: "c2", Value: "correct2", Score: 1},
		{ID: "w1", Value: "wrong", Score: 0},
		{ID: "w2", Value: "wrong2", Score: -3}, // magnitude ignored; only sign matters
	}
	cases := []struct {
		name     string
		selected []string
		cfg      domain.NegativeMarkConfig
		want     float64
	}{
		{"none mode", []string{"c1", "w1"}, domain.NegativeMarkConfig{Mode: domain.NegativeMarkNone}, 1},
		{"per_wrong 2 wrong", []string{"c1", "w1", "w2"}, domain.NegativeMarkConfig{Mode: domain.NegativeMarkPerWrong, NegativeValue: 0.5}, 1 - 1.0},
		{"accumulative 2/2", []string{"w1", "w2"}, domain.NegativeMarkConfig{Mode: domain.NegativeMarkAccumulative, WrongsPerPoint: 2}, -1},
		{"can go negative", []string{"w1", "w2"}, domain.NegativeMarkConfig{Mode: domain.NegativeMarkPerWrong, NegativeValue: 1}, -2},
		{"both correct no penalty", []string{"c1", "c2"}, domain.NegativeMarkConfig{Mode: domain.NegativeMarkPerWrong, NegativeValue: 1}, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := gradeChoice(opts, c.selected, c.cfg); got != c.want {
				t.Fatalf("gradeChoice = %v, want %v", got, c.want)
			}
		})
	}
}
