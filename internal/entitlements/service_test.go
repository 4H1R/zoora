package entitlements

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type mockRepo struct {
	users    int64
	storage  int64
	rooms    int64
	countErr error
}

func (m *mockRepo) GetOrgPlan(context.Context, uuid.UUID) (domain.Plan, *time.Time, error) {
	return domain.PlanFree, nil, nil
}
func (m *mockRepo) CountUsers(context.Context, uuid.UUID) (int64, error) {
	return m.users, m.countErr
}
func (m *mockRepo) SumStorageBytes(context.Context, uuid.UUID) (int64, error) {
	return m.storage, m.countErr
}
func (m *mockRepo) CountActiveLiveRooms(context.Context, uuid.UUID) (int64, error) {
	return m.rooms, m.countErr
}

var orgID = uuid.New()

func TestCheckUserLimitRejectsAtCeiling(t *testing.T) {
	svc := NewService(&mockRepo{users: 10})
	ent := domain.PlanCatalog[domain.PlanFree] // max users 10
	err := svc.CheckUserLimit(context.Background(), orgID, ent)
	if !errors.Is(err, domain.ErrPlanLimitReached) {
		t.Fatalf("expected limit reached, got %v", err)
	}
}

func TestCheckUserLimitAllowsBelowCeiling(t *testing.T) {
	svc := NewService(&mockRepo{users: 9})
	ent := domain.PlanCatalog[domain.PlanFree]
	if err := svc.CheckUserLimit(context.Background(), orgID, ent); err != nil {
		t.Fatalf("expected allow, got %v", err)
	}
}

func TestCheckUserLimitUnlimitedAlwaysAllows(t *testing.T) {
	svc := NewService(&mockRepo{users: 1_000_000})
	ent := domain.PlanCatalog[domain.PlanEnterprise] // 0 = unlimited
	if err := svc.CheckUserLimit(context.Background(), orgID, ent); err != nil {
		t.Fatal("unlimited must always allow")
	}
}

func TestCheckStorageLimitRejectsWhenAddExceeds(t *testing.T) {
	// Free = 2 GB. Used = 2 GB already, +1 byte must reject.
	svc := NewService(&mockRepo{storage: 2 * 1024 * 1024 * 1024})
	ent := domain.PlanCatalog[domain.PlanFree]
	if err := svc.CheckStorageLimit(context.Background(), orgID, ent, 1); !errors.Is(err, domain.ErrPlanLimitReached) {
		t.Fatalf("expected storage limit reached, got %v", err)
	}
}

func TestCheckStorageLimitAllowsUnderCeiling(t *testing.T) {
	svc := NewService(&mockRepo{storage: 1024})
	ent := domain.PlanCatalog[domain.PlanFree]
	if err := svc.CheckStorageLimit(context.Background(), orgID, ent, 1024); err != nil {
		t.Fatalf("expected allow, got %v", err)
	}
}

func TestCheckConcurrentRoomsLimit(t *testing.T) {
	svc := NewService(&mockRepo{rooms: 1})
	ent := domain.PlanCatalog[domain.PlanFree] // concurrent rooms 1
	if err := svc.CheckConcurrentRoomsLimit(context.Background(), orgID, ent); !errors.Is(err, domain.ErrPlanLimitReached) {
		t.Fatalf("expected rooms limit reached, got %v", err)
	}
}
