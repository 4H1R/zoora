package domain

import "testing"

func TestMaxScorePerQuestionType(t *testing.T) {
	cases := []struct {
		name string
		q    Question
		want float64
	}{
		{
			"short answer takes highest variant",
			Question{Type: QuestionTypeShortAnswer, Options: []QuestionOption{
				{ID: "a", Value: "x", Score: 2},
				{ID: "b", Value: "y", Score: 3},
			}},
			3,
		},
		{
			"descriptive sums rubric concept weights",
			Question{Type: QuestionTypeDescriptive, Options: []QuestionOption{
				{ID: "a", Value: "concept one", Score: 2},
				{ID: "b", Value: "concept two", Score: 1.5},
			}},
			3.5,
		},
		{
			"descriptive single score-holder option unchanged",
			Question{Type: QuestionTypeDescriptive, Options: []QuestionOption{
				{ID: "a", Score: 4},
			}},
			4,
		},
		{
			"multi-select choice sums positives",
			Question{Type: QuestionTypeChoice, Options: []QuestionOption{
				{ID: "a", Value: "x", Score: 1},
				{ID: "b", Value: "y", Score: 2},
				{ID: "c", Value: "z", Score: 0},
			}},
			3,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.q.MaxScore(); got != c.want {
				t.Fatalf("MaxScore() = %v, want %v", got, c.want)
			}
		})
	}
}
