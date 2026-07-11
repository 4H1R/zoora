package imports

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.ImportJobRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, job *domain.ImportJob) error {
	if err := database.DB(ctx, r.db).Create(job).Error; err != nil {
		return fmt.Errorf("imports.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ImportJob, error) {
	var job domain.ImportJob
	if err := database.DB(ctx, r.db).First(&job, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("imports.repository.FindByID: %w", err)
	}
	return &job, nil
}

func (r *repository) Update(ctx context.Context, job *domain.ImportJob) error {
	if err := database.DB(ctx, r.db).Save(job).Error; err != nil {
		return fmt.Errorf("imports.repository.Update: %w", err)
	}
	return nil
}

func (r *repository) UpdateProgress(ctx context.Context, id uuid.UUID, processed, created, skipped, failed int) error {
	err := database.DB(ctx, r.db).Model(&domain.ImportJob{}).Where("id = ?", id).
		Updates(map[string]any{
			"processed_rows": processed,
			"created_count":  created,
			"skipped_count":  skipped,
			"failed_count":   failed,
		}).Error
	if err != nil {
		return fmt.Errorf("imports.repository.UpdateProgress: %w", err)
	}
	return nil
}

func (r *repository) Latest(ctx context.Context, orgID uuid.UUID, t domain.ImportType) (*domain.ImportJob, error) {
	var job domain.ImportJob
	err := database.DB(ctx, r.db).
		Where("organization_id = ? AND type = ?", orgID, t).
		Order("created_at DESC").
		First(&job).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("imports.repository.Latest: %w", err)
	}
	return &job, nil
}
