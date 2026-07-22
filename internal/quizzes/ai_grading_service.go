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

// gradeAnswersAI grades a submission's eligible descriptive answers in one batch
// call, retries any missing answer individually, applies results per mode, and
// recomputes the total. Pure over its inputs (no DB) so it is unit-testable; the
// DB wrapper is gradeSubmissionAI. orgID feeds metering context.
//
// Returns the number of eligible descriptive answers that were successfully
// scored this run and the number that failed (unresolved by the model, or with a
// non-positive max score). Callers use these counts to attribute the submission
// as done vs failed on the job.
func gradeAnswersAI(
	ctx context.Context,
	llmClient domain.LLM,
	sub *domain.QuizSubmission,
	questions []domain.Question,
	mode domain.AIGradingMode,
	force bool,
	orgID uuid.UUID,
) (updated *domain.QuizSubmission, scoredCount, failedCount int, err error) {
	qByID := make(map[uuid.UUID]domain.Question, len(questions))
	for _, q := range questions {
		qByID[q.ID] = q
	}

	// Build the batch of eligible descriptive answers.
	var items []gradeItem
	answerIdx := make(map[uuid.UUID]int, len(sub.Answers))
	for i := range sub.Answers {
		a := sub.Answers[i]
		q, ok := qByID[a.QuestionID]
		if !ok || q.Type != domain.QuestionTypeDescriptive {
			continue
		}
		if !shouldGrade(a, force) {
			continue
		}
		// A descriptive question with no positive max score cannot be scored
		// meaningfully (the model has nothing to grade out of). Mark it failed so
		// it surfaces for manual grading instead of silently becoming 0/graded.
		if q.MaxScore() <= 0 {
			sub.Answers[i].AIStatus = domain.AIAnswerStatusFailed
			failedCount++
			continue
		}
		answerIdx[a.QuestionID] = i
		items = append(items, gradeItem{Question: q, Answer: a.Value})
	}

	if len(items) > 0 {
		scored := gradeBatchWithRetry(ctx, llmClient, items, orgID)

		// Apply results; mark unresolved as failed.
		for _, it := range items {
			idx := answerIdx[it.Question.ID]
			if s, ok := scored[it.Question.ID]; ok {
				applyAIScore(&sub.Answers[idx], s, mode, force)
				scoredCount++
			} else {
				sub.Answers[idx].AIStatus = domain.AIAnswerStatusFailed
				failedCount++
			}
		}
	}

	// Recompute total from earned scores.
	var total float64
	for _, a := range sub.Answers {
		total += a.EarnedScore
	}
	sub.TotalScore = total
	return sub, scoredCount, failedCount, nil
}

// gradeBatchWithRetry does the batch call, then retries any missing answer as a
// single-item call (option A fallback). Failures leave the id out of the map.
func gradeBatchWithRetry(ctx context.Context, llmClient domain.LLM, items []gradeItem, orgID uuid.UUID) map[uuid.UUID]aiScore {
	result := make(map[uuid.UUID]aiScore, len(items))

	system, user := buildGradingPrompt(items)
	if raw, err := callLLM(ctx, llmClient, system, user, orgID); err == nil {
		if scored, _, perr := parseAndValidate(raw, items); perr == nil {
			for id, s := range scored {
				result[id] = s
			}
		}
	}

	// Retry every still-missing answer individually.
	for _, it := range items {
		if _, ok := result[it.Question.ID]; ok {
			continue
		}
		single := []gradeItem{it}
		sys, usr := buildGradingPrompt(single)
		raw, err := callLLM(ctx, llmClient, sys, usr, orgID)
		if err != nil {
			continue
		}
		scored, _, perr := parseAndValidate(raw, single)
		if perr != nil {
			continue
		}
		if s, ok := scored[it.Question.ID]; ok {
			result[it.Question.ID] = s
		}
	}
	return result
}

func callLLM(ctx context.Context, llmClient domain.LLM, system, user string, orgID uuid.UUID) (string, error) {
	resp, err := llmClient.Generate(ctx, domain.LLMRequest{
		System:         system,
		Messages:       []domain.LLMMessage{{Role: domain.LLMRoleUser, Content: user}},
		JSONMode:       true,
		Temperature:    0,
		Feature:        "ai_grading",
		OrganizationID: orgID,
	})
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

// gradeSubmissionAI is the worker entry point: load, grade, persist, advance job.
func (s *service) gradeSubmissionAI(ctx context.Context, p domain.QuizAIGradeSubmissionPayload) error {
	if s.llm == nil {
		// AI disabled after this task was enqueued; count it as failed so the job can complete.
		_ = s.aiJobs.IncrementProgress(ctx, p.JobID, 0, 1)
		return nil
	}

	sub, err := s.submissions.FindByID(ctx, p.SubmissionID)
	if err != nil {
		return s.countFailureIfExhausted(ctx, p.JobID, err)
	}
	ids := make([]uuid.UUID, 0, len(sub.Answers))
	for _, a := range sub.Answers {
		ids = append(ids, a.QuestionID)
	}
	questions, err := s.questions.FindByIDs(ctx, ids)
	if err != nil {
		return s.countFailureIfExhausted(ctx, p.JobID, err)
	}

	// Descriptive-question id set: used both to gate completion and (implicitly)
	// by gradeAnswersAI, which only grades descriptive answers.
	descriptiveIDs := make(map[uuid.UUID]bool, len(questions))
	for _, q := range questions {
		if q.Type == domain.QuestionTypeDescriptive {
			descriptiveIDs[q.ID] = true
		}
	}

	updated, scoredCount, failedCount, err := gradeAnswersAI(ctx, s.llm, sub, questions, p.Mode, p.Force, p.OrganizationID)
	if err != nil {
		return s.countFailureIfExhausted(ctx, p.JobID, err)
	}

	// Apply mode: if every descriptive answer now carries a grade, advance the
	// submission to graded.
	if p.Mode == domain.AIGradingModeApply && allDescriptiveGraded(updated, descriptiveIDs) {
		updated.Status = domain.SubmissionStatusGraded
	}
	if err := s.submissions.Update(ctx, updated); err != nil {
		return s.countFailureIfExhausted(ctx, p.JobID, err)
	}

	// Attribute the submission: if it had eligible descriptive answers but none
	// could be scored this run, count it failed; otherwise count it done.
	if scoredCount == 0 && failedCount > 0 {
		return s.aiJobs.IncrementProgress(ctx, p.JobID, 0, 1)
	}
	return s.aiJobs.IncrementProgress(ctx, p.JobID, 1, 0)
}

// countFailureIfExhausted returns err to trigger an Asynq retry, unless this is
// the final attempt — then it records the submission as failed and acks (nil) so
// the job can reach completion instead of hanging forever.
func (s *service) countFailureIfExhausted(ctx context.Context, jobID uuid.UUID, err error) error {
	retried, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	if retried >= maxRetry {
		if incErr := s.aiJobs.IncrementProgress(ctx, jobID, 0, 1); incErr != nil {
			s.logger.Error("count ai grading failure", "job_id", jobID.String(), "error", incErr)
		}
		s.logger.Error("ai grading submission failed permanently", "job_id", jobID.String(), "error", err)
		return nil
	}
	return err
}

// allDescriptiveGraded reports whether every descriptive answer has a grade
// (ai or manual). An answer is considered ungraded when GradedBy is empty —
// covering both failed AI attempts and answers not yet processed. Non-descriptive
// answers are auto-graded via EarnedScore and are ignored here.
func allDescriptiveGraded(sub *domain.QuizSubmission, descriptiveIDs map[uuid.UUID]bool) bool {
	for _, a := range sub.Answers {
		if !descriptiveIDs[a.QuestionID] {
			continue
		}
		if a.GradedBy == "" {
			return false
		}
	}
	return true
}
