package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MediaModelChangelog namespaces changelog media in the public bucket. These
// strings are also written by the frontend, so do not change without updating
// the client.
const MediaModelChangelog = "changelog"

// ChangelogEntry is a platform-global "What's New" post. It has no
// organization_id — it is authored on the admin host and read by everyone.
// published_at NULL means draft; once set it is immutable (edits never
// re-notify). Ordering is always (published_at DESC, id).
type ChangelogEntry struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Version     *string    `gorm:"type:varchar(50)" json:"version,omitempty"`
	TitleEn     string     `gorm:"type:varchar(255);not null" json:"title_en"`
	TitleFa     string     `gorm:"type:varchar(255);not null;default:''" json:"title_fa"`
	BodyEn      string     `gorm:"type:text;not null" json:"body_en"`
	BodyFa      string     `gorm:"type:text;not null;default:''" json:"body_fa"`
	IsMajor     bool       `gorm:"not null;default:false" json:"is_major"`
	PublishedAt *time.Time `gorm:"column:published_at" json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ChangelogEntry) TableName() string { return "changelog_entries" }

// --- DTOs ---

type CreateChangelogDTO struct {
	Version *string `json:"version" binding:"omitempty,max=50"`
	TitleEn string  `json:"title_en" binding:"required,max=255"`
	TitleFa string  `json:"title_fa" binding:"omitempty,max=255"`
	BodyEn  string  `json:"body_en" binding:"required"`
	BodyFa  string  `json:"body_fa"`
	IsMajor bool    `json:"is_major"`
}

type UpdateChangelogDTO struct {
	Version *string `json:"version" binding:"omitempty,max=50"`
	TitleEn *string `json:"title_en" binding:"omitempty,max=255"`
	TitleFa *string `json:"title_fa" binding:"omitempty,max=255"`
	BodyEn  *string `json:"body_en"`
	BodyFa  *string `json:"body_fa"`
	IsMajor *bool   `json:"is_major"`
}

// ChangelogStatus drives the header badge + one-time major modal.
type ChangelogStatus struct {
	CurrentVersion *string         `json:"current_version"`
	UnseenCount    int64           `json:"unseen_count"`
	HasMajorUnseen bool            `json:"has_major_unseen"`
	LatestMajor    *ChangelogEntry `json:"latest_major,omitempty"`
}

// ChangelogMediaPresignDTO requests an upload slot for changelog media.
type ChangelogMediaPresignDTO struct {
	EntryID  string `json:"entry_id" binding:"required,uuid"`
	FileName string `json:"file_name" binding:"required,max=255"`
	MimeType string `json:"mime_type" binding:"required,max=100"`
	Size     int64  `json:"size" binding:"required,gt=0"`
}

type ChangelogMediaPresignResponse struct {
	UploadURL string    `json:"upload_url"` // presigned PUT (short-lived)
	PublicURL string    `json:"public_url"` // permanent, embed in markdown
	MediaID   uuid.UUID `json:"media_id"`
}

// --- interfaces ---

type ChangelogRepository interface {
	Create(ctx context.Context, e *ChangelogEntry) error
	Update(ctx context.Context, e *ChangelogEntry) error
	FindByID(ctx context.Context, id uuid.UUID) (*ChangelogEntry, error)
	Delete(ctx context.Context, id uuid.UUID) error
	// ListPublished returns published entries (published_at NOT NULL) newest
	// first, paginated by limit/offset, plus the total count.
	ListPublished(ctx context.Context, limit, offset int) ([]ChangelogEntry, int64, error)
	// AdminList returns all entries incl. drafts, newest by created_at first.
	AdminList(ctx context.Context, limit, offset int) ([]ChangelogEntry, int64, error)
	// LatestPublished returns the newest published entry (for current version).
	LatestPublished(ctx context.Context) (*ChangelogEntry, error)
	// CountUnseen counts published entries newer than the given marker.
	CountUnseen(ctx context.Context, since *time.Time) (int64, error)
	// LatestMajorUnseen returns the newest published major entry newer than
	// the marker, or nil.
	LatestMajorUnseen(ctx context.Context, since *time.Time) (*ChangelogEntry, error)
	// GetLastSeen / UpdateLastSeen read+write users.changelog_last_seen_at.
	// Living on this repo keeps the changelog package from importing users.
	GetLastSeen(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	UpdateLastSeen(ctx context.Context, userID uuid.UUID, t time.Time) error
}

type ChangelogService interface {
	// Public
	ListPublished(ctx context.Context, p ListParams) ([]ChangelogEntry, int64, error)
	Status(ctx context.Context) (*ChangelogStatus, error)
	MarkSeen(ctx context.Context) error
	// Admin
	AdminList(ctx context.Context, p ListParams) ([]ChangelogEntry, int64, error)
	AdminGet(ctx context.Context, id uuid.UUID) (*ChangelogEntry, error)
	AdminCreate(ctx context.Context, dto CreateChangelogDTO) (*ChangelogEntry, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto UpdateChangelogDTO) (*ChangelogEntry, error)
	AdminPublish(ctx context.Context, id uuid.UUID) (*ChangelogEntry, error)
	AdminUnpublish(ctx context.Context, id uuid.UUID) (*ChangelogEntry, error)
	AdminDelete(ctx context.Context, id uuid.UUID) error
	AdminPresignMedia(ctx context.Context, dto ChangelogMediaPresignDTO) (*ChangelogMediaPresignResponse, error)
}
