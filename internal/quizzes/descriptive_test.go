package quizzes

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

func TestStripHidesAIFieldsFromStudents(t *testing.T) {
	score := 4.0
	sub := &domain.QuizSubmission{Answers: []domain.SubmissionAnswer{{
		SuggestedScore: &score,
		AIRationale:    "secret hint",
		AIStatus:       domain.AIAnswerStatusScored,
	}}}
	stripSimilarity(sub) // existing student-facing strip
	a := sub.Answers[0]
	if a.SuggestedScore != nil || a.AIRationale != "" {
		t.Fatal("AI suggestion/rationale must be stripped from student reads")
	}
}
