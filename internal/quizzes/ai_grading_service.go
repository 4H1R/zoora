package quizzes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// StartAIGrading validates access, creates a durable job, and fans out one task
// per eligible submission on the AI queue. Returns the job for polling.
func (s *service) StartAIGrading(ctx context.Context, quizID uuid.UUID, dto domain.StartAIGradingDTO) (*domain.AIGradingJob, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.HasFeature(domain.FeatureAI) {
		return nil, domain.ErrFeatureNotInPlan
	}
	if s.llm == nil || s.aiJobs == nil {
		return nil, fmt.Errorf("ai grading: LLM provider not configured")
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}

	subs, err := s.submissions.FindByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// SubmissionQuestion carries no question Type, so descriptive answers are
	// detected by loading the referenced questions and checking their type.
	descriptive, err := s.descriptiveQuestionIDs(ctx, subs)
	if err != nil {
		return nil, err
	}

	// Only grade submitted/graded submissions that actually have descriptive answers.
	eligible := make([]domain.QuizSubmission, 0, len(subs))
	for _, sub := range subs {
		if sub.Status == domain.SubmissionStatusInProgress {
			continue
		}
		if submissionHasDescriptive(sub, descriptive) {
			eligible = append(eligible, sub)
		}
	}
	if len(eligible) == 0 {
		return nil, domain.NewValidationError(map[string]string{"quiz": "no descriptive answers to grade"})
	}

	var orgID uuid.UUID
	if caller.OrgID != nil {
		orgID = *caller.OrgID
	}
	job := &domain.AIGradingJob{
		OrganizationID: orgID,
		QuizID:         quizID,
		CreatedBy:      caller.UserID,
		Mode:           dto.Mode,
		Status:         domain.AIGradingStatusPending,
		Total:          len(eligible),
	}
	if err := s.aiJobs.Create(ctx, job); err != nil {
		return nil, err
	}

	for _, sub := range eligible {
		s.enqueueAIGrade(ctx, job, sub.ID, orgID, dto)
	}
	return job, nil
}

// GetAIGradingJob returns job progress. Manager-only via quiz ownership.
func (s *service) GetAIGradingJob(ctx context.Context, jobID uuid.UUID) (*domain.AIGradingJob, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if s.aiJobs == nil {
		return nil, fmt.Errorf("ai grading: LLM provider not configured")
	}
	job, err := s.aiJobs.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, job.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	return job, nil
}

// descriptiveQuestionIDs resolves which of the questions referenced across the
// given submissions are descriptive, by loading them once via FindByIDs.
func (s *service) descriptiveQuestionIDs(ctx context.Context, subs []domain.QuizSubmission) (map[uuid.UUID]bool, error) {
	idSet := make(map[uuid.UUID]struct{})
	for _, sub := range subs {
		for _, a := range sub.Answers {
			idSet[a.QuestionID] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return map[uuid.UUID]bool{}, nil
	}
	ids := make([]uuid.UUID, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	questions, err := s.questions.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	descriptive := make(map[uuid.UUID]bool, len(questions))
	for _, q := range questions {
		if q.Type == domain.QuestionTypeDescriptive {
			descriptive[q.ID] = true
		}
	}
	return descriptive, nil
}

func (s *service) enqueueAIGrade(ctx context.Context, job *domain.AIGradingJob, submissionID, orgID uuid.UUID, dto domain.StartAIGradingDTO) {
	payload, err := json.Marshal(domain.QuizAIGradeSubmissionPayload{
		JobID:          job.ID,
		SubmissionID:   submissionID,
		OrganizationID: orgID,
		Mode:           dto.Mode,
		Force:          dto.Force,
	})
	if err != nil {
		s.logger.Error("marshal ai-grade payload", "submission_id", submissionID.String(), "error", err)
		return
	}
	task := asynq.NewTask(domain.TypeQuizAIGradeSubmission, payload)
	_, err = s.queue.Enqueue(task,
		asynq.Queue(domain.QueueAI),
		asynq.TaskID(fmt.Sprintf("ai-grade-%s-%s", job.ID.String(), submissionID.String())),
		asynq.MaxRetry(2),
	)
	if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
		s.logger.Error("enqueue ai-grade", "submission_id", submissionID.String(), "error", err)
	}
}

// submissionHasDescriptive reports whether the submission answered any question
// that is descriptive (per the resolved descriptive-question set).
func submissionHasDescriptive(sub domain.QuizSubmission, descriptive map[uuid.UUID]bool) bool {
	for _, a := range sub.Answers {
		if descriptive[a.QuestionID] {
			return true
		}
	}
	return false
}
