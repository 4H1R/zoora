package questionbanks

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type bankRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.QuestionBankRepository {
	return &bankRepository{db: db}
}

func (r *bankRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.QuestionBank{})
}

func (r *bankRepository) Create(ctx context.Context, bank *domain.QuestionBank) error {
	if err := database.DB(ctx, r.db).Create(bank).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("questionbanks.repository.Create: %w", err)
	}
	return nil
}

func (r *bankRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuestionBank, error) {
	var bank domain.QuestionBank
	if err := r.baseQuery(ctx).First(&bank, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("questionbanks.repository.FindByID: %w", err)
	}
	return &bank, nil
}

func (r *bankRepository) Update(ctx context.Context, bank *domain.QuestionBank) error {
	result := database.DB(ctx, r.db).Save(bank)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("questionbanks.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bankRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.QuestionBank{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("questionbanks.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bankRepository) List(ctx context.Context, orgID uuid.UUID, p domain.ListParams) ([]domain.QuestionBank, int64, error) {
	base := r.baseQuery(ctx).Where("organization_id = ?", orgID)
	var banks []domain.QuestionBank
	total, err := listparams.Paginate(base, p, &banks)
	if err != nil {
		return nil, 0, fmt.Errorf("questionbanks.repository.List: %w", err)
	}
	return banks, total, nil
}

func (r *bankRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.QuestionBank{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("questionbanks.repository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bankRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.QuestionBank, error) {
	var bank domain.QuestionBank
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.QuestionBank{}).First(&bank, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("questionbanks.repository.FindByIDIncludingDeleted: %w", err)
	}
	return &bank, nil
}

func (r *bankRepository) AdminList(ctx context.Context, q domain.AdminListQuestionBanksQuery) ([]domain.QuestionBank, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.QuestionBank{})
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.OrganizationID != nil {
		base = base.Where("organization_id = ?", *q.OrganizationID)
	}
	var banks []domain.QuestionBank
	total, err := listparams.Paginate(base, q.ListParams, &banks)
	if err != nil {
		return nil, 0, fmt.Errorf("questionbanks.repository.AdminList: %w", err)
	}
	return banks, total, nil
}

// questionRepository handles persistence of individual questions within banks.
type questionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) domain.QuestionRepository {
	return &questionRepository{db: db}
}

func (r *questionRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.Question{})
}

func (r *questionRepository) Create(ctx context.Context, question *domain.Question) error {
	if err := database.DB(ctx, r.db).Create(question).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("questionbanks.questionRepository.Create: %w", err)
	}
	return nil
}

func (r *questionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	var q domain.Question
	if err := r.baseQuery(ctx).First(&q, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("questionbanks.questionRepository.FindByID: %w", err)
	}
	return &q, nil
}

func (r *questionRepository) Update(ctx context.Context, question *domain.Question) error {
	result := database.DB(ctx, r.db).Save(question)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("questionbanks.questionRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *questionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Question{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("questionbanks.questionRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *questionRepository) ListByBank(ctx context.Context, bankID uuid.UUID, q domain.ListQuestionsQuery) ([]domain.Question, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Question{})
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	base = base.Where("bank_id = ?", bankID)
	if q.Type != nil {
		base = base.Where("type = ?", *q.Type)
	}
	var questions []domain.Question
	total, err := listparams.Paginate(base, q.ListParams, &questions)
	if err != nil {
		return nil, 0, fmt.Errorf("questionbanks.questionRepository.ListByBank: %w", err)
	}
	return questions, total, nil
}

func (r *questionRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Question, error) {
	var questions []domain.Question
	if err := r.baseQuery(ctx).Where("id IN ?", ids).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("questionbanks.questionRepository.FindByIDs: %w", err)
	}
	return questions, nil
}

func (r *questionRepository) CountByBank(ctx context.Context, bankID uuid.UUID) (int64, error) {
	var count int64
	if err := r.baseQuery(ctx).Where("bank_id = ?", bankID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("questionbanks.questionRepository.CountByBank: %w", err)
	}
	return count, nil
}

func (r *questionRepository) RandomByBank(ctx context.Context, bankID uuid.UUID, count int) ([]domain.Question, error) {
	var questions []domain.Question
	if err := r.baseQuery(ctx).
		Where("bank_id = ?", bankID).
		Order("RANDOM()").
		Limit(count).
		Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("questionbanks.questionRepository.RandomByBank: %w", err)
	}
	return questions, nil
}

func (r *questionRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.Question{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("questionbanks.questionRepository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
