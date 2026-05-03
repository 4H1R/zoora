package users

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

func NewRepository(db *gorm.DB) domain.UserRepository {
	return &repository{db: db}
}

func (r *repository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.User{})
}

func (r *repository) findOne(ctx context.Context, conditions ...any) (*domain.User, error) {
	var user domain.User
	if err := r.baseQuery(ctx).First(&user, conditions...).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("users.repository: %w", err)
	}
	return &user, nil
}

func (r *repository) Create(ctx context.Context, user *domain.User) error {
	if err := database.DB(ctx, r.db).Create(user).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("users.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.findOne(ctx, "username = ?", username)
}

func (r *repository) Update(ctx context.Context, user *domain.User) error {
	result := database.DB(ctx, r.db).Save(user)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("users.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.User{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("users.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) List(ctx context.Context, q domain.ListUsersQuery) ([]domain.User, int64, error) {
	base := r.baseQuery(ctx)
	if q.OrganizationID != "" {
		base = base.Where("organization_id = ?", q.OrganizationID)
	}
	var users []domain.User
	total, err := listparams.Paginate(base, q.ListParams, &users)
	if err != nil {
		return nil, 0, fmt.Errorf("users.repository.List: %w", err)
	}
	return users, total, nil
}

// HardDelete removes the row permanently, bypassing soft-delete.
func (r *repository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.User{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("users.repository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// FindByIDIncludingDeleted returns the user even if soft-deleted. Used by
// admin flows that need to inspect or restore rows.
func (r *repository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.User{}).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("users.repository.FindByIDIncludingDeleted: %w", err)
	}
	return &user, nil
}

// AdminList applies typed filters from the query struct and defers
// search/order/pagination to listparams. Soft-deleted rows are excluded
// unless IncludeDeleted is true.
func (r *repository) AdminList(ctx context.Context, q domain.AdminListUsersQuery) ([]domain.User, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.User{})
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.OrganizationID != "" {
		base = base.Where("organization_id = ?", q.OrganizationID)
	}
	if q.IsAdmin != nil {
		base = base.Where("is_admin = ?", *q.IsAdmin)
	}
	var users []domain.User
	total, err := listparams.Paginate(base, q.ListParams, &users)
	if err != nil {
		return nil, 0, fmt.Errorf("users.repository.AdminList: %w", err)
	}
	return users, total, nil
}

// CountAll returns the total user count including soft-deleted rows.
func (r *repository) CountAll(ctx context.Context) (int64, error) {
	var total int64
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.User{}).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("users.repository.CountAll: %w", err)
	}
	return total, nil
}
