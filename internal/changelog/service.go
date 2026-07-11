package changelog

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

const presignUploadExpiry = 15 * time.Minute

// objectStorage is the storage subset the changelog service needs for public
// media. Concrete impl is *storage.Client.
type objectStorage interface {
	PublicPresignUpload(ctx context.Context, key string, expiry time.Duration) (string, error)
	PublicURL(key string) string
	DeletePublicObject(ctx context.Context, key string) error
}

type service struct {
	repo      domain.ChangelogRepository
	mediaRepo domain.MediaRepository // media rows for changelog assets
	storage   objectStorage
	rdb       *redis.Client // nil disables status caching (unit tests)
	logger    *slog.Logger
}

// NewService builds the changelog service. mediaRepo/storage/rdb may be nil in
// unit tests that only exercise read/status paths.
func NewService(repo domain.ChangelogRepository, storage objectStorage, rdb *redis.Client, logger *slog.Logger) domain.ChangelogService {
	return &service{repo: repo, storage: storage, rdb: rdb, logger: logger}
}

// NewServiceWithMedia wires the media repository for admin media management.
func NewServiceWithMedia(repo domain.ChangelogRepository, mediaRepo domain.MediaRepository, storage objectStorage, rdb *redis.Client, logger *slog.Logger) domain.ChangelogService {
	return &service{repo: repo, mediaRepo: mediaRepo, storage: storage, rdb: rdb, logger: logger}
}

func (s *service) ListPublished(ctx context.Context, p domain.ListParams) ([]domain.ChangelogEntry, int64, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, 0, domain.ErrForbidden
	}
	return s.repo.ListPublished(ctx, p.Limit(), p.Offset())
}

func (s *service) Status(ctx context.Context) (*domain.ChangelogStatus, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	if s.rdb != nil {
		if st, err := cache.GetChangelogStatus(ctx, s.rdb, caller.UserID); err == nil {
			return st, nil
		}
	}

	seen, err := s.repo.GetLastSeen(ctx, caller.UserID)
	if err != nil {
		return nil, err
	}
	unseen, err := s.repo.CountUnseen(ctx, seen)
	if err != nil {
		return nil, err
	}
	major, err := s.repo.LatestMajorUnseen(ctx, seen)
	if err != nil {
		return nil, err
	}
	latest, err := s.repo.LatestPublished(ctx)
	if err != nil {
		return nil, err
	}
	st := &domain.ChangelogStatus{
		UnseenCount:    unseen,
		HasMajorUnseen: major != nil,
		LatestMajor:    major,
	}
	if latest != nil {
		st.CurrentVersion = latest.Version
	}

	if s.rdb != nil {
		if err := cache.SetChangelogStatus(ctx, s.rdb, caller.UserID, st); err != nil {
			s.logger.WarnContext(ctx, "caching changelog status", "error", err)
		}
	}
	return st, nil
}

func (s *service) MarkSeen(ctx context.Context) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	// Server clock — never trust a client-sent timestamp (skew).
	if err := s.repo.UpdateLastSeen(ctx, caller.UserID, time.Now()); err != nil {
		return err
	}
	// Drop the cached badge so it clears on the next poll immediately.
	if s.rdb != nil {
		if err := cache.InvalidateChangelogStatus(ctx, s.rdb, caller.UserID); err != nil {
			s.logger.WarnContext(ctx, "invalidating changelog status", "error", err)
		}
	}
	return nil
}
