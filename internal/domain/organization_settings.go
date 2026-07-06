package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const DefaultAttendancePresentThresholdPercent = 75

// OrganizationSettings holds org-wide configuration. One row per organization
// (typed columns; add new settings as new columns).
type OrganizationSettings struct {
	ID                                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID                    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"organization_id"`
	AttendancePresentThresholdPercent int       `gorm:"not null;default:75" json:"attendance_present_threshold_percent"`
	// SMSEnabled gates the SMS delivery channel per org (platform pays per
	// message). SuperAdmin-controlled — not part of the org-facing update DTO.
	SMSEnabled bool      `gorm:"not null;default:false" json:"sms_enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (OrganizationSettings) TableName() string { return "organization_settings" }

// NewDefaultOrganizationSettings builds a settings row with default values for
// the given organization.
func NewDefaultOrganizationSettings(orgID uuid.UUID) *OrganizationSettings {
	return &OrganizationSettings{
		OrganizationID:                    orgID,
		AttendancePresentThresholdPercent: DefaultAttendancePresentThresholdPercent,
	}
}

type UpdateOrganizationSettingsDTO struct {
	AttendancePresentThresholdPercent *int `json:"attendance_present_threshold_percent" binding:"omitempty,min=1,max=100"`
}

type AdminUpdateOrgSettingsDTO struct {
	SMSEnabled *bool `json:"sms_enabled" binding:"required"`
}

type OrganizationSettingsRepository interface {
	Create(ctx context.Context, s *OrganizationSettings) error
	FindByOrgID(ctx context.Context, orgID uuid.UUID) (*OrganizationSettings, error)
	Update(ctx context.Context, s *OrganizationSettings) error
}

type OrganizationSettingsService interface {
	Get(ctx context.Context, orgID uuid.UUID) (*OrganizationSettings, error)
	Update(ctx context.Context, orgID uuid.UUID, dto UpdateOrganizationSettingsDTO) (*OrganizationSettings, error)
	// AdminUpdate mutates superAdmin-only settings (SMS gate).
	AdminUpdate(ctx context.Context, orgID uuid.UUID, dto AdminUpdateOrgSettingsDTO) (*OrganizationSettings, error)
}

// OrganizationSettingsProvider is the read-only port other features (e.g.
// attendance) depend on, so they never import the orgsettings package directly.
type OrganizationSettingsProvider interface {
	GetByOrgID(ctx context.Context, orgID uuid.UUID) (*OrganizationSettings, error)
}
