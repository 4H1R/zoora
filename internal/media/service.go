package media

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
)

const (
	presignExpiry         = 15 * time.Minute
	presignDownloadExpiry = 1 * time.Hour
)

// objectStorage is the subset of the S3 storage client the media service needs.
// Depending on the interface (not the concrete *storage.Client) keeps the
// service unit-testable with a mock.
type objectStorage interface {
	GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	DeleteObject(ctx context.Context, key string) error
}

type service struct {
	repo    domain.MediaRepository
	storage objectStorage
	ent     entitlements.Service
	logger  *slog.Logger
}

func NewService(repo domain.MediaRepository, storage objectStorage, ent entitlements.Service, logger *slog.Logger) domain.MediaService {
	return &service{repo: repo, storage: storage, ent: ent, logger: logger}
}

func (s *service) PresignUpload(ctx context.Context, dto domain.PresignUploadDTO) (*domain.PresignUploadResponse, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	// Enforce the org's storage quota against declared upload size.
	if caller.OrgID != nil && s.ent != nil {
		if err := s.ent.CheckStorageLimit(ctx, *caller.OrgID, caller.Ent, dto.Size); err != nil {
			return nil, err
		}
	}

	modelID, _ := uuid.Parse(dto.ModelID)

	m := &domain.Media{
		OrganizationID:   caller.OrgID,
		ModelType:        dto.ModelType,
		ModelID:          modelID,
		CollectionName:   dto.CollectionName,
		Name:             dto.FileName,
		FileName:         dto.FileName,
		MimeType:         dto.MimeType,
		Disk:             "s3",
		Size:             dto.Size,
		CustomProperties: json.RawMessage(`{}`),
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	url, err := s.storage.GeneratePresignedUploadURL(ctx, m.S3Key(), presignExpiry)
	if err != nil {
		return nil, err
	}

	s.logger.Info("presigned upload generated",
		"media_id", m.ID.String(),
		"key", m.S3Key(),
		"user_id", caller.UserID.String(),
	)

	return &domain.PresignUploadResponse{
		UploadURL: url,
		Key:       m.S3Key(),
		Media:     m,
	}, nil
}

func (s *service) PresignDownload(ctx context.Context, id uuid.UUID) (*domain.PresignDownloadResponse, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	key := m.S3Key()
	url, err := s.storage.GeneratePresignedDownloadURL(ctx, key, presignDownloadExpiry)
	if err != nil {
		return nil, err
	}
	return &domain.PresignDownloadResponse{URL: url, Key: key}, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	_, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	return s.repo.FindByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermMediaDeleteAny) {
		return domain.ErrForbidden
	}
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// Drop the object first; if the row delete then fails the object is already
	// gone (idempotent) and a retry re-deletes harmlessly. Object removal is
	// best-effort so a storage hiccup never strands the DB row.
	if err := s.storage.DeleteObject(ctx, m.S3Key()); err != nil {
		s.logger.Error("media delete: remove object", "media_id", id.String(), "key", m.S3Key(), "error", err)
	}
	return s.repo.Delete(ctx, id)
}

// CleanupByModel deletes every media row in a collection together with its S3
// object. System-level (no caller authz) — enqueued from background jobs, e.g.
// live-room slide teardown. Object deletion is best-effort per item so one
// storage failure never blocks the rest; row deletion errors abort so the task
// can retry.
func (s *service) CleanupByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) error {
	items, err := s.repo.ListByModel(ctx, modelType, modelID, collection)
	if err != nil {
		return err
	}
	for i := range items {
		m := items[i]
		if err := s.storage.DeleteObject(ctx, m.S3Key()); err != nil {
			s.logger.Error("media cleanup: remove object", "media_id", m.ID.String(), "key", m.S3Key(), "error", err)
		}
		if err := s.repo.Delete(ctx, m.ID); err != nil {
			return err
		}
	}
	s.logger.Info("media collection cleaned up",
		"model_type", modelType, "model_id", modelID.String(), "collection", collection, "count", len(items),
	)
	return nil
}

func (s *service) ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]domain.Media, error) {
	_, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListByModel(ctx, modelType, modelID, collection)
}
