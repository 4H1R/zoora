package questionbanks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo      domain.QuestionBankRepository
	questions domain.QuestionRepository
	media     domain.MediaRepository
	logger    *slog.Logger
}

func NewService(
	repo domain.QuestionBankRepository,
	questions domain.QuestionRepository,
	media domain.MediaRepository,
	logger *slog.Logger,
) domain.QuestionBankService {
	return &service{repo: repo, questions: questions, media: media, logger: logger}
}

func (s *service) validateMetadataMedia(ctx context.Context, items []domain.QuestionMetadata) error {
	if err := domain.ValidateQuestionMetadata(items); err != nil {
		return err
	}
	for i, item := range items {
		m, err := s.media.FindByID(ctx, item.MediaID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.NewValidationError(map[string]string{
					fmt.Sprintf("metadata[%d].media_id", i): "media not found",
				})
			}
			return err
		}
		if m.ModelType != domain.QuestionMediaModelType {
			return domain.NewValidationError(map[string]string{
				fmt.Sprintf("metadata[%d].media_id", i): "media must belong to a question",
			})
		}
		if item.Type == domain.QuestionMetadataPhoto && !strings.HasPrefix(m.MimeType, "image/") {
			return domain.NewValidationError(map[string]string{
				fmt.Sprintf("metadata[%d].media_id", i): "media is not an image",
			})
		}
	}
	return nil
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
	options := clearOptionImagesForNonChoice(dto.Type, dto.Options)
	if err := domain.ValidateQuestionOptions(dto.Type, options); err != nil {
		return nil, err
	}
	if err := s.validateMetadataMedia(ctx, dto.Metadata); err != nil {
		return nil, err
	}
	mode, val, wpp := domain.NormalizeNegativeMark(dto.NegativeMarkMode, dto.NegativeValue, dto.WrongsPerPoint)
	if dto.Type != domain.QuestionTypeChoice {
		mode, val, wpp = domain.NegativeMarkNone, 0, 0
	}
	if err := domain.ValidateNegativeMark(mode, val, wpp); err != nil {
		return nil, err
	}
	metadata := dto.Metadata
	if metadata == nil {
		metadata = []domain.QuestionMetadata{}
	}
	question := &domain.Question{
		BankID:           bankID,
		OrganizationID:   bank.OrganizationID,
		Text:             dto.Text,
		Type:             dto.Type,
		Options:          options,
		Metadata:         metadata,
		NegativeMarkMode: mode,
		NegativeValue:    val,
		WrongsPerPoint:   wpp,
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
	question.Options = clearOptionImagesForNonChoice(question.Type, question.Options)
	if dto.Options != nil || dto.Type != nil {
		if err := domain.ValidateQuestionOptions(question.Type, question.Options); err != nil {
			return nil, err
		}
	}
	if dto.NegativeMarkMode != nil {
		question.NegativeMarkMode = *dto.NegativeMarkMode
	}
	if dto.NegativeValue != nil {
		question.NegativeValue = *dto.NegativeValue
	}
	if dto.WrongsPerPoint != nil {
		question.WrongsPerPoint = *dto.WrongsPerPoint
	}
	mode, val, wpp := domain.NormalizeNegativeMark(question.NegativeMarkMode, question.NegativeValue, question.WrongsPerPoint)
	if question.Type != domain.QuestionTypeChoice {
		mode, val, wpp = domain.NegativeMarkNone, 0, 0
	}
	if err := domain.ValidateNegativeMark(mode, val, wpp); err != nil {
		return nil, err
	}
	question.NegativeMarkMode, question.NegativeValue, question.WrongsPerPoint = mode, val, wpp
	if dto.Metadata != nil {
		if err := s.validateMetadataMedia(ctx, dto.Metadata); err != nil {
			return nil, err
		}
		question.Metadata = dto.Metadata
	}
	if err := s.questions.Update(ctx, question); err != nil {
		return nil, err
	}
	return question, nil
}

// clearOptionImagesForNonChoice strips ImageMediaID from options when the
// question type is not choice (images are choice-only).
func clearOptionImagesForNonChoice(t domain.QuestionType, in []domain.QuestionOption) []domain.QuestionOption {
	if t == domain.QuestionTypeChoice || in == nil {
		return in
	}
	out := make([]domain.QuestionOption, len(in))
	for i, o := range in {
		o.ImageMediaID = nil
		out[i] = o
	}
	return out
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
