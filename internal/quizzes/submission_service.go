package quizzes

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

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
			sa.EarnedScore = gradeAnswer(q, a)
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

func gradeAnswer(q *domain.Question, a domain.SubmitAnswerDTO) float64 {
	switch q.Type {
	case domain.QuestionTypeChoice:
		return gradeChoice(q.Options, a.SelectedOptionIDs)
	case domain.QuestionTypeShortAnswer:
		return gradeShortAnswer(q.Options, a.Value)
	case domain.QuestionTypeDescriptive:
		return 0
	}
	return 0
}

func gradeChoice(options []domain.QuestionOption, selectedIDs []string) float64 {
	optMap := make(map[string]float64, len(options))
	for _, o := range options {
		optMap[o.ID] = o.Score
	}
	var total float64
	for _, id := range selectedIDs {
		if score, ok := optMap[id]; ok {
			total += score
		}
	}
	return total
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
