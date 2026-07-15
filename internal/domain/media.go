package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Polymorphic media identifiers. These strings are also written by the
// frontend when it uploads (e.g. live-room slides), so they must not change
// without updating the client.
const (
	MediaModelLiveRoom    = "live_room"
	MediaCollectionSlides = "slides"
	// MediaModelOrganization is the model type for standalone files uploaded
	// into an org's Shared folder on the files page; ModelID is the org ID.
	MediaModelOrganization = "organization"
	MediaCollectionShared  = "shared"
	// MediaModelConversation is the model type for chat attachments;
	// ModelID is the conversation ID. Reuses the existing media presign
	// flow — no dedicated upload endpoint (see conversations package).
	MediaModelConversation = "conversation"
	MediaCollectionAttach  = "attachments"
	// MediaModelTicket is the model type for ticket attachments; ModelID is
	// the CLASS ID (the ticket id doesn't exist yet when the first message's
	// files are presigned). Validation in the tickets service checks
	// ModelID == ticket.ClassID.
	MediaModelTicket = "ticket"
	// MediaModelOfflineRoom is the model type for offline-room attachments;
	// ModelID is the offline room ID. Uploaded via the shared media presign
	// flow (see frontend offlines/attachments.ts), collection "attachments".
	MediaModelOfflineRoom = "offline_room"
)

// Owner kinds for the files "by owner" view. Each media row (and each
// recording) resolves up its ownership chain to exactly one named owner so
// the org can see which class / question bank / conversation eats storage.
// Rows whose parent can't be resolved (deleted parent, unknown model_type)
// fall into the MediaOwnerOther bucket — visible + deletable so orphans can
// be cleaned. See MediaRepository.ListOwnerMedia.
const (
	MediaOwnerClass        = "class"
	MediaOwnerQuestionBank = "question_bank"
	MediaOwnerConversation = "conversation"
	MediaOwnerShared       = "shared"
	MediaOwnerOther        = "other"
)

// MediaOwner is one row of the files "by owner" view: a named parent entity
// with its aggregate storage across every file (and recording) it owns.
// OwnerID is nil for the Shared and Other buckets. Display names are resolved
// server-side (unlike MediaFolder, whose names are translated client-side).
type MediaOwner struct {
	OwnerKind string     `json:"owner_kind"`
	OwnerID   *uuid.UUID `json:"owner_id,omitempty"`
	Name      string     `json:"name"`
	FileCount int64      `json:"file_count"`
	TotalSize int64      `json:"total_size"`
}

// OwnerFile is one file inside an owner's drill-down. It unifies two storage
// backends: media rows (Source "media", Deletable) and recordings (Source
// "recording", read-only in v1 — no per-recording delete endpoint exists).
type OwnerFile struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`     // "media" | "recording"
	ModelType string    `json:"model_type"` // media model_type, or "recording"
	Name      string    `json:"name"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	Deletable bool      `json:"deletable"`
	CreatedAt time.Time `json:"created_at"`
}

// StorageQuota is the reconciliation header for the "by owner" view: the sum
// of every owner's bytes equals UsedBytes (media + recordings), so per-owner
// rows visibly add up to the org's plan usage. LimitBytes is 0 when Unlimited.
type StorageQuota struct {
	UsedBytes  int64 `json:"used_bytes"`
	LimitBytes int64 `json:"limit_bytes"`
	Unlimited  bool  `json:"unlimited"`
}

// MediaOwnersResponse is the "by owner" list payload: a size-sorted, paginated
// slice of owners plus the storage quota header.
type MediaOwnersResponse struct {
	Owners   []MediaOwner `json:"owners"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Quota    StorageQuota `json:"quota"`
}

type Media struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID   *uuid.UUID      `gorm:"type:uuid" json:"organization_id,omitempty"`
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

// OrgStoragePrefix is the S3 key prefix under which every object owned by an
// organization lives. Deleting an org purges all objects under this prefix in
// one sweep — the single source of truth for the per-tenant namespace shared
// with S3Key below.
func OrgStoragePrefix(orgID uuid.UUID) string {
	return "orgs/" + orgID.String() + "/"
}

// S3Key returns the object storage path for this media record. Objects are
// namespaced per tenant under orgs/{org_id}/ so a single bucket isolates each
// organization's files by key prefix. Records with no organization (e.g.
// platform-admin uploads) fall back to the un-prefixed path.
func (m Media) S3Key() string {
	base := m.ModelType + "/" + m.ModelID.String() + "/" + m.CollectionName + "/" + m.FileName
	if m.OrganizationID != nil {
		return OrgStoragePrefix(*m.OrganizationID) + base
	}
	return base
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
	// ListFolders aggregates an org's media rows by model_type.
	ListFolders(ctx context.Context, orgID uuid.UUID) ([]MediaFolder, error)
	// ListFiles pages through one org's media rows of a single model_type.
	ListFiles(ctx context.Context, orgID uuid.UUID, modelType string, p ListParams) ([]Media, int64, error)
	// ListOwnerMedia aggregates an org's media rows by resolved owner
	// (class / question_bank / conversation / shared / other). Recordings are
	// NOT included — sum them via ListOwnerRecordings and merge by (kind,id).
	ListOwnerMedia(ctx context.Context, orgID uuid.UUID) ([]MediaOwner, error)
	// ListOwnerRecordings aggregates an org's recordings by their class
	// (kind is always MediaOwnerClass on each returned row).
	ListOwnerRecordings(ctx context.Context, orgID uuid.UUID) ([]MediaOwner, error)
	// ListOwnerFiles pages the files under one resolved owner, unioning media
	// with (for class owners) read-only recordings, entirely in SQL. ownerID is
	// nil for the shared/other buckets.
	ListOwnerFiles(ctx context.Context, orgID uuid.UUID, ownerKind string, ownerID *uuid.UUID, p ListParams) ([]OwnerFile, int64, error)
}

type PresignDownloadResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

// MediaFolder is one row of the org files page's folder view: a model_type
// bucket with aggregate stats. Folder display names are translated client-side.
type MediaFolder struct {
	ModelType string `json:"model_type"`
	FileCount int64  `json:"file_count"`
	TotalSize int64  `json:"total_size"`
}

type MediaService interface {
	PresignUpload(ctx context.Context, dto PresignUploadDTO) (*PresignUploadResponse, error)
	// PresignDownload returns a presigned GET URL valid for the given expiry;
	// non-positive falls back to the service default, over-max clamps to 7d.
	PresignDownload(ctx context.Context, id uuid.UUID, expiry time.Duration) (*PresignDownloadResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]Media, error)
	// ListFolders returns the org files page's folder view. Requires
	// media:view_any (or admin) and an org-scoped caller.
	ListFolders(ctx context.Context) ([]MediaFolder, error)
	// ListFiles pages one folder (model_type) of the caller's org.
	ListFiles(ctx context.Context, modelType string, p ListParams) ([]Media, int64, error)
	// ListOwners returns the files "by owner" view: owners resolved from every
	// media row + recording, size-sorted and paginated, plus a storage quota
	// header. Requires media:view_any (or admin) and an org-scoped caller.
	ListOwners(ctx context.Context, p ListParams) (*MediaOwnersResponse, error)
	// ListOwnerFiles pages the files under a single owner, unioning media +
	// (for class owners) recordings. ownerID is nil for shared/other buckets.
	ListOwnerFiles(ctx context.Context, ownerKind string, ownerID *uuid.UUID, p ListParams) ([]OwnerFile, int64, error)
	// CleanupByModel purges a whole collection (rows + S3 objects) for a model.
	// System-level: no caller authz — invoked from background jobs only.
	CleanupByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) error
}
