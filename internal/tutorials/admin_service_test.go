package tutorials

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type stubRepo struct {
	entry       *domain.Tutorial
	saved       *domain.Tutorial
	maxPosition int
	reordered   []uuid.UUID
}

func (s *stubRepo) Create(_ context.Context, tu *domain.Tutorial) error {
	tu.ID = uuid.New()
	s.entry = tu
	return nil
}
func (s *stubRepo) Update(_ context.Context, tu *domain.Tutorial) error {
	s.saved = tu
	s.entry = tu
	return nil
}
func (s *stubRepo) FindByID(context.Context, uuid.UUID) (*domain.Tutorial, error) {
	return s.entry, nil
}
func (s *stubRepo) Delete(context.Context, uuid.UUID) error                  { return nil }
func (s *stubRepo) ListPublished(context.Context) ([]domain.Tutorial, error) { return nil, nil }
func (s *stubRepo) AdminList(context.Context) ([]domain.Tutorial, error)     { return nil, nil }
func (s *stubRepo) MaxPosition(context.Context) (int, error)                 { return s.maxPosition, nil }
func (s *stubRepo) Reorder(_ context.Context, ids []uuid.UUID) error {
	s.reordered = ids
	return nil
}

func adminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
}

func TestCreateAppendsToEnd(t *testing.T) {
	repo := &stubRepo{maxPosition: 4}
	svc := NewService(repo, nil)

	got, err := svc.AdminCreate(adminCtx(), domain.CreateTutorialDTO{TitleEn: "Intro", AparatHash: "abc"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got.Position != 5 {
		t.Fatalf("position: got %d want 5 (max+1)", got.Position)
	}
	if got.PublishedAt != nil {
		t.Fatal("new tutorial should be a draft")
	}
}

func TestPublishIsIdempotentOnTimestamp(t *testing.T) {
	repo := &stubRepo{}
	svc := NewService(repo, nil)
	first := time.Now().Add(-time.Hour)
	repo.entry = &domain.Tutorial{ID: uuid.New(), PublishedAt: &first}

	got, err := svc.AdminPublish(adminCtx(), repo.entry.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if got.PublishedAt == nil || !got.PublishedAt.Equal(first) {
		t.Fatalf("published_at moved: got %v want %v", got.PublishedAt, first)
	}
}

func TestReorderParsesIDs(t *testing.T) {
	repo := &stubRepo{}
	svc := NewService(repo, nil)
	a, b := uuid.New(), uuid.New()

	if err := svc.AdminReorder(adminCtx(), domain.ReorderTutorialsDTO{IDs: []string{a.String(), b.String()}}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
	if len(repo.reordered) != 2 || repo.reordered[0] != a || repo.reordered[1] != b {
		t.Fatalf("reordered ids mismatch: %v", repo.reordered)
	}
}

func TestReorderRejectsBadID(t *testing.T) {
	svc := NewService(&stubRepo{}, nil)
	if err := svc.AdminReorder(adminCtx(), domain.ReorderTutorialsDTO{IDs: []string{"not-a-uuid"}}); err == nil {
		t.Fatal("expected validation error for bad uuid")
	}
}

func TestAdminActionsRequireAdmin(t *testing.T) {
	svc := NewService(&stubRepo{}, nil)
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New()}) // not admin
	if _, err := svc.AdminCreate(ctx, domain.CreateTutorialDTO{TitleEn: "x", AparatHash: "y"}); err == nil {
		t.Fatal("expected ErrForbidden for non-admin")
	}
}
