package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ImportType string

const (
	ImportTypeUsers        ImportType = "users"
	ImportTypeClasses      ImportType = "classes"
	ImportTypeClassMembers ImportType = "class_members"
)

type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusProcessing ImportStatus = "processing"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
)

// ImportJob tracks one bulk xlsx import run. The result file (which carries
// plaintext generated passwords) lives only in Redis with a 24h TTL — this
// row persists counters and status, nothing sensitive.
type ImportJob struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null" json:"user_id"`
	MediaID        uuid.UUID    `gorm:"type:uuid;not null" json:"media_id"`
	Type           ImportType   `gorm:"type:varchar(20);not null" json:"type"`
	Status         ImportStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	TotalRows      int          `gorm:"not null;default:0" json:"total_rows"`
	ProcessedRows  int          `gorm:"not null;default:0" json:"processed_rows"`
	CreatedCount   int          `gorm:"not null;default:0" json:"created_count"`
	SkippedCount   int          `gorm:"not null;default:0" json:"skipped_count"`
	FailedCount    int          `gorm:"not null;default:0" json:"failed_count"`
	Error          *string      `json:"error,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type CreateImportJobDTO struct {
	Type    ImportType `json:"type" binding:"required,oneof=users classes class_members"`
	MediaID uuid.UUID  `json:"media_id" binding:"required"`
}

type ImportJobRepository interface {
	Create(ctx context.Context, job *ImportJob) error
	FindByID(ctx context.Context, id uuid.UUID) (*ImportJob, error)
	Update(ctx context.Context, job *ImportJob) error
	UpdateProgress(ctx context.Context, id uuid.UUID, processed, created, skipped, failed int) error
	// Latest returns the newest job of the given type in the org, ErrNotFound when none.
	Latest(ctx context.Context, orgID uuid.UUID, t ImportType) (*ImportJob, error)
}

type ImportService interface {
	Create(ctx context.Context, dto CreateImportJobDTO) (*ImportJob, error)
	Get(ctx context.Context, id uuid.UUID) (*ImportJob, error)
	// Latest returns (nil, nil) when the org has no job of that type yet.
	Latest(ctx context.Context, t ImportType) (*ImportJob, error)
	// Result returns the stored result xlsx bytes, ErrNotFound after TTL expiry.
	Result(ctx context.Context, id uuid.UUID) ([]byte, error)
	// ProcessJob is the worker entrypoint.
	ProcessJob(ctx context.Context, payload ImportProcessPayload) error
}
