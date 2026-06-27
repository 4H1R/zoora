package quizzes

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

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
