package qa

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo   domain.QARepository
	votes  domain.QAVoteRepository
	authz  domain.ModelAuthorizer
	tx     domain.Transactor
	audit  domain.AuditRecorder
	logger *slog.Logger
	// broadcaster pushes realtime events; may be nil (tests/worker) -> no-op.
	broadcaster *broadcaster
}

func NewService(
	repo domain.QARepository,
	votes domain.QAVoteRepository,
	authz domain.ModelAuthorizer,
	tx domain.Transactor,
	audit domain.AuditRecorder,
	logger *slog.Logger,
	broadcaster *broadcaster,
) domain.QAService {
	return &service{repo: repo, votes: votes, authz: authz, tx: tx, audit: audit, logger: logger, broadcaster: broadcaster}
}

// qaLabelMaxRunes bounds the audit target label taken from free-text question
// bodies. Rune-based so multibyte (Persian) text isn't split mid-character.
const qaLabelMaxRunes = 80

func qaLabel(text string) string {
	r := []rune(text)
	if len(r) <= qaLabelMaxRunes {
		return text
	}
	return string(r[:qaLabelMaxRunes]) + "…"
}

func (s *service) Ask(ctx context.Context, dto domain.CreateQAQuestionDTO) (*domain.QAQuestion, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	allowed, err := s.authz.CanParticipate(ctx, caller, dto.ModelType, dto.ModelID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	open, err := s.repo.CountOpenByUser(ctx, dto.ModelType, dto.ModelID, caller.UserID)
	if err != nil {
		return nil, err
	}
	if open >= domain.MaxOpenQuestionsPerUser {
		return nil, domain.NewValidationError(map[string]string{
			"text": "you already have the maximum number of open questions",
		})
	}
	q := &domain.QAQuestion{
		UserID:    caller.UserID,
		ModelType: dto.ModelType,
		ModelID:   dto.ModelID,
		Text:      dto.Text,
		Status:    domain.QAStatusOpen,
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, q); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditCreated,
			TargetType:  domain.AuditTargetQA,
			TargetID:    &q.ID,
			TargetLabel: qaLabel(q.Text),
			Metadata: map[string]any{
				"model_type": q.ModelType,
				"model_id":   q.ModelID.String(),
			},
		})
	})
	if err != nil {
		return nil, err
	}
	s.broadcastCreated(ctx, q, caller)
	return q, nil
}

func (s *service) List(ctx context.Context, q domain.ListQAQuestionsQuery) ([]domain.QAQuestionView, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	// A model filter is required so we can authorize the viewer against it.
	if q.ModelType == nil || q.ModelID == nil {
		return nil, 0, domain.NewValidationError(map[string]string{
			"model_id": "model_type and model_id are required",
		})
	}
	allowed, err := s.authz.CanParticipate(ctx, caller, *q.ModelType, *q.ModelID)
	if err != nil {
		return nil, 0, err
	}
	if !allowed {
		return nil, 0, domain.ErrForbidden
	}
	scope := domain.QAListScope{
		ViewerID:  caller.UserID,
		ModelType: q.ModelType,
		ModelID:   q.ModelID,
		Status:    q.Status,
	}
	return s.repo.List(ctx, scope, q.ListParams)
}

func (s *service) UpdateText(ctx context.Context, id uuid.UUID, dto domain.UpdateQAQuestionDTO) (*domain.QAQuestion, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if q.UserID != caller.UserID {
		return nil, domain.ErrForbidden
	}
	if q.Status != domain.QAStatusOpen {
		return nil, domain.NewValidationError(map[string]string{
			"text": "cannot edit a closed question",
		})
	}
	oldText := q.Text
	q.Text = dto.Text
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, q); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetQA,
			TargetID:    &q.ID,
			TargetLabel: qaLabel(q.Text),
			Metadata: map[string]any{
				"changed": map[string]any{"text": map[string]any{"from": qaLabel(oldText), "to": qaLabel(q.Text)}},
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if q.UserID != caller.UserID {
		mod, err := s.authz.CanModerate(ctx, caller, q.ModelType, q.ModelID)
		if err != nil {
			return err
		}
		if !mod {
			return domain.ErrForbidden
		}
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Delete(ctx, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditDeleted,
			TargetType:  domain.AuditTargetQA,
			TargetID:    &id,
			TargetLabel: qaLabel(q.Text),
			Metadata: map[string]any{
				"model_type": q.ModelType,
				"model_id":   q.ModelID.String(),
			},
		})
	})
	if err != nil {
		return err
	}
	s.broadcastStatus(ctx, q, "deleted")
	return nil
}

func (s *service) ToggleVote(ctx context.Context, id uuid.UUID) (bool, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return false, 0, domain.ErrForbidden
	}
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return false, 0, err
	}
	allowed, err := s.authz.CanParticipate(ctx, caller, q.ModelType, q.ModelID)
	if err != nil {
		return false, 0, err
	}
	if !allowed {
		return false, 0, domain.ErrForbidden
	}
	if q.Status != domain.QAStatusOpen {
		return false, 0, domain.NewValidationError(map[string]string{"vote": "question is closed"})
	}
	if q.UserID == caller.UserID {
		return false, 0, domain.NewValidationError(map[string]string{"vote": "cannot vote on your own question"})
	}

	removed, err := s.votes.Delete(ctx, id, caller.UserID)
	if err != nil {
		return false, 0, err
	}
	voted := false
	if !removed {
		if err := s.votes.Create(ctx, &domain.QAVote{QuestionID: id, UserID: caller.UserID}); err != nil {
			return false, 0, err
		}
		voted = true
	}
	count, err := s.votes.CountByQuestion(ctx, id)
	if err != nil {
		return false, 0, err
	}
	s.broadcastVoted(ctx, q, count)
	return voted, count, nil
}

func (s *service) Resolve(ctx context.Context, id uuid.UUID) (*domain.QAQuestion, error) {
	return s.close(ctx, id, domain.QAStatusResolved)
}

func (s *service) Dismiss(ctx context.Context, id uuid.UUID) (*domain.QAQuestion, error) {
	return s.close(ctx, id, domain.QAStatusDismissed)
}

func (s *service) close(ctx context.Context, id uuid.UUID, status string) (*domain.QAQuestion, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	mod, err := s.authz.CanModerate(ctx, caller, q.ModelType, q.ModelID)
	if err != nil {
		return nil, err
	}
	if !mod {
		return nil, domain.ErrForbidden
	}
	oldStatus := q.Status
	now := time.Now()
	q.Status = status
	q.ClosedAt = &now
	q.ClosedBy = &caller.UserID
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, q); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetQA,
			TargetID:    &q.ID,
			TargetLabel: qaLabel(q.Text),
			Metadata: map[string]any{
				"changed": map[string]any{"status": map[string]any{"from": oldStatus, "to": status}},
			},
		})
	})
	if err != nil {
		return nil, err
	}
	s.broadcastStatus(ctx, q, status)
	return q, nil
}

func (s *service) Reopen(ctx context.Context, id uuid.UUID) (*domain.QAQuestion, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	mod, err := s.authz.CanModerate(ctx, caller, q.ModelType, q.ModelID)
	if err != nil {
		return nil, err
	}
	if !mod {
		return nil, domain.ErrForbidden
	}
	oldStatus := q.Status
	q.Status = domain.QAStatusOpen
	q.ClosedAt = nil
	q.ClosedBy = nil
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, q); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetQA,
			TargetID:    &q.ID,
			TargetLabel: qaLabel(q.Text),
			Metadata: map[string]any{
				"changed": map[string]any{"status": map[string]any{"from": oldStatus, "to": domain.QAStatusOpen}},
			},
		})
	})
	if err != nil {
		return nil, err
	}
	s.broadcastStatus(ctx, q, domain.QAStatusOpen)
	return q, nil
}
