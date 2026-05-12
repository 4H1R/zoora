package quizzes

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo        domain.QuizRepository
	rules       domain.QuizRuleRepository
	rooms       domain.QuizRoomRepository
	submissions domain.QuizSubmissionRepository
	questions   domain.QuestionRepository
	classes     domain.ClassRepository
	members     domain.ClassMemberRepository
	logger      *slog.Logger
}

func NewService(
	repo domain.QuizRepository,
	rules domain.QuizRuleRepository,
	rooms domain.QuizRoomRepository,
	submissions domain.QuizSubmissionRepository,
	questions domain.QuestionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	logger *slog.Logger,
) domain.QuizService {
	return &service{
		repo:        repo,
		rules:       rules,
		rooms:       rooms,
		submissions: submissions,
		questions:   questions,
		classes:     classes,
		members:     members,
		logger:      logger,
	}
}

func canManageQuiz(caller domain.Caller, quiz *domain.Quiz) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermQuizzesUpdateAny) {
		return true
	}
	return caller.UserID == quiz.UserID
}

func canDeleteQuiz(caller domain.Caller, quiz *domain.Quiz) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermQuizzesDeleteAny) {
		return true
	}
	return caller.UserID == quiz.UserID
}

func (s *service) canViewQuiz(ctx context.Context, caller domain.Caller, quiz *domain.Quiz) (bool, error) {
	if canManageQuiz(caller, quiz) {
		return true, nil
	}
	if caller.HasPermission(domain.PermQuizzesViewAny) {
		return true, nil
	}
	ok, err := s.members.Exists(ctx, quiz.ClassID, caller.UserID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *service) Create(ctx context.Context, dto domain.CreateQuizDTO) (*domain.Quiz, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, dto.ClassID)
	if err != nil {
		return nil, err
	}
	if !canManageClass(caller, class) {
		return nil, domain.ErrForbidden
	}
	quiz := &domain.Quiz{
		OrganizationID:  class.OrganizationID,
		UserID:          caller.UserID,
		ClassID:         dto.ClassID,
		Title:           dto.Title,
		Description:     dto.Description,
		DurationMinutes: dto.DurationMinutes,
	}
	if err := s.repo.Create(ctx, quiz); err != nil {
		return nil, err
	}
	s.logger.Info("quiz created",
		"quiz_id", quiz.ID.String(),
		"class_id", quiz.ClassID.String(),
		"created_by", caller.UserID.String(),
	)
	return quiz, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, id)
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
	return quiz, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateQuizDTO) (*domain.Quiz, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	if dto.Title != nil {
		quiz.Title = *dto.Title
	}
	if dto.Description != nil {
		quiz.Description = *dto.Description
	}
	if dto.DurationMinutes != nil {
		quiz.DurationMinutes = *dto.DurationMinutes
	}
	if err := s.repo.Update(ctx, quiz); err != nil {
		return nil, err
	}
	return quiz, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteQuiz(caller, quiz) {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("quiz deleted",
		"quiz_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) List(ctx context.Context, q domain.ListQuizzesQuery) ([]domain.Quiz, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	scope.ClassID = q.ClassID
	if canManage(caller) {
		scope.IncludeDeleted = q.IncludeDeleted
	}
	return s.repo.List(ctx, scope, q.ListParams)
}

func (s *service) resolveListScope(caller domain.Caller) domain.QuizListScope {
	if caller.IsAdmin {
		return domain.QuizListScope{All: true}
	}
	if caller.HasPermission(domain.PermQuizzesViewAny) || caller.HasPermission(domain.PermQuizzesUpdateAny) {
		return domain.QuizListScope{All: true, OrganizationID: caller.OrgID}
	}
	userID := caller.UserID
	return domain.QuizListScope{
		OwnerID:      &userID,
		MemberUserID: &userID,
	}
}

func canManage(caller domain.Caller) bool {
	return caller.IsAdmin || caller.HasPermission(domain.PermQuizzesUpdateAny)
}

func canManageClass(caller domain.Caller, class *domain.Class) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermClassesUpdateAny) {
		return true
	}
	return caller.UserID == class.UserID
}

// --- Rules ---

func (s *service) CreateRule(ctx context.Context, quizID uuid.UUID, dto domain.CreateQuizRuleDTO) (*domain.QuizRule, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	rule := &domain.QuizRule{
		QuizID:      quizID,
		Type:        dto.Type,
		BankID:      dto.BankID,
		QuestionIDs: dto.QuestionIDs,
		Count:       dto.Count,
		IsDynamic:   dto.IsDynamic,
	}
	if err := s.rules.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *service) GetRule(ctx context.Context, id uuid.UUID) (*domain.QuizRule, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	rule, err := s.rules.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, rule.QuizID)
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
	return rule, nil
}

func (s *service) UpdateRule(ctx context.Context, id uuid.UUID, dto domain.UpdateQuizRuleDTO) (*domain.QuizRule, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	rule, err := s.rules.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, rule.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	if dto.Type != nil {
		rule.Type = *dto.Type
	}
	if dto.BankID != nil {
		rule.BankID = dto.BankID
	}
	if dto.QuestionIDs != nil {
		rule.QuestionIDs = dto.QuestionIDs
	}
	if dto.Count != nil {
		rule.Count = *dto.Count
	}
	if dto.IsDynamic != nil {
		rule.IsDynamic = *dto.IsDynamic
	}
	if err := s.rules.Update(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *service) DeleteRule(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	rule, err := s.rules.FindByID(ctx, id)
	if err != nil {
		return err
	}
	quiz, err := s.repo.FindByID(ctx, rule.QuizID)
	if err != nil {
		return err
	}
	if !canManageQuiz(caller, quiz) {
		return domain.ErrForbidden
	}
	return s.rules.Delete(ctx, id)
}

func (s *service) ListRules(ctx context.Context, quizID uuid.UUID, q domain.ListQuizRulesQuery) ([]domain.QuizRule, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, 0, err
	}
	visible, err := s.canViewQuiz(ctx, caller, quiz)
	if err != nil {
		return nil, 0, err
	}
	if !visible {
		return nil, 0, domain.ErrForbidden
	}
	return s.rules.ListByQuiz(ctx, quizID, q.ListParams)
}

// --- Rooms ---

func (s *service) CreateRoom(ctx context.Context, quizID uuid.UUID, dto domain.CreateQuizRoomDTO) (*domain.QuizRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	room := &domain.QuizRoom{
		QuizID:         quizID,
		ClassSessionID: dto.ClassSessionID,
	}
	if err := s.rooms.Create(ctx, room); err != nil {
		return nil, err
	}
	s.logger.Info("quiz room created",
		"room_id", room.ID.String(),
		"quiz_id", quizID.String(),
		"session_id", dto.ClassSessionID.String(),
		"created_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) GetRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, room.QuizID)
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
	return room, nil
}

func (s *service) StartRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	if room.StartedAt != nil {
		return nil, domain.ErrConflict
	}
	now := time.Now()
	room.StartedAt = &now
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}
	s.logger.Info("quiz room started",
		"room_id", id.String(),
		"started_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) EndRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	if room.EndedAt != nil {
		return nil, domain.ErrConflict
	}
	now := time.Now()
	room.EndedAt = &now
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}
	s.logger.Info("quiz room ended",
		"room_id", id.String(),
		"ended_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) ListRooms(ctx context.Context, quizID uuid.UUID, q domain.ListQuizRoomsQuery) ([]domain.QuizRoom, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, 0, err
	}
	visible, err := s.canViewQuiz(ctx, caller, quiz)
	if err != nil {
		return nil, 0, err
	}
	if !visible {
		return nil, 0, domain.ErrForbidden
	}
	return s.rooms.ListByQuiz(ctx, quizID, q.ListParams)
}
