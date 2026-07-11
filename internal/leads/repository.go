package leads

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

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.LeadRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, lead *domain.Lead) error {
	if err := database.DB(ctx, r.db).Create(lead).Error; err != nil {
		return fmt.Errorf("leads.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	var lead domain.Lead
	if err := database.DB(ctx, r.db).First(&lead, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("leads.repository.FindByID: %w", err)
	}
	return &lead, nil
}

// FindOpenByPhone returns the newest non-terminal lead for a phone, else ErrNotFound.
func (r *repository) FindOpenByPhone(ctx context.Context, phone string) (*domain.Lead, error) {
	var lead domain.Lead
	err := database.DB(ctx, r.db).
		Where("phone = ? AND status IN ?", phone, []domain.LeadStatus{domain.LeadStatusNew, domain.LeadStatusContacted}).
		Order("created_at DESC").
		First(&lead).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("leads.repository.FindOpenByPhone: %w", err)
	}
	return &lead, nil
}

func (r *repository) Update(ctx context.Context, lead *domain.Lead) error {
	result := database.DB(ctx, r.db).Save(lead)
	if result.Error != nil {
		return fmt.Errorf("leads.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) AdminList(ctx context.Context, q domain.AdminListLeadsQuery) ([]domain.Lead, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Lead{})
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	var leads []domain.Lead
	total, err := listparams.Paginate(base, q.ListParams, &leads)
	if err != nil {
		return nil, 0, fmt.Errorf("leads.repository.AdminList: %w", err)
	}
	return leads, total, nil
}

func (r *repository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Lead{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("leads.repository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
