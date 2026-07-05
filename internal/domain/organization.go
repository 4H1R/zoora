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
	Slug        string             `gorm:"not null;uniqueIndex" json:"slug"`
	Description string             `json:"description"`
	Status      OrganizationStatus `gorm:"not null;default:'active'" json:"status"`
	Plan          Plan       `gorm:"type:varchar(20);not null;default:'free'" json:"plan"`
	PlanExpiresAt *time.Time `json:"plan_expires_at,omitempty"`
	// TotalUsers is computed (live COUNT of non-deleted users), not a stored column.
	TotalUsers int            `gorm:"-" json:"total_users"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// EffectivePlan returns the plan actually in force now (Free if expired).
func (o *Organization) EffectivePlan(now time.Time) Plan {
	return EffectiveEntitlements(o.Plan, o.PlanExpiresAt, now).Plan
}

type CreateOrganizationDTO struct {
	Name        string             `json:"name" binding:"required,min=2"`
	Slug        string             `json:"slug" binding:"required,min=2,max=63"`
	Description string             `json:"description"`
	Status      OrganizationStatus `json:"status" binding:"omitempty,oneof=active trial suspended archived"`
}

type UpdateOrganizationDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=2"`
	Slug        *string `json:"slug" binding:"omitempty,min=2,max=63"`
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

type AdminListOrganizationsQuery struct {
	Status         *OrganizationStatus `form:"status" binding:"omitempty,oneof=active trial suspended archived"`
	IncludeDeleted bool                `form:"include_deleted"`
	ListParams     ListParams          `form:"-"`
}

type AdminCreateOrganizationDTO struct {
	Name        string             `json:"name" binding:"required,min=2"`
	Slug        string             `json:"slug" binding:"required,min=2,max=63"`
	Description string             `json:"description"`
	Status      OrganizationStatus `json:"status" binding:"omitempty,oneof=active trial suspended archived"`
}

type AdminUpdateOrganizationDTO struct {
	Name        *string             `json:"name" binding:"omitempty,min=2"`
	Slug        *string             `json:"slug" binding:"omitempty,min=2,max=63"`
	Description *string             `json:"description"`
	Status      *OrganizationStatus `json:"status" binding:"omitempty,oneof=active trial suspended archived"`
}

// SetPlanDTO is the request body for PUT /admin/organizations/:id/plan.
type SetPlanDTO struct {
	Plan      Plan       `json:"plan" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"` // nil = perpetual
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	FindByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	FindBySlug(ctx context.Context, slug string) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f OrganizationFilter) ([]Organization, int64, error)
	GetStats(ctx context.Context) (*OrganizationStats, error)

	// Admin-only operations.
	AdminList(ctx context.Context, q AdminListOrganizationsQuery) ([]Organization, int64, error)
	HardDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	UpdatePlan(ctx context.Context, id uuid.UUID, plan Plan, expiresAt *time.Time) error
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
	SetPlan(ctx context.Context, id uuid.UUID, dto SetPlanDTO) (*Organization, error)
}
