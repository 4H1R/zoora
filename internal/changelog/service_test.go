package changelog

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type mockRepo struct {
	lastSeen    *time.Time
	unseen      int64
	latestMajor *domain.ChangelogEntry
	latestPub   *domain.ChangelogEntry
	updatedSeen *time.Time
}

func (m *mockRepo) Create(context.Context, *domain.ChangelogEntry) error { return nil }
func (m *mockRepo) Update(context.Context, *domain.ChangelogEntry) error { return nil }
func (m *mockRepo) FindByID(context.Context, uuid.UUID) (*domain.ChangelogEntry, error) {
	return nil, nil
}
func (m *mockRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (m *mockRepo) ListPublished(context.Context, int, int) ([]domain.ChangelogEntry, int64, error) {
	return nil, 0, nil
}

func (m *mockRepo) AdminList(context.Context, int, int) ([]domain.ChangelogEntry, int64, error) {
	return nil, 0, nil
}

func (m *mockRepo) LatestPublished(context.Context) (*domain.ChangelogEntry, error) {
	return m.latestPub, nil
}
func (m *mockRepo) CountUnseen(context.Context, *time.Time) (int64, error) { return m.unseen, nil }
func (m *mockRepo) LatestMajorUnseen(context.Context, *time.Time) (*domain.ChangelogEntry, error) {
	return m.latestMajor, nil
}

func (m *mockRepo) GetLastSeen(context.Context, uuid.UUID) (*time.Time, error) {
	return m.lastSeen, nil
}

func (m *mockRepo) UpdateLastSeen(_ context.Context, _ uuid.UUID, t time.Time) error {
	m.updatedSeen = &t
	return nil
}

func ctxWithCaller() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New()})
}

func TestStatusReportsUnseenAndMajor(t *testing.T) {
	v := "v2.4.0"
	major := &domain.ChangelogEntry{ID: uuid.New(), IsMajor: true, TitleEn: "Big"}
	repo := &mockRepo{
		unseen:      3,
		latestMajor: major,
		latestPub:   &domain.ChangelogEntry{Version: &v},
	}
	svc := NewService(repo, nil, nil, nil)

	got, err := svc.Status(ctxWithCaller())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UnseenCount != 3 {
		t.Fatalf("UnseenCount = %d, want 3", got.UnseenCount)
	}
	if !got.HasMajorUnseen || got.LatestMajor == nil {
		t.Fatalf("expected major unseen, got %+v", got)
	}
	if got.CurrentVersion == nil || *got.CurrentVersion != "v2.4.0" {
		t.Fatalf("CurrentVersion = %v, want v2.4.0", got.CurrentVersion)
	}
}

func TestStatusRequiresCaller(t *testing.T) {
	svc := NewService(&mockRepo{}, nil, nil, nil)
	if _, err := svc.Status(context.Background()); err == nil {
		t.Fatal("expected ErrForbidden without caller")
	}
}
