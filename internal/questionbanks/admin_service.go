package questionbanks

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) AdminList(ctx context.Context, q domain.AdminListQuestionBanksQuery) ([]domain.QuestionBank, int64, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, 0, err
	}
	if q.ListParams.Page < 1 {
		q.ListParams.Page = 1
	}
	if q.ListParams.PageSize <= 0 {
		q.ListParams.PageSize = domain.DefaultPageSize
	}
	return s.repo.AdminList(ctx, q)
}

func (s *service) AdminListQuestions(ctx context.Context, q domain.AdminListQuestionsQuery) ([]domain.Question, int64, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, 0, err
	}
	if q.ListParams.Page < 1 {
		q.ListParams.Page = 1
	}
	if q.ListParams.PageSize <= 0 {
		q.ListParams.PageSize = domain.DefaultPageSize
	}
	return s.questions.AdminList(ctx, q)
}

func (s *service) AdminCreate(ctx context.Context, dto domain.AdminCreateQuestionBankDTO) (*domain.QuestionBank, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	bank := &domain.QuestionBank{
		OrganizationID: dto.OrganizationID,
		Name:           dto.Name,
		Description:    dto.Description,
	}
	if err := s.repo.Create(ctx, bank); err != nil {
		return nil, err
	}
	s.logger.Info("admin created question bank",
		"bank_id", bank.ID.String(),
		"created_by", caller.UserID.String(),
	)
	return bank, nil
}

func (s *service) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateQuestionBankDTO) (*domain.QuestionBank, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	bank, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Name != nil {
		bank.Name = *dto.Name
	}
	if dto.Description != nil {
		bank.Description = *dto.Description
	}
	if err := s.repo.Update(ctx, bank); err != nil {
		return nil, err
	}
	return bank, nil
}

func (s *service) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.repo.HardDelete(ctx, id); err != nil {
		return err
	}
	s.logger.Warn("admin hard-deleted question bank",
		"bank_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) AdminHardDeleteQuestion(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.questions.HardDelete(ctx, id); err != nil {
		return err
	}
	s.logger.Warn("admin hard-deleted question",
		"question_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}
