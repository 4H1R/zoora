package quizzes

import (
	"context"
	"io"
	"log/slog"
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

// --- lightweight stub repos for gradeSubmissionAI-level tests. Each embeds its
// domain interface so only the methods exercised here need bodies. ---

type stubSubmissionRepo struct {
	domain.QuizSubmissionRepository
	sub     *domain.QuizSubmission
	updated *domain.QuizSubmission
}

func (r *stubSubmissionRepo) FindByID(context.Context, uuid.UUID) (*domain.QuizSubmission, error) {
	return r.sub, nil
}

func (r *stubSubmissionRepo) Update(_ context.Context, sub *domain.QuizSubmission) error {
	r.updated = sub
	return nil
}

type stubQuestionRepo struct {
	domain.QuestionRepository
	questions []domain.Question
}

func (r *stubQuestionRepo) FindByIDs(context.Context, []uuid.UUID) ([]domain.Question, error) {
	return r.questions, nil
}

type stubAIJobRepo struct {
	domain.AIGradingJobRepository
	doneDelta   int
	failedDelta int
	calls       int
}

func (r *stubAIJobRepo) IncrementProgress(_ context.Context, _ uuid.UUID, done, failed int) error {
	r.doneDelta += done
	r.failedDelta += failed
	r.calls++
	return nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
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

	updated, scored, failed, err := gradeAnswersAI(context.Background(), llm, sub, questions, domain.AIGradingModeApply, false, uuid.New())
	if err != nil {
		t.Fatalf("gradeAnswersAI: %v", err)
	}
	if scored != 2 || failed != 0 {
		t.Fatalf("expected scored=2 failed=0, got scored=%d failed=%d", scored, failed)
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

	updated, scored, failed, err := gradeAnswersAI(context.Background(), llm, sub, questions, domain.AIGradingModeApply, false, uuid.New())
	if err != nil {
		t.Fatalf("gradeAnswersAI should not hard-fail on one bad answer: %v", err)
	}
	if scored != 0 || failed != 1 {
		t.Fatalf("expected scored=0 failed=1, got scored=%d failed=%d", scored, failed)
	}
	if updated.Answers[0].AIStatus != domain.AIAnswerStatusFailed {
		t.Fatalf("expected failed status, got %q", updated.Answers[0].AIStatus)
	}
	if updated.Answers[0].GradedBy == domain.GradedByAI {
		t.Fatal("failed answer must not be marked graded")
	}
}

// A descriptive question whose max score is 0 (no positive-scored option) cannot
// be graded meaningfully; it must be marked failed rather than silently set to
// 0/graded, and it must never reach the model.
func TestGradeAnswersAISkipsZeroMaxQuestion(t *testing.T) {
	q1 := uuid.New()
	sub := &domain.QuizSubmission{
		ID:      uuid.New(),
		Status:  domain.SubmissionStatusSubmitted,
		Answers: []domain.SubmissionAnswer{{QuestionID: q1, Value: "a"}},
	}
	// Max score 0 → ineligible.
	questions := []domain.Question{descQuestion(q1, "q1", "m", 0)}
	// No LLM response should ever be consumed; empty slice would panic if it were.
	llm := &stubLLM{responses: nil}

	updated, scored, failed, err := gradeAnswersAI(context.Background(), llm, sub, questions, domain.AIGradingModeApply, false, uuid.New())
	if err != nil {
		t.Fatalf("gradeAnswersAI: %v", err)
	}
	if llm.calls != 0 {
		t.Fatalf("zero-max question must not call the model, got %d calls", llm.calls)
	}
	if scored != 0 || failed != 1 {
		t.Fatalf("expected scored=0 failed=1, got scored=%d failed=%d", scored, failed)
	}
	if updated.Answers[0].AIStatus != domain.AIAnswerStatusFailed {
		t.Fatalf("expected failed status, got %q", updated.Answers[0].AIStatus)
	}
	if updated.Answers[0].GradedBy != "" || updated.Answers[0].EarnedScore != 0 {
		t.Fatalf("zero-max answer must not be graded/scored: %+v", updated.Answers[0])
	}
}

func TestAllDescriptiveGradedFalseWhenUnprocessed(t *testing.T) {
	q1, q2 := uuid.New(), uuid.New()
	descriptiveIDs := map[uuid.UUID]bool{q1: true, q2: true}
	sub := &domain.QuizSubmission{Answers: []domain.SubmissionAnswer{
		{QuestionID: q1, GradedBy: domain.GradedByAI},
		// q2 never processed: not manual, not ai, and NOT failed-status either.
		{QuestionID: q2, GradedBy: ""},
	}}
	if allDescriptiveGraded(sub, descriptiveIDs) {
		t.Fatal("an unprocessed descriptive answer (graded_by empty) must not count as graded")
	}

	// Once q2 gets a manual grade, all descriptive answers are graded.
	sub.Answers[1].GradedBy = domain.GradedByManual
	if !allDescriptiveGraded(sub, descriptiveIDs) {
		t.Fatal("all descriptive answers graded should return true")
	}
}

func TestAllDescriptiveGradedIgnoresNonDescriptive(t *testing.T) {
	q1, q2 := uuid.New(), uuid.New()
	descriptiveIDs := map[uuid.UUID]bool{q1: true}
	sub := &domain.QuizSubmission{Answers: []domain.SubmissionAnswer{
		{QuestionID: q1, GradedBy: domain.GradedByAI},
		// A non-descriptive (auto-graded) answer with empty graded_by must be ignored.
		{QuestionID: q2, GradedBy: "", EarnedScore: 2},
	}}
	if !allDescriptiveGraded(sub, descriptiveIDs) {
		t.Fatal("non-descriptive answers must not block completion")
	}
}

// When every eligible descriptive answer fails to score, gradeSubmissionAI must
// count the submission as failed (not done) and must NOT flip status to graded.
func TestGradeSubmissionAIAllFailedCountsFailedNotGraded(t *testing.T) {
	q1 := uuid.New()
	sub := &domain.QuizSubmission{
		ID:      uuid.New(),
		Status:  domain.SubmissionStatusSubmitted,
		Answers: []domain.SubmissionAnswer{{QuestionID: q1, Value: "a"}},
	}
	subRepo := &stubSubmissionRepo{sub: sub}
	qRepo := &stubQuestionRepo{questions: []domain.Question{descQuestion(q1, "q1", "m", 5)}}
	jobs := &stubAIJobRepo{}
	svc := &service{
		submissions: subRepo,
		questions:   qRepo,
		aiJobs:      jobs,
		llm:         &stubLLM{responses: []string{"garbage", "still garbage"}},
		logger:      discardLogger(),
	}

	err := svc.gradeSubmissionAI(context.Background(), domain.QuizAIGradeSubmissionPayload{
		JobID:        uuid.New(),
		SubmissionID: sub.ID,
		Mode:         domain.AIGradingModeApply,
	})
	if err != nil {
		t.Fatalf("gradeSubmissionAI: %v", err)
	}
	if jobs.doneDelta != 0 || jobs.failedDelta != 1 {
		t.Fatalf("expected failed increment (0,1), got (%d,%d)", jobs.doneDelta, jobs.failedDelta)
	}
	if subRepo.updated == nil {
		t.Fatal("submission should have been persisted")
	}
	if subRepo.updated.Status == domain.SubmissionStatusGraded {
		t.Fatal("status must not flip to graded when no descriptive answer was scored")
	}
}

// When the LLM was disabled after enqueue (s.llm == nil), the task must ack and
// count the submission failed so the job can complete instead of hanging.
func TestGradeSubmissionAINilLLMCountsFailed(t *testing.T) {
	jobs := &stubAIJobRepo{}
	svc := &service{aiJobs: jobs, logger: discardLogger()} // llm nil

	err := svc.gradeSubmissionAI(context.Background(), domain.QuizAIGradeSubmissionPayload{
		JobID:        uuid.New(),
		SubmissionID: uuid.New(),
		Mode:         domain.AIGradingModeApply,
	})
	if err != nil {
		t.Fatalf("nil-llm path must ack (nil error), got %v", err)
	}
	if jobs.doneDelta != 0 || jobs.failedDelta != 1 {
		t.Fatalf("expected failed increment (0,1), got (%d,%d)", jobs.doneDelta, jobs.failedDelta)
	}
}
