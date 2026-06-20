package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusTrial     OrganizationStatus = "trial"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusArchived  OrganizationStatus = "archived"
)

type Organization struct {
	ID          uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Name        string             `gorm:"not null" json:"name"`
	Description string             `json:"description"`
	Status      OrganizationStatus `gorm:"not null;default:'active'" json:"status"`
	// TotalUsers is computed (live COUNT of non-deleted users), not a stored column.
	TotalUsers int `gorm:"-" json:"total_users"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateOrganizationDTO struct {
	Name        string             `json:"name" binding:"required,min=2"`
	Description string             `json:"description"`
	Status      OrganizationStatus `json:"status" binding:"omitempty,oneof=active trial suspended archived"`
}

type UpdateOrganizationDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=2"`
	Description *string `json:"description"`
}

type OrganizationFilter struct {
	Search string
	Status *OrganizationStatus
	Offset int
	Limit  int
}

type OrganizationStats struct {
	TotalOrganizations   int64 `json:"total_organizations"`
	ActiveCount          int64 `json:"active_count"`
	TrialCount           int64 `json:"trial_count"`
	SuspendedCount       int64 `json:"suspended_count"`
	ArchivedCount        int64 `json:"archived_count"`
	DeletedOrganizations int64 `json:"deleted_organizations"`
	TotalUsers           int64 `json:"total_users"`
}

// AdminListOrganizationsQuery is the query for GET /admin/organizations.
type AdminListOrganizationsQuery struct {
	Status         *OrganizationStatus `form:"status" binding:"omitempty,oneof=active trial suspended archived"`
	IncludeDeleted bool                `form:"include_deleted"`
	ListParams     ListParams          `form:"-"`
}

// AdminCreateOrganizationDTO is the body for POST /admin/organizations.
type AdminCreateOrganizationDTO struct {
	Name        string             `json:"name" binding:"required,min=2"`
	Description string             `json:"description"`
	Status      OrganizationStatus `json:"status" binding:"omitempty,oneof=active trial suspended archived"`
}

// AdminUpdateOrganizationDTO is the body for PUT /admin/organizations/:id.
type AdminUpdateOrganizationDTO struct {
	Name        *string             `json:"name" binding:"omitempty,min=2"`
	Description *string             `json:"description"`
	Status      *OrganizationStatus `json:"status" binding:"omitempty,oneof=active trial suspended archived"`
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	FindByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f OrganizationFilter) ([]Organization, int64, error)
	GetStats(ctx context.Context) (*OrganizationStats, error)

	// Admin-only operations.
	AdminList(ctx context.Context, q AdminListOrganizationsQuery) ([]Organization, int64, error)
	HardDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
}

type OrganizationService interface {
	Create(ctx context.Context, dto CreateOrganizationDTO) (*Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateOrganizationDTO) (*Organization, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f OrganizationFilter) ([]Organization, int64, error)
	GetStats(ctx context.Context) (*OrganizationStats, error)

	// Admin surface. Require caller.IsAdmin.
	AdminList(ctx context.Context, q AdminListOrganizationsQuery) ([]Organization, int64, error)
	AdminCreate(ctx context.Context, dto AdminCreateOrganizationDTO) (*Organization, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateOrganizationDTO) (*Organization, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
	AdminRestore(ctx context.Context, id uuid.UUID) error
}
