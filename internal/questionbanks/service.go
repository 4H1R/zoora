package questionbanks

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo      domain.QuestionBankRepository
	questions domain.QuestionRepository
	logger    *slog.Logger
}

func NewService(
	repo domain.QuestionBankRepository,
	questions domain.QuestionRepository,
	logger *slog.Logger,
) domain.QuestionBankService {
	return &service{repo: repo, questions: questions, logger: logger}
}

func canManageBank(caller domain.Caller, bank *domain.QuestionBank) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermQuestionBanksUpdateAny) {
		if caller.OrgID != nil && bank.OrganizationID != *caller.OrgID {
			return false
		}
		return true
	}
	return false
}

func canDeleteBank(caller domain.Caller, bank *domain.QuestionBank) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermQuestionBanksDeleteAny) {
		if caller.OrgID != nil && bank.OrganizationID != *caller.OrgID {
			return false
		}
		return true
	}
	return false
}

func canViewBank(caller domain.Caller, bank *domain.QuestionBank) bool {
	if canManageBank(caller, bank) {
		return true
	}
	if caller.HasPermission(domain.PermQuestionBanksViewAny) {
		if caller.OrgID != nil && bank.OrganizationID != *caller.OrgID {
			return false
		}
		return true
	}
	return caller.OrgID != nil && bank.OrganizationID == *caller.OrgID
}

func (s *service) Create(ctx context.Context, dto domain.CreateQuestionBankDTO) (*domain.QuestionBank, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin &&
		!caller.HasPermission(domain.PermQuestionBanksCreate) &&
		!caller.HasPermission(domain.PermQuestionBanksCreateAny) &&
		!caller.HasPermission(domain.PermQuestionBanksUpdateAny) {
		return nil, domain.ErrForbidden
	}
	if caller.OrgID == nil && !caller.IsAdmin {
		return nil, domain.ErrForbidden
	}
	bank := &domain.QuestionBank{
		Name:        dto.Name,
		Description: dto.Description,
	}
	if caller.OrgID != nil {
		bank.OrganizationID = *caller.OrgID
	}
	if err := s.repo.Create(ctx, bank); err != nil {
		return nil, err
	}
	s.logger.Info("question bank created",
		"bank_id", bank.ID.String(),
		"created_by", caller.UserID.String(),
	)
	return bank, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.QuestionBank, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canViewBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	return bank, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateQuestionBankDTO) (*domain.QuestionBank, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageBank(caller, bank) {
		return nil, domain.ErrForbidden
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

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteBank(caller, bank) {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("question bank deleted",
		"bank_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) List(ctx context.Context, p domain.ListParams) ([]domain.QuestionBank, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	if caller.IsAdmin {
		return s.repo.AdminList(ctx, domain.AdminListQuestionBanksQuery{ListParams: p})
	}
	if caller.OrgID == nil {
		return nil, 0, domain.ErrForbidden
	}
	if caller.HasPermission(domain.PermQuestionBanksViewAny) || caller.HasPermission(domain.PermQuestionBanksUpdateAny) {
		return s.repo.AdminList(ctx, domain.AdminListQuestionBanksQuery{OrganizationID: caller.OrgID, ListParams: p})
	}
	return s.repo.List(ctx, *caller.OrgID, p)
}

// --- Questions ---

func (s *service) CreateQuestion(ctx context.Context, bankID uuid.UUID, dto domain.CreateQuestionDTO) (*domain.Question, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, bankID)
	if err != nil {
		return nil, err
	}
	if !canManageBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	question := &domain.Question{
		BankID:         bankID,
		OrganizationID: bank.OrganizationID,
		Text:           dto.Text,
		Type:           dto.Type,
		Options:        dto.Options,
	}
	if err := s.questions.Create(ctx, question); err != nil {
		return nil, err
	}
	return question, nil
}

func (s *service) GetQuestion(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	question, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bank, err := s.repo.FindByID(ctx, question.BankID)
	if err != nil {
		return nil, err
	}
	if !canViewBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	return question, nil
}

func (s *service) UpdateQuestion(ctx context.Context, id uuid.UUID, dto domain.UpdateQuestionDTO) (*domain.Question, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	question, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bank, err := s.repo.FindByID(ctx, question.BankID)
	if err != nil {
		return nil, err
	}
	if !canManageBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	if dto.Text != nil {
		question.Text = *dto.Text
	}
	if dto.Type != nil {
		question.Type = *dto.Type
	}
	if dto.Options != nil {
		question.Options = dto.Options
	}
	if err := s.questions.Update(ctx, question); err != nil {
		return nil, err
	}
	return question, nil
}

func (s *service) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	question, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return err
	}
	bank, err := s.repo.FindByID(ctx, question.BankID)
	if err != nil {
		return err
	}
	if !canDeleteBank(caller, bank) {
		return domain.ErrForbidden
	}
	return s.questions.Delete(ctx, id)
}

func (s *service) ListQuestions(ctx context.Context, bankID uuid.UUID, q domain.ListQuestionsQuery) ([]domain.Question, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, bankID)
	if err != nil {
		return nil, 0, err
	}
	if !canViewBank(caller, bank) {
		return nil, 0, domain.ErrForbidden
	}
	if !canManageBank(caller, bank) {
		q.IncludeDeleted = false
	}
	return s.questions.ListByBank(ctx, bankID, q)
}
