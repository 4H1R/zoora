package quizzes

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// ListQuestionsForTaking returns the ordered question list that a student
// sees while taking the quiz. Composes rules + bank questions server-side and
// strips answer-key data (option scores, short_answer correct values) so the
// caller can never derive the grading rubric.
func (s *service) ListQuestionsForTaking(ctx context.Context, quizID uuid.UUID) ([]domain.Question, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	visible, err := s.canViewQuiz(ctx, caller, quiz)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, domain.ErrForbidden
	}

	rules, _, err := s.rules.ListByQuiz(ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}

	quizWide := domain.NegativeMarkConfig{Mode: quiz.NegativeMarkMode, NegativeValue: quiz.NegativeValue, WrongsPerPoint: quiz.WrongsPerPoint}

	seen := make(map[uuid.UUID]struct{})
	out := make([]domain.Question, 0)
	for _, r := range rules {
		ovByQ := make(map[uuid.UUID]domain.QuizQuestionNegativeOverride, len(r.NegativeOverrides))
		for _, ov := range r.NegativeOverrides {
			ovByQ[ov.QuestionID] = ov
		}
		picked, err := s.pickQuestionsForRule(ctx, r)
		if err != nil {
			return nil, err
		}
		for i := range picked {
			if _, dup := seen[picked[i].ID]; dup {
				continue
			}
			seen[picked[i].ID] = struct{}{}
			sanitized := sanitizeQuestionForTaking(picked[i])
			if sanitized.Type == domain.QuestionTypeChoice {
				var ovPtr *domain.QuizQuestionNegativeOverride
				if ov, ok := ovByQ[picked[i].ID]; ok {
					ovPtr = &ov
				}
				cfg := domain.ResolveNegativeMark(picked[i], ovPtr, r.NegativeDefaultConfig(), quizWide)
				if cfg.Mode != domain.NegativeMarkNone {
					sanitized.NegativeConfig = &cfg
				}
			}
			out = append(out, sanitized)
		}
	}
	return out, nil
}

func (s *service) pickQuestionsForRule(ctx context.Context, r domain.QuizRule) ([]domain.Question, error) {
	switch r.Type {
	case domain.QuizRuleTypeManual:
		if len(r.QuestionIDs) == 0 {
			return nil, nil
		}
		qs, err := s.questions.FindByIDs(ctx, r.QuestionIDs)
		if err != nil {
			return nil, err
		}
		byID := make(map[uuid.UUID]domain.Question, len(qs))
		for i := range qs {
			byID[qs[i].ID] = qs[i]
		}
		ordered := make([]domain.Question, 0, len(r.QuestionIDs))
		for _, id := range r.QuestionIDs {
			if q, ok := byID[id]; ok {
				ordered = append(ordered, q)
			}
		}
		return ordered, nil
	case domain.QuizRuleTypeRandom:
		if r.BankID == nil || r.Count == 0 {
			return nil, nil
		}
		return s.questions.RandomByBank(ctx, *r.BankID, r.Count)
	}
	return nil, nil
}

// sanitizeQuestionForTaking removes answer-key data from a question. Choice
// options keep id+value but lose score; short_answer/descriptive lose options
// entirely because their options hold correct answers.
func sanitizeQuestionForTaking(q domain.Question) domain.Question {
	switch q.Type {
	case domain.QuestionTypeChoice:
		multi := q.IsMultiSelect()
		opts := make([]domain.QuestionOption, len(q.Options))
		for i, o := range q.Options {
			opts[i] = domain.QuestionOption{ID: o.ID, Value: o.Value, ImageMediaID: o.ImageMediaID}
		}
		q.Options = opts
		q.IsMultiSelectFlag = &multi
	case domain.QuestionTypeShortAnswer, domain.QuestionTypeDescriptive:
		q.Options = []domain.QuestionOption{}
	}
	return q
}

func (s *service) StartSubmission(ctx context.Context, quizID uuid.UUID, dto domain.StartQuizSubmissionDTO) (*domain.QuizSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	enrolled, err := s.members.Exists(ctx, quiz.ClassID, caller.UserID)
	if err != nil {
		return nil, err
	}
	if !enrolled {
		return nil, domain.ErrForbidden
	}

	room, err := s.rooms.FindByID(ctx, dto.QuizRoomID)
	if err != nil {
		return nil, err
	}
	if room.QuizID != quizID {
		return nil, domain.ErrNotFound
	}
	if !room.IsRoomOpen() {
		return nil, domain.NewValidationError(map[string]string{"quiz_room": "room is not open"})
	}

	existing, err := s.submissions.FindByQuizAndUser(ctx, quizID, caller.UserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrConflict
	}

	sub := &domain.QuizSubmission{
		QuizID:    quizID,
		UserID:    caller.UserID,
		Status:    domain.SubmissionStatusInProgress,
		Answers:   []domain.SubmissionAnswer{},
		StartedAt: time.Now(),
	}
	if err := s.submissions.Create(ctx, sub); err != nil {
		return nil, err
	}

	s.logger.Info("quiz submission started",
		"submission_id", sub.ID.String(),
		"quiz_id", quizID.String(),
		"user_id", caller.UserID.String(),
	)
	return sub, nil
}

func (s *service) SubmitQuiz(ctx context.Context, submissionID uuid.UUID, dto domain.SubmitQuizDTO) (*domain.QuizSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	sub, err := s.submissions.FindByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	if sub.UserID != caller.UserID {
		return nil, domain.ErrForbidden
	}
	if sub.Status != domain.SubmissionStatusInProgress {
		return nil, domain.ErrConflict
	}

	quiz, err := s.repo.FindByID(ctx, sub.QuizID)
	if err != nil {
		return nil, err
	}

	room, err := s.rooms.FindOpenByQuizID(ctx, sub.QuizID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewValidationError(map[string]string{"quiz_room": "room is not open"})
		}
		return nil, err
	}

	now := time.Now()
	deadline := sub.StartedAt.Add(time.Duration(quiz.DurationMinutes) * time.Minute)
	if room.EndedAt != nil && room.EndedAt.Before(deadline) {
		deadline = *room.EndedAt
	}
	grace := time.Duration(domain.SubmissionGracePeriod) * time.Second
	late := now.After(deadline.Add(grace))

	questionIDs := make([]uuid.UUID, 0, len(dto.Answers))
	for _, a := range dto.Answers {
		questionIDs = append(questionIDs, a.QuestionID)
	}

	questions, err := s.questions.FindByIDs(ctx, questionIDs)
	if err != nil {
		return nil, err
	}
	qMap := make(map[uuid.UUID]*domain.Question, len(questions))
	for i := range questions {
		qMap[questions[i].ID] = &questions[i]
	}

	// Build per-question effective negative-marking config from the quiz rules
	// (per-question override) + quiz-wide default.
	rules, _, err := s.rules.ListByQuiz(ctx, sub.QuizID, domain.ListParams{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}
	quizWide := domain.NegativeMarkConfig{Mode: quiz.NegativeMarkMode, NegativeValue: quiz.NegativeValue, WrongsPerPoint: quiz.WrongsPerPoint}
	overrideByQ := make(map[uuid.UUID]domain.QuizQuestionNegativeOverride)
	// Rule-wide defaults resolved by question: manual rules attach their default
	// to explicit QuestionIDs; random rules attach theirs to every question in
	// the referenced bank (question IDs are unknown until grade time).
	defaultByQ := make(map[uuid.UUID]*domain.NegativeMarkConfig)
	defaultByBank := make(map[uuid.UUID]*domain.NegativeMarkConfig)
	for _, r := range rules {
		for _, ov := range r.NegativeOverrides {
			overrideByQ[ov.QuestionID] = ov
		}
		rd := r.NegativeDefaultConfig()
		if rd == nil {
			continue
		}
		switch r.Type {
		case domain.QuizRuleTypeManual:
			for _, qid := range r.QuestionIDs {
				defaultByQ[qid] = rd
			}
		case domain.QuizRuleTypeRandom:
			if r.BankID != nil {
				defaultByBank[*r.BankID] = rd
			}
		}
	}
	cfgFor := func(q *domain.Question) domain.NegativeMarkConfig {
		var ovPtr *domain.QuizQuestionNegativeOverride
		if ov, ok := overrideByQ[q.ID]; ok {
			ovPtr = &ov
		}
		ruleDefault := defaultByQ[q.ID]
		if ruleDefault == nil {
			ruleDefault = defaultByBank[q.BankID]
		}
		return domain.ResolveNegativeMark(*q, ovPtr, ruleDefault, quizWide)
	}

	answers := make([]domain.SubmissionAnswer, 0, len(dto.Answers))
	var totalScore float64

	for _, a := range dto.Answers {
		sa := domain.SubmissionAnswer{
			QuestionID:        a.QuestionID,
			SelectedOptionIDs: a.SelectedOptionIDs,
			Value:             a.Value,
			SpentSeconds:      a.SpentSeconds,
		}

		q, exists := qMap[a.QuestionID]
		if exists {
			sa.EarnedScore = gradeAnswer(q, a, cfgFor(q))
		}
		totalScore += sa.EarnedScore
		answers = append(answers, sa)
	}

	sub.Answers = answers
	sub.TotalScore = totalScore
	sub.SubmittedAt = &now
	sub.Status = domain.SubmissionStatusSubmitted

	if err := s.submissions.Update(ctx, sub); err != nil {
		return nil, err
	}

	s.logger.Info("quiz submitted",
		"submission_id", sub.ID.String(),
		"quiz_id", sub.QuizID.String(),
		"user_id", caller.UserID.String(),
		"total_score", sub.TotalScore,
		"late", late,
	)
	return sub, nil
}

func (s *service) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.QuizSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	sub, err := s.submissions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if sub.UserID == caller.UserID {
		return sub, nil
	}

	quiz, err := s.repo.FindByID(ctx, sub.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	return sub, nil
}

func (s *service) ListSubmissions(ctx context.Context, quizID uuid.UUID, q domain.ListSubmissionsQuery) ([]domain.QuizSubmission, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}

	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, 0, err
	}

	if !canManageQuiz(caller, quiz) {
		userID := caller.UserID
		q.UserID = &userID
	}

	return s.submissions.ListByQuiz(ctx, quizID, q)
}

func (s *service) GradeSubmission(ctx context.Context, id uuid.UUID, dto domain.GradeSubmissionDTO) (*domain.QuizSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	sub, err := s.submissions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub.Status == domain.SubmissionStatusInProgress {
		return nil, domain.NewValidationError(map[string]string{"status": "submission not yet submitted"})
	}

	quiz, err := s.repo.FindByID(ctx, sub.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}

	gradeMap := make(map[uuid.UUID]float64, len(dto.Grades))
	for _, g := range dto.Grades {
		gradeMap[g.QuestionID] = g.EarnedScore
	}

	var totalScore float64
	for i, a := range sub.Answers {
		if score, ok := gradeMap[a.QuestionID]; ok {
			sub.Answers[i].EarnedScore = score
		}
		totalScore += sub.Answers[i].EarnedScore
	}

	sub.TotalScore = totalScore
	sub.Status = domain.SubmissionStatusGraded

	if err := s.submissions.Update(ctx, sub); err != nil {
		return nil, err
	}

	s.logger.Info("quiz submission graded",
		"submission_id", sub.ID.String(),
		"graded_by", caller.UserID.String(),
		"total_score", sub.TotalScore,
	)
	return sub, nil
}

func gradeAnswer(q *domain.Question, a domain.SubmitAnswerDTO, cfg domain.NegativeMarkConfig) float64 {
	switch q.Type {
	case domain.QuestionTypeChoice:
		return gradeChoice(q.Options, a.SelectedOptionIDs, cfg)
	case domain.QuestionTypeShortAnswer:
		return gradeShortAnswer(q.Options, a.Value)
	case domain.QuestionTypeDescriptive:
		return 0
	}
	return 0
}

// gradeChoice computes earned = positive - penalty. Only the SIGN of an
// option's score matters: Score>0 is correct, Score<=0 is a distractor.
// Penalty is driven by the resolved negative-mark config. Result may be negative.
func gradeChoice(options []domain.QuestionOption, selectedIDs []string, cfg domain.NegativeMarkConfig) float64 {
	optMap := make(map[string]float64, len(options))
	for _, o := range options {
		optMap[o.ID] = o.Score
	}
	var positive float64
	wrongCount := 0
	for _, id := range selectedIDs {
		score, ok := optMap[id]
		if !ok {
			continue
		}
		if score > 0 {
			positive += score
		} else {
			wrongCount++
		}
	}
	return positive - cfg.Penalty(wrongCount)
}

func gradeShortAnswer(options []domain.QuestionOption, value string) float64 {
	normalized := normalizeString(value)
	for _, o := range options {
		if o.Score > 0 && normalizeString(o.Value) == normalized {
			return o.Score
		}
	}
	return 0
}

func normalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(strings.Join(strings.Fields(s), " ")))
}
