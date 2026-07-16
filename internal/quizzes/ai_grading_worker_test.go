package quizzes

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type stubLLM struct {
	responses []string // one per Generate call, in order
	calls     int
}

func (s *stubLLM) Generate(context.Context, domain.LLMRequest) (domain.LLMResponse, error) {
	r := s.responses[s.calls]
	s.calls++
	return domain.LLMResponse{Text: r, Model: "gemini-2.0-flash"}, nil
}

func TestGradeSubmissionAIApplyWritesScores(t *testing.T) {
	q1 := uuid.New()
	q2 := uuid.New()
	sub := &domain.QuizSubmission{
		ID:     uuid.New(),
		Status: domain.SubmissionStatusSubmitted,
		Answers: []domain.SubmissionAnswer{
			{QuestionID: q1, Value: "answer one"},
			{QuestionID: q2, Value: "answer two"},
		},
	}
	questions := []domain.Question{
		descQuestion(q1, "q1", "model1", 5),
		descQuestion(q2, "q2", "model2", 10),
	}
	// Batch returns q1 only; q2 must be retried per-answer.
	batch := `{"scores":[{"question_id":"` + q1.String() + `","score":4,"rationale":"خوب"}]}`
	retryQ2 := `{"scores":[{"question_id":"` + q2.String() + `","score":8,"rationale":"قابل قبول"}]}`
	llm := &stubLLM{responses: []string{batch, retryQ2}}

	updated, err := gradeAnswersAI(context.Background(), llm, sub, questions, domain.AIGradingModeApply, false, uuid.New())
	if err != nil {
		t.Fatalf("gradeAnswersAI: %v", err)
	}
	if updated.Answers[0].EarnedScore != 4 || updated.Answers[0].GradedBy != domain.GradedByAI {
		t.Fatalf("q1 not applied: %+v", updated.Answers[0])
	}
	if updated.Answers[1].EarnedScore != 8 || updated.Answers[1].AIStatus != domain.AIAnswerStatusScored {
		t.Fatalf("q2 retry not applied: %+v", updated.Answers[1])
	}
	if updated.TotalScore != 12 {
		t.Fatalf("total not recomputed, got %v", updated.TotalScore)
	}
}

func TestGradeAnswersAIMarksFailedWhenRetryFails(t *testing.T) {
	q1 := uuid.New()
	sub := &domain.QuizSubmission{
		ID:      uuid.New(),
		Status:  domain.SubmissionStatusSubmitted,
		Answers: []domain.SubmissionAnswer{{QuestionID: q1, Value: "a"}},
	}
	questions := []domain.Question{descQuestion(q1, "q1", "m", 5)}
	// Both batch and retry return garbage → answer marked failed, no score set.
	llm := &stubLLM{responses: []string{"garbage", "still garbage"}}

	updated, err := gradeAnswersAI(context.Background(), llm, sub, questions, domain.AIGradingModeApply, false, uuid.New())
	if err != nil {
		t.Fatalf("gradeAnswersAI should not hard-fail on one bad answer: %v", err)
	}
	if updated.Answers[0].AIStatus != domain.AIAnswerStatusFailed {
		t.Fatalf("expected failed status, got %q", updated.Answers[0].AIStatus)
	}
	if updated.Answers[0].GradedBy == domain.GradedByAI {
		t.Fatal("failed answer must not be marked graded")
	}
}
