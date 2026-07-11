package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// LeadStatus is the pipeline stage a sales lead sits in.
type LeadStatus string

const (
	LeadStatusNew       LeadStatus = "new"
	LeadStatusContacted LeadStatus = "contacted"
	LeadStatusConverted LeadStatus = "converted"
	LeadStatusRejected  LeadStatus = "rejected"
)

func (s LeadStatus) Valid() bool {
	switch s {
	case LeadStatusNew, LeadStatusContacted, LeadStatusConverted, LeadStatusRejected:
		return true
	}
	return false
}

// Lead is a contact captured from the public "Get started" form on the
// marketing site. A platform admin works it through the status pipeline and, on
// convert, provisions the org + owner account, linking OrganizationID back.
//
// Hard-deleted, not soft-deleted (deliberately no DeletedAt): junk leads are
// purged outright rather than archived.
type Lead struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Name    string    `gorm:"not null" json:"name"`
	Phone   string    `gorm:"not null" json:"phone"`
	OrgName string    `gorm:"not null" json:"org_name"`
	// Plan is the free-text plan key from the pricing card the visitor clicked
	// (e.g. "pro", "free"); advisory sales context, not an enforced Plan.
	Plan   string     `json:"plan"`
	Note   string     `json:"note"`
	Status LeadStatus `gorm:"type:varchar(20);not null;default:'new'" json:"status"`
	// OrganizationID links the org this lead became, set on convert.
	OrganizationID *uuid.UUID `gorm:"type:uuid" json:"organization_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateLeadDTO is the public form payload. Unauthenticated — no caller.
type CreateLeadDTO struct {
	Name    string `json:"name" binding:"required,min=2,max=255"`
	Phone   string `json:"phone" binding:"required,min=3,max=32"`
	OrgName string `json:"org_name" binding:"required,min=2,max=255"`
	Plan    string `json:"plan" binding:"omitempty,max=32"`
	Note    string `json:"note" binding:"omitempty,max=2000"`
	// Website is a honeypot: hidden from real users via CSS, only bots fill it.
	// A non-empty value means the submission is silently dropped as spam.
	Website string `json:"website"`
}

type UpdateLeadStatusDTO struct {
	Status LeadStatus `json:"status" binding:"required,oneof=new contacted converted rejected"`
}

// ConvertLeadDTO provisions an org + owner from a lead in one atomic step. The
// admin picks the real slug/plan and sets the owner's login credentials.
type ConvertLeadDTO struct {
	OrgName       string     `json:"org_name" binding:"required,min=2,max=255"`
	Slug          string     `json:"slug" binding:"required,min=2,max=63"`
	Plan          Plan       `json:"plan" binding:"required"`
	PlanExpiresAt *time.Time `json:"plan_expires_at"`
	OwnerName     string     `json:"owner_name" binding:"required,min=2,max=255"`
	OwnerUsername string     `json:"owner_username" binding:"required,min=3,max=255"`
	OwnerPassword string     `json:"owner_password" binding:"required,min=8,max=255"`
}

type AdminListLeadsQuery struct {
	Status     *LeadStatus `form:"status" binding:"omitempty,oneof=new contacted converted rejected"`
	ListParams ListParams  `form:"-"`
}

type LeadRepository interface {
	Create(ctx context.Context, lead *Lead) error
	FindByID(ctx context.Context, id uuid.UUID) (*Lead, error)
	// FindOpenByPhone returns the most recent non-terminal (new/contacted) lead
	// for a phone number, or ErrNotFound. Backs submit-time dedupe.
	FindOpenByPhone(ctx context.Context, phone string) (*Lead, error)
	Update(ctx context.Context, lead *Lead) error
	AdminList(ctx context.Context, q AdminListLeadsQuery) ([]Lead, int64, error)
	HardDelete(ctx context.Context, id uuid.UUID) error
}

type LeadService interface {
	// Submit records a public lead. Honeypot hits are dropped (nil, nil). An
	// open lead with the same phone is updated in place instead of duplicated.
	Submit(ctx context.Context, dto CreateLeadDTO) (*Lead, error)

	// Admin surface. Require caller.IsAdmin.
	AdminList(ctx context.Context, q AdminListLeadsQuery) ([]Lead, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, dto UpdateLeadStatusDTO) (*Lead, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
	// Convert atomically creates the org (+ settings) and owner user, marks the
	// lead converted, and links the new org.
	Convert(ctx context.Context, id uuid.UUID, dto ConvertLeadDTO) (*Lead, error)
}
