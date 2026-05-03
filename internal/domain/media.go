package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ModelType        string          `gorm:"type:varchar(100);not null" json:"model_type"`
	ModelID          uuid.UUID       `gorm:"type:uuid;not null" json:"model_id"`
	CollectionName   string          `gorm:"type:varchar(100);not null;default:''" json:"collection_name"`
	Name             string          `gorm:"type:varchar(255);not null;default:''" json:"name"`
	FileName         string          `gorm:"type:varchar(255);not null" json:"file_name"`
	MimeType         string          `gorm:"type:varchar(100);not null;default:''" json:"mime_type"`
	Disk             string          `gorm:"type:varchar(50);not null;default:'s3'" json:"disk"`
	Size             int64           `gorm:"not null;default:0" json:"size"`
	CustomProperties json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"custom_properties"`
	OrderColumn      int             `gorm:"not null;default:0" json:"order_column"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// S3Key returns the object storage path for this media record.
func (m Media) S3Key() string {
	return m.ModelType + "/" + m.ModelID.String() + "/" + m.CollectionName + "/" + m.FileName
}

type CreateMediaDTO struct {
	ModelType        string          `json:"model_type" binding:"required,max=100"`
	ModelID          uuid.UUID       `json:"model_id" binding:"required"`
	CollectionName   string          `json:"collection_name" binding:"max=100"`
	Name             string          `json:"name" binding:"max=255"`
	FileName         string          `json:"file_name" binding:"required,max=255"`
	MimeType         string          `json:"mime_type" binding:"required,max=100"`
	Disk             string          `json:"disk" binding:"omitempty,max=50"`
	Size             int64           `json:"size" binding:"gte=0"`
	CustomProperties json.RawMessage `json:"custom_properties"`
	OrderColumn      int             `json:"order_column" binding:"gte=0"`
}

type PresignUploadDTO struct {
	ModelType      string `json:"model_type" binding:"required,max=100"`
	ModelID        string `json:"model_id" binding:"required,uuid"`
	CollectionName string `json:"collection_name" binding:"max=100"`
	FileName       string `json:"file_name" binding:"required,max=255"`
	MimeType       string `json:"mime_type" binding:"required,max=100"`
	Size           int64  `json:"size" binding:"required,gt=0"`
}

type PresignUploadResponse struct {
	UploadURL string `json:"upload_url"`
	Key       string `json:"key"`
	Media     *Media `json:"media"`
}

type MediaRepository interface {
	Create(ctx context.Context, m *Media) error
	FindByID(ctx context.Context, id uuid.UUID) (*Media, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]Media, error)
}

type MediaService interface {
	PresignUpload(ctx context.Context, dto PresignUploadDTO) (*PresignUploadResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]Media, error)
}
