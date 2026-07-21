package changelog

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type stubRepo struct {
	entry *domain.ChangelogEntry
	saved *domain.ChangelogEntry
}

func (s *stubRepo) Create(_ context.Context, e *domain.ChangelogEntry) error {
	e.ID = uuid.New()
	s.entry = e
	return nil
}

func (s *stubRepo) Update(_ context.Context, e *domain.ChangelogEntry) error {
	s.saved = e
	s.entry = e
	return nil
}

func (s *stubRepo) FindByID(context.Context, uuid.UUID) (*domain.ChangelogEntry, error) {
	return s.entry, nil
}
func (s *stubRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (s *stubRepo) ListPublished(context.Context, int, int) ([]domain.ChangelogEntry, int64, error) {
	return nil, 0, nil
}

func (s *stubRepo) AdminList(context.Context, int, int) ([]domain.ChangelogEntry, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) LatestPublished(context.Context) (*domain.ChangelogEntry, error) { return nil, nil }
func (s *stubRepo) CountUnseen(context.Context, *time.Time) (int64, error)          { return 0, nil }
func (s *stubRepo) LatestMajorUnseen(context.Context, *time.Time) (*domain.ChangelogEntry, error) {
	return nil, nil
}
func (s *stubRepo) GetLastSeen(context.Context, uuid.UUID) (*time.Time, error) { return nil, nil }
func (s *stubRepo) UpdateLastSeen(context.Context, uuid.UUID, time.Time) error { return nil }

func adminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
}

func TestPublishIsIdempotentOnTimestamp(t *testing.T) {
	repo := &stubRepo{}
	svc := NewService(repo, nil, nil, nil)
	first := time.Now().Add(-time.Hour)
	repo.entry = &domain.ChangelogEntry{ID: uuid.New(), PublishedAt: &first}

	got, err := svc.AdminPublish(adminCtx(), repo.entry.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if got.PublishedAt == nil || !got.PublishedAt.Equal(first) {
		t.Fatalf("published_at moved: got %v want %v", got.PublishedAt, first)
	}
}

func TestPublishRequiresAdmin(t *testing.T) {
	repo := &stubRepo{}
	svc := NewService(repo, nil, nil, nil)
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New()}) // not admin
	if _, err := svc.AdminPublish(ctx, uuid.New()); err == nil {
		t.Fatal("expected ErrForbidden for non-admin")
	}
}
