package orgsettings

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

func NewRepository(db *gorm.DB) domain.OrganizationSettingsRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, s *domain.OrganizationSettings) error {
	if err := database.DB(ctx, r.db).Create(s).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("orgsettings.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	var s domain.OrganizationSettings
	if err := database.DB(ctx, r.db).First(&s, "organization_id = ?", orgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("orgsettings.repository.FindByOrgID: %w", err)
	}
	return &s, nil
}

func (r *repository) Update(ctx context.Context, s *domain.OrganizationSettings) error {
	if err := database.DB(ctx, r.db).Save(s).Error; err != nil {
		return fmt.Errorf("orgsettings.repository.Update: %w", err)
	}
	return nil
}
