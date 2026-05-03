package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Name      PermissionName `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Role struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid;index" json:"organization_id"`
	IsPreset       bool           `gorm:"not null;default:false" json:"is_preset"`
	Name           string         `gorm:"not null" json:"name"`
	Permissions    []Permission   `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateRoleDTO struct {
	OrganizationID *uuid.UUID  `json:"organization_id"`
	Name           string      `json:"name" binding:"required,min=2"`
	PermissionIDs  []uuid.UUID `json:"permission_ids" binding:"required,min=1"`
	IsPreset       bool        `json:"is_preset"`
}

type UpdateRoleDTO struct {
	Name          string      `json:"name" binding:"omitempty,min=2"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
	IsPreset      *bool       `json:"is_preset"`
}

type PermissionRepository interface {
	List(ctx context.Context) ([]Permission, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]Permission, error)
}

type RoleFilter struct {
	OrganizationID *uuid.UUID
	IncludePreset  bool
}

type AdminRoleFilter struct {
	OrganizationID *uuid.UUID
	IncludePreset  bool
	ListParams
}

type RoleStats struct {
	TotalRoles       int64 `json:"total_roles"`
	TotalPermissions int64 `json:"total_permissions"`
}

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
	FindPresetByName(ctx context.Context, name string) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f RoleFilter) ([]Role, error)
	AdminList(ctx context.Context, f AdminRoleFilter) ([]Role, int64, error)
	SetPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	Stats(ctx context.Context, orgID *uuid.UUID) (*RoleStats, error)
	GetPermissionNames(ctx context.Context, roleID uuid.UUID) ([]string, error)
}

type RoleService interface {
	Create(ctx context.Context, dto CreateRoleDTO) (*Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateRoleDTO) (*Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f RoleFilter) ([]Role, error)
	AdminList(ctx context.Context, f AdminRoleFilter) ([]Role, int64, error)
	Stats(ctx context.Context, orgID *uuid.UUID) (*RoleStats, error)
}
