package roles

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) domain.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) List(ctx context.Context) ([]domain.Permission, error) {
	var perms []domain.Permission
	if err := database.DB(ctx, r.db).Order("name ASC").Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("permissions.repository.List: %w", err)
	}
	return perms, nil
}

func (r *permissionRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Permission, error) {
	var perms []domain.Permission
	if err := database.DB(ctx, r.db).Where("id IN ?", ids).Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("permissions.repository.FindByIDs: %w", err)
	}
	return perms, nil
}
