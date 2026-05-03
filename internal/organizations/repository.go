package organizations

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

func NewRepository(db *gorm.DB) domain.OrganizationRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, org *domain.Organization) error {
	if err := database.DB(ctx, r.db).Create(org).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("organizations.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	var org domain.Organization
	if err := database.DB(ctx, r.db).First(&org, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("organizations.repository.FindByID: %w", err)
	}
	return &org, nil
}

func (r *repository) Update(ctx context.Context, org *domain.Organization) error {
	result := database.DB(ctx, r.db).Save(org)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("organizations.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Organization{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("organizations.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) GetStats(ctx context.Context) (*domain.OrganizationStats, error) {
	db := database.DB(ctx, r.db)
	var stats domain.OrganizationStats

	if err := db.Model(&domain.Organization{}).Unscoped().Count(&stats.TotalOrganizations).Error; err != nil {
		return nil, fmt.Errorf("organizations.repository.GetStats: %w", err)
	}

	type statusCount struct {
		Status string
		Count  int64
	}
	var counts []statusCount
	if err := db.Model(&domain.Organization{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("organizations.repository.GetStats: %w", err)
	}
	var nonDeleted int64
	for _, sc := range counts {
		nonDeleted += sc.Count
		switch domain.OrganizationStatus(sc.Status) {
		case domain.OrganizationStatusActive:
			stats.ActiveCount = sc.Count
		case domain.OrganizationStatusTrial:
			stats.TrialCount = sc.Count
		case domain.OrganizationStatusSuspended:
			stats.SuspendedCount = sc.Count
		case domain.OrganizationStatusArchived:
			stats.ArchivedCount = sc.Count
		}
	}
	stats.DeletedOrganizations = stats.TotalOrganizations - nonDeleted

	if err := db.Model(&domain.Organization{}).Select("COALESCE(SUM(total_users), 0)").Scan(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("organizations.repository.GetStats: %w", err)
	}
	return &stats, nil
}

// AdminList supports IncludeDeleted, status filter, and uses the standard
// ListParams pagination/search/ordering pattern.
func (r *repository) AdminList(ctx context.Context, q domain.AdminListOrganizationsQuery) ([]domain.Organization, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Organization{})
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	var orgs []domain.Organization
	total, err := listparams.Paginate(base, q.ListParams, &orgs)
	if err != nil {
		return nil, 0, fmt.Errorf("organizations.repository.AdminList: %w", err)
	}
	return orgs, total, nil
}

// HardDelete permanently removes an organization, bypassing soft-delete.
func (r *repository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.Organization{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("organizations.repository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Restore clears deleted_at on a soft-deleted org, making it visible again.
func (r *repository) Restore(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().
		Model(&domain.Organization{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("organizations.repository.Restore: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) List(ctx context.Context, f domain.OrganizationFilter) ([]domain.Organization, int64, error) {
	var orgs []domain.Organization
	var total int64
	base := database.DB(ctx, r.db).Model(&domain.Organization{})
	if f.Search != "" {
		base = base.Where("name ILIKE ?", "%"+f.Search+"%")
	}
	if f.Status != nil {
		base = base.Where("status = ?", *f.Status)
	}
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("organizations.repository.List: %w", err)
	}
	if err := base.Session(&gorm.Session{}).Offset(f.Offset).Limit(f.Limit).Order("created_at DESC").Find(&orgs).Error; err != nil {
		return nil, 0, fmt.Errorf("organizations.repository.List: %w", err)
	}
	return orgs, total, nil
}
