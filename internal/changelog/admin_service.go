package changelog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) AdminList(ctx context.Context, p domain.ListParams) ([]domain.ChangelogEntry, int64, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, 0, err
	}
	return s.repo.AdminList(ctx, p.Limit(), p.Offset())
}

func (s *service) AdminGet(ctx context.Context, id uuid.UUID) (*domain.ChangelogEntry, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *service) AdminCreate(ctx context.Context, dto domain.CreateChangelogDTO) (*domain.ChangelogEntry, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	e := &domain.ChangelogEntry{
		Version: dto.Version,
		TitleEn: dto.TitleEn,
		TitleFa: dto.TitleFa,
		BodyEn:  dto.BodyEn,
		BodyFa:  dto.BodyFa,
		IsMajor: dto.IsMajor,
		// PublishedAt stays nil → draft.
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *service) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.UpdateChangelogDTO) (*domain.ChangelogEntry, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Version != nil {
		e.Version = dto.Version
	}
	if dto.TitleEn != nil {
		e.TitleEn = *dto.TitleEn
	}
	if dto.TitleFa != nil {
		e.TitleFa = *dto.TitleFa
	}
	if dto.BodyEn != nil {
		e.BodyEn = *dto.BodyEn
	}
	if dto.BodyFa != nil {
		e.BodyFa = *dto.BodyFa
	}
	if dto.IsMajor != nil {
		e.IsMajor = *dto.IsMajor
	}
	// NOTE: published_at is intentionally NOT touched here — edits are silent.
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *service) AdminPublish(ctx context.Context, id uuid.UUID) (*domain.ChangelogEntry, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Set published_at only on first publish; re-publishing a live entry is a
	// no-op so it never re-badges users.
	if e.PublishedAt == nil {
		now := time.Now()
		e.PublishedAt = &now
		if err := s.repo.Update(ctx, e); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (s *service) AdminUnpublish(ctx context.Context, id uuid.UUID) (*domain.ChangelogEntry, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	e.PublishedAt = nil // explicit admin act; republish will reset the timestamp
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *service) AdminDelete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	// Purge attached public media (rows + objects) before dropping the entry.
	if s.mediaRepo != nil && s.storage != nil {
		items, err := s.mediaRepo.ListByModel(ctx, domain.MediaModelChangelog, id, "")
		if err == nil {
			for i := range items {
				m := items[i]
				if dErr := s.storage.DeletePublicObject(ctx, m.S3Key()); dErr != nil && s.logger != nil {
					s.logger.Error("changelog media cleanup: object", "media_id", m.ID.String(), "error", dErr)
				}
				_ = s.mediaRepo.Delete(ctx, m.ID)
			}
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *service) AdminPresignMedia(ctx context.Context, dto domain.ChangelogMediaPresignDTO) (*domain.ChangelogMediaPresignResponse, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	entryID, err := uuid.Parse(dto.EntryID)
	if err != nil {
		return nil, domain.NewValidationError(map[string]string{"entry_id": "must be a valid UUID"})
	}
	// The entry must exist so media can attach (and be cleaned up) by its id.
	if _, err := s.repo.FindByID(ctx, entryID); err != nil {
		return nil, err
	}
	m := &domain.Media{
		OrganizationID:   nil, // platform-global → S3Key() yields un-prefixed path
		ModelType:        domain.MediaModelChangelog,
		ModelID:          entryID,
		CollectionName:   "",
		Name:             dto.FileName,
		FileName:         dto.FileName,
		MimeType:         dto.MimeType,
		Disk:             "s3-public",
		Size:             dto.Size,
		CustomProperties: json.RawMessage(`{}`),
	}
	if err := s.mediaRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	uploadURL, err := s.storage.PublicPresignUpload(ctx, m.S3Key(), presignUploadExpiry)
	if err != nil {
		return nil, err
	}
	return &domain.ChangelogMediaPresignResponse{
		UploadURL: uploadURL,
		PublicURL: s.storage.PublicURL(m.S3Key()),
		MediaID:   m.ID,
	}, nil
}
