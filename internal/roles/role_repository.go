package roles

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

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.Role{})
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	if err := database.DB(ctx, r.db).Create(role).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("roles.repository.Create: %w", err)
	}
	return nil
}

func (r *roleRepository) FindPresetByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := database.DB(ctx, r.db).Preload("Permissions").First(&role, "name = ? AND is_preset = true", name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("roles.repository.FindPresetByName: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := database.DB(ctx, r.db).Preload("Permissions").First(&role, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("roles.repository.FindByID: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	result := database.DB(ctx, r.db).Save(role)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("roles.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Role{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("roles.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roleRepository) List(ctx context.Context, f domain.RoleFilter) ([]domain.Role, error) {
	base := r.baseQuery(ctx).Preload("Permissions")
	if f.OrganizationID != nil {
		if f.IncludePreset {
			base = base.Where("organization_id = ? OR is_preset = true", *f.OrganizationID)
		} else {
			base = base.Where("organization_id = ?", *f.OrganizationID)
		}
	}
	var roleList []domain.Role
	if err := base.Order("created_at DESC").Find(&roleList).Error; err != nil {
		return nil, fmt.Errorf("roles.repository.List: %w", err)
	}
	return roleList, nil
}

func (r *roleRepository) AdminList(ctx context.Context, f domain.AdminRoleFilter) ([]domain.Role, int64, error) {
	base := r.baseQuery(ctx).Preload("Permissions")
	if f.OrganizationID != nil {
		if f.IncludePreset {
			base = base.Where("organization_id = ? OR is_preset = true", *f.OrganizationID)
		} else {
			base = base.Where("organization_id = ?", *f.OrganizationID)
		}
	}
	var roleList []domain.Role
	total, err := listparams.Paginate(base, f.ListParams, &roleList)
	if err != nil {
		return nil, 0, fmt.Errorf("roles.repository.AdminList: %w", err)
	}
	return roleList, total, nil
}

func (r *roleRepository) Stats(ctx context.Context, orgID *uuid.UUID) (*domain.RoleStats, error) {
	db := database.DB(ctx, r.db)

	q := db.Model(&domain.Role{})
	if orgID != nil {
		q = q.Where("organization_id = ?", *orgID)
	}
	var totalRoles int64
	if err := q.Count(&totalRoles).Error; err != nil {
		return nil, fmt.Errorf("roles.repository.Stats count roles: %w", err)
	}

	var totalPerms int64
	if err := db.Model(&domain.Permission{}).Count(&totalPerms).Error; err != nil {
		return nil, fmt.Errorf("roles.repository.Stats count permissions: %w", err)
	}

	return &domain.RoleStats{TotalRoles: totalRoles, TotalPermissions: totalPerms}, nil
}

func (r *roleRepository) SetPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	if err := database.DB(ctx, r.db).Where("role_id = ?", roleID).Delete(&domain.RolePermission{}).Error; err != nil {
		return fmt.Errorf("roles.repository.SetPermissions delete: %w", err)
	}
	if len(permissionIDs) == 0 {
		return nil
	}
	rps := make([]domain.RolePermission, 0, len(permissionIDs))
	for _, permID := range permissionIDs {
		rps = append(rps, domain.RolePermission{RoleID: roleID, PermissionID: permID})
	}
	if err := database.DB(ctx, r.db).CreateInBatches(rps, 100).Error; err != nil {
		return fmt.Errorf("roles.repository.SetPermissions create: %w", err)
	}
	return nil
}

func (r *roleRepository) GetPermissionNames(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	var names []string
	err := database.DB(ctx, r.db).
		Table("permissions").
		Select("DISTINCT permissions.name").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Pluck("name", &names).Error
	if err != nil {
		return nil, fmt.Errorf("roles.repository.GetPermissionNames: %w", err)
	}
	return names, nil
}
