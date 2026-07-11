package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Tutorial is a platform-global "how to use Zoora" video. Like ChangelogEntry
// it has no organization_id — authored on the admin host, read by everyone.
// The video itself lives on Aparat (only the video hash is stored); the row is
// pure metadata. published_at NULL means draft. The public library is ordered
// by position ASC (curated), not by date.
type Tutorial struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	TitleEn       string     `gorm:"type:varchar(255);not null" json:"title_en"`
	TitleFa       string     `gorm:"type:varchar(255);not null;default:''" json:"title_fa"`
	DescriptionEn string     `gorm:"type:text;not null;default:''" json:"description_en"`
	DescriptionFa string     `gorm:"type:text;not null;default:''" json:"description_fa"`
	AparatHash    string     `gorm:"type:varchar(64);not null" json:"aparat_hash"`
	ThumbnailURL  string     `gorm:"type:text;not null;default:''" json:"thumbnail_url"`
	Position      int        `gorm:"not null;default:0" json:"position"`
	PublishedAt   *time.Time `gorm:"column:published_at" json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Tutorial) TableName() string { return "tutorials" }

// --- DTOs ---

type CreateTutorialDTO struct {
	TitleEn       string `json:"title_en" binding:"required,max=255"`
	TitleFa       string `json:"title_fa" binding:"omitempty,max=255"`
	DescriptionEn string `json:"description_en"`
	DescriptionFa string `json:"description_fa"`
	AparatHash    string `json:"aparat_hash" binding:"required,max=64"`
	ThumbnailURL  string `json:"thumbnail_url" binding:"omitempty,url"`
}

type UpdateTutorialDTO struct {
	TitleEn       *string `json:"title_en" binding:"omitempty,max=255"`
	TitleFa       *string `json:"title_fa" binding:"omitempty,max=255"`
	DescriptionEn *string `json:"description_en"`
	DescriptionFa *string `json:"description_fa"`
	AparatHash    *string `json:"aparat_hash" binding:"omitempty,max=64"`
	ThumbnailURL  *string `json:"thumbnail_url"`
}

// ReorderTutorialsDTO carries the full ordered list of tutorial ids. The
// backend rewrites position = index for each, so the client sends every id.
type ReorderTutorialsDTO struct {
	IDs []string `json:"ids" binding:"required,min=1,dive,uuid"`
}

// AparatOEmbedResponse is the trimmed Aparat oEmbed payload the editor needs.
// Fetched server-side because Aparat's oEmbed endpoint sends no CORS headers,
// so the admin browser cannot read it directly.
type AparatOEmbedResponse struct {
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// --- interfaces ---

type TutorialRepository interface {
	Create(ctx context.Context, tu *Tutorial) error
	Update(ctx context.Context, tu *Tutorial) error
	FindByID(ctx context.Context, id uuid.UUID) (*Tutorial, error)
	Delete(ctx context.Context, id uuid.UUID) error
	// ListPublished returns published tutorials (published_at NOT NULL) in
	// curated order (position ASC, id). No pagination — the library is small
	// and the client filters/searches it in-browser.
	ListPublished(ctx context.Context) ([]Tutorial, error)
	// AdminList returns all tutorials incl. drafts, position ASC.
	AdminList(ctx context.Context) ([]Tutorial, error)
	// MaxPosition returns the highest position in use (0 if the table is empty),
	// so a new draft can append to the end.
	MaxPosition(ctx context.Context) (int, error)
	// Reorder sets position = index for the given ordered ids in one tx.
	Reorder(ctx context.Context, ids []uuid.UUID) error
}

type TutorialService interface {
	// Public
	ListPublished(ctx context.Context) ([]Tutorial, error)
	// Admin
	AdminList(ctx context.Context) ([]Tutorial, error)
	AdminGet(ctx context.Context, id uuid.UUID) (*Tutorial, error)
	AdminCreate(ctx context.Context, dto CreateTutorialDTO) (*Tutorial, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto UpdateTutorialDTO) (*Tutorial, error)
	AdminPublish(ctx context.Context, id uuid.UUID) (*Tutorial, error)
	AdminUnpublish(ctx context.Context, id uuid.UUID) (*Tutorial, error)
	AdminDelete(ctx context.Context, id uuid.UUID) error
	AdminReorder(ctx context.Context, dto ReorderTutorialsDTO) error
	// AdminAparatOEmbed proxies Aparat's oEmbed lookup (server-side, to dodge
	// the missing CORS headers) so the editor can prefill title + thumbnail.
	AdminAparatOEmbed(ctx context.Context, hash string) (*AparatOEmbedResponse, error)
}
