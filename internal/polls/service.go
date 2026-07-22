package polls

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo      domain.PollRepository
	answers   domain.PollAnswerRepository
	modelAuth domain.ModelAuthorizer
	logger    *slog.Logger
}

func NewService(
	repo domain.PollRepository,
	answers domain.PollAnswerRepository,
	modelAuth domain.ModelAuthorizer,
	logger *slog.Logger,
) domain.PollService {
	return &service{
		repo:      repo,
		answers:   answers,
		modelAuth: modelAuth,
		logger:    logger,
	}
}

func (s *service) Create(ctx context.Context, dto domain.CreatePollDTO) (*domain.Poll, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	ok, err := s.modelAuth.CanModerate(ctx, caller, dto.ModelType, dto.ModelID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}
	poll := &domain.Poll{
		UserID:              caller.UserID,
		ModelType:           dto.ModelType,
		ModelID:             dto.ModelID,
		Name:                dto.Name,
		AllowedAnswersCount: dto.AllowedAnswersCount,
		Options:             dto.Options,
	}
	if err := s.repo.Create(ctx, poll); err != nil {
		return nil, err
	}
	s.logger.Info("poll created",
		"poll_id", poll.ID.String(),
		"model_type", poll.ModelType,
		"model_id", poll.ModelID.String(),
		"created_by", caller.UserID.String(),
	)
	return poll, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Poll, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	poll, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	allowed, err := s.modelAuth.CanParticipate(ctx, caller, poll.ModelType, poll.ModelID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	return poll, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdatePollDTO) (*domain.Poll, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	poll, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ok, err = s.modelAuth.CanModerate(ctx, caller, poll.ModelType, poll.ModelID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}
	if dto.Name != nil {
		poll.Name = *dto.Name
	}
	if dto.AllowedAnswersCount != nil {
		poll.AllowedAnswersCount = *dto.AllowedAnswersCount
	}
	if dto.Options != nil {
		poll.Options = dto.Options
	}
	if err := s.repo.Update(ctx, poll); err != nil {
		return nil, err
	}
	return poll, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	poll, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	ok, err = s.modelAuth.CanModerate(ctx, caller, poll.ModelType, poll.ModelID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("poll deleted",
		"poll_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) List(ctx context.Context, q domain.ListPollsQuery) ([]domain.Poll, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	// Non-admins must scope the listing to a single model they can access. This
	// keeps the empty (all-org) scope for update_any holders unreachable without
	// an authorized per-model filter.
	if !caller.IsAdmin {
		if q.ModelType == nil || q.ModelID == nil {
			return nil, 0, domain.ErrForbidden
		}
		allowed, err := s.modelAuth.CanParticipate(ctx, caller, *q.ModelType, *q.ModelID)
		if err != nil {
			return nil, 0, err
		}
		if !allowed {
			return nil, 0, domain.ErrForbidden
		}
	}
	scope := s.resolveListScope(caller)
	scope.ModelType = q.ModelType
	scope.ModelID = q.ModelID
	if caller.HasAny(domain.PermPollsUpdateAny) {
		scope.IncludeDeleted = q.IncludeDeleted
	}
	return s.repo.List(ctx, scope, q.ListParams)
}

func (s *service) resolveListScope(caller domain.Caller) domain.PollListScope {
	if caller.IsAdmin {
		return domain.PollListScope{AllOrgs: true}
	}
	if caller.HasPermission(domain.PermPollsUpdateAny) {
		return domain.PollListScope{}
	}
	userID := caller.UserID
	return domain.PollListScope{
		OwnerID: &userID,
	}
}

func (s *service) Answer(ctx context.Context, pollID uuid.UUID, dto domain.AnswerPollDTO) ([]domain.PollAnswer, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	poll, err := s.repo.FindByID(ctx, pollID)
	if err != nil {
		return nil, err
	}
	allowed, err := s.modelAuth.CanParticipate(ctx, caller, poll.ModelType, poll.ModelID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	if poll.IsClosed() {
		return nil, domain.ErrPollClosed
	}
	if len(dto.Options) > poll.AllowedAnswersCount {
		return nil, domain.NewValidationError(map[string]string{
			"options": "exceeds allowed answers count",
		})
	}
	validOptions := make(map[string]bool, len(poll.Options))
	for _, o := range poll.Options {
		validOptions[o.Value] = true
	}
	for _, opt := range dto.Options {
		if !validOptions[opt] {
			return nil, domain.NewValidationError(map[string]string{
				"options": "invalid option: " + opt,
			})
		}
	}
	if err := s.answers.DeleteByPollAndUser(ctx, pollID, caller.UserID); err != nil {
		return nil, err
	}
	var created []domain.PollAnswer
	for _, opt := range dto.Options {
		answer := &domain.PollAnswer{
			UserID: caller.UserID,
			PollID: pollID,
			Option: opt,
		}
		if err := s.answers.Create(ctx, answer); err != nil {
			return nil, err
		}
		created = append(created, *answer)
	}
	return created, nil
}

func (s *service) CloseByModel(ctx context.Context, modelType string, modelID uuid.UUID) error {
	return s.repo.CloseByModel(ctx, modelType, modelID)
}

func (s *service) ListAnswers(ctx context.Context, pollID uuid.UUID, q domain.ListPollAnswersQuery) ([]domain.PollAnswer, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	poll, err := s.repo.FindByID(ctx, pollID)
	if err != nil {
		return nil, 0, err
	}
	// Per-voter disclosure is a privacy boundary: require moderation, not just
	// participation.
	allowed, err := s.modelAuth.CanModerate(ctx, caller, poll.ModelType, poll.ModelID)
	if err != nil {
		return nil, 0, err
	}
	if !allowed {
		return nil, 0, domain.ErrForbidden
	}
	return s.answers.ListByPoll(ctx, pollID, q)
}

func (s *service) Results(ctx context.Context, pollID uuid.UUID) (*domain.PollResults, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	poll, err := s.repo.FindByID(ctx, pollID)
	if err != nil {
		return nil, err
	}
	allowed, err := s.modelAuth.CanParticipate(ctx, caller, poll.ModelType, poll.ModelID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	counts, total, err := s.answers.CountByOption(ctx, pollID)
	if err != nil {
		return nil, err
	}
	return &domain.PollResults{
		PollID: pollID,
		Counts: counts,
		Total:  total,
	}, nil
}
