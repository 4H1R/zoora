package media

import (
	"context"
	"encoding/json"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
)

const (
	presignExpiry            = 15 * time.Minute
	presignDownloadExpiry    = 1 * time.Hour
	presignDownloadMaxExpiry = 7 * 24 * time.Hour // SigV4 hard limit
	// maxSharedUploadSize caps the declared size of Shared-folder uploads.
	// Advisory: a presigned PUT cannot enforce the actual byte count.
	maxSharedUploadSize = 200 << 20
)

// objectStorage is the subset of the S3 storage client the media service needs.
// Depending on the interface (not the concrete *storage.Client) keeps the
// service unit-testable with a mock.
type objectStorage interface {
	GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	DeleteObject(ctx context.Context, key string) error
}

// storageUsageReader reports an org's total stored bytes (media + recordings)
// for the files page quota header. Satisfied by entitlements.Repository; kept
// narrow so the media service doesn't depend on the whole entitlements repo.
type storageUsageReader interface {
	SumStorageBytes(ctx context.Context, orgID uuid.UUID) (int64, error)
}

type service struct {
	repo    domain.MediaRepository
	storage objectStorage
	ent     entitlements.Service
	usage   storageUsageReader
	logger  *slog.Logger
}

func NewService(repo domain.MediaRepository, storage objectStorage, ent entitlements.Service, usage storageUsageReader, logger *slog.Logger) domain.MediaService {
	return &service{repo: repo, storage: storage, ent: ent, usage: usage, logger: logger}
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

	sharedUpload := dto.ModelType == domain.MediaModelOrganization
	if sharedUpload {
		if caller.OrgID == nil || dto.ModelID != caller.OrgID.String() {
			return nil, domain.NewValidationError(map[string]string{"model_id": "must be your organization id"})
		}
		if dto.Size > maxSharedUploadSize {
			return nil, domain.NewValidationError(map[string]string{"size": "must be at most 200 MB"})
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

	if sharedUpload {
		// All Shared uploads live under one model_id + collection, so the raw
		// file name IS the S3 key tail — prefix it to prevent overwrites.
		m.CollectionName = domain.MediaCollectionShared
		m.FileName = uuid.NewString()[:8] + "-" + dto.FileName
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

// authorizeOrgAccess hides media belonging to another tenant. Platform-global
// rows (nil OrganizationID, e.g. changelog assets) stay reachable by any
// authenticated caller; admins bypass. Returns ErrNotFound rather than
// ErrForbidden so cross-org probing can't confirm an ID exists.
func authorizeOrgAccess(caller domain.Caller, m *domain.Media) error {
	if m.OrganizationID == nil || caller.IsAdmin {
		return nil
	}
	if caller.OrgID == nil || *caller.OrgID != *m.OrganizationID {
		return domain.ErrNotFound
	}
	return nil
}

func (s *service) PresignDownload(ctx context.Context, id uuid.UUID, expiry time.Duration) (*domain.PresignDownloadResponse, error) {
	if expiry <= 0 {
		expiry = presignDownloadExpiry
	}
	expiry = min(expiry, presignDownloadMaxExpiry)
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := authorizeOrgAccess(caller, m); err != nil {
		return nil, err
	}
	key := m.S3Key()
	url, err := s.storage.GeneratePresignedDownloadURL(ctx, key, expiry)
	if err != nil {
		return nil, err
	}
	return &domain.PresignDownloadResponse{URL: url, Key: key}, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := authorizeOrgAccess(caller, m); err != nil {
		return nil, err
	}
	return m, nil
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
	if err := authorizeOrgAccess(caller, m); err != nil {
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

// requireOrgViewAny is the shared gate for the org files page endpoints.
func requireOrgViewAny(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || caller.OrgID == nil {
		return domain.Caller{}, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermMediaViewAny) {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) ListFolders(ctx context.Context) ([]domain.MediaFolder, error) {
	caller, err := requireOrgViewAny(ctx)
	if err != nil {
		return nil, err
	}
	folders, err := s.repo.ListFolders(ctx, *caller.OrgID)
	if err != nil {
		return nil, err
	}
	// The Shared folder is always offered so the page can accept uploads
	// before the first file exists.
	for _, f := range folders {
		if f.ModelType == domain.MediaModelOrganization {
			return folders, nil
		}
	}
	return append(folders, domain.MediaFolder{ModelType: domain.MediaModelOrganization}), nil
}

func (s *service) ListFiles(ctx context.Context, modelType string, p domain.ListParams) ([]domain.Media, int64, error) {
	caller, err := requireOrgViewAny(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.ListFiles(ctx, *caller.OrgID, modelType, p)
}

func (s *service) ListOwners(ctx context.Context, p domain.ListParams) (*domain.MediaOwnersResponse, error) {
	caller, err := requireOrgViewAny(ctx)
	if err != nil {
		return nil, err
	}
	orgID := *caller.OrgID

	mediaOwners, err := s.repo.ListOwnerMedia(ctx, orgID)
	if err != nil {
		return nil, err
	}
	recOwners, err := s.repo.ListOwnerRecordings(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Merge media + recordings that resolve to the same (kind, id) — a class
	// row sums its slides, attachments, and recordings into one bucket.
	type key struct {
		kind string
		id   uuid.UUID // uuid.Nil for shared/other
	}
	merged := make(map[key]*domain.MediaOwner)
	order := make([]key, 0, len(mediaOwners)+len(recOwners))
	add := func(o domain.MediaOwner) {
		var id uuid.UUID
		if o.OwnerID != nil {
			id = *o.OwnerID
		}
		k := key{kind: o.OwnerKind, id: id}
		cur, ok := merged[k]
		if !ok {
			cp := o
			merged[k] = &cp
			order = append(order, k)
			return
		}
		cur.FileCount += o.FileCount
		cur.TotalSize += o.TotalSize
		if cur.Name == "" {
			cur.Name = o.Name
		}
	}
	for _, o := range mediaOwners {
		add(o)
	}
	for _, o := range recOwners {
		add(o)
	}

	owners := make([]domain.MediaOwner, 0, len(order))
	for _, k := range order {
		owners = append(owners, *merged[k])
	}
	// Size-sorted, largest first — the whole point is "what eats space".
	sort.SliceStable(owners, func(i, j int) bool { return owners[i].TotalSize > owners[j].TotalSize })

	total := int64(len(owners))
	start := min(p.Offset(), len(owners))
	end := min(start+p.Limit(), len(owners))
	page := owners[start:end]

	quota := domain.StorageQuota{}
	if s.usage != nil {
		used, uErr := s.usage.SumStorageBytes(ctx, orgID)
		if uErr != nil {
			return nil, uErr
		}
		quota.UsedBytes = used
	}
	if caller.Ent.Unlimited(domain.LimitStorageGB) {
		quota.Unlimited = true
	} else {
		quota.LimitBytes = caller.Ent.Limit(domain.LimitStorageGB) * 1024 * 1024 * 1024
	}

	return &domain.MediaOwnersResponse{
		Owners:   page,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
		Quota:    quota,
	}, nil
}

func (s *service) ListOwnerFiles(ctx context.Context, ownerKind string, ownerID *uuid.UUID, p domain.ListParams) ([]domain.OwnerFile, int64, error) {
	caller, err := requireOrgViewAny(ctx)
	if err != nil {
		return nil, 0, err
	}
	// The repo unions media + (for class owners) read-only recordings, filters,
	// sorts, and pages entirely in SQL.
	return s.repo.ListOwnerFiles(ctx, *caller.OrgID, ownerKind, ownerID, p)
}

func (s *service) ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]domain.Media, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	items, err := s.repo.ListByModel(ctx, modelType, modelID, collection)
	if err != nil {
		return nil, err
	}
	visible := items[:0]
	for _, m := range items {
		if authorizeOrgAccess(caller, &m) == nil {
			visible = append(visible, m)
		}
	}
	return visible, nil
}
