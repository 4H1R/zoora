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
	svc := NewService(&mockRepo{users: 50})
	ent := domain.PlanCatalog[domain.PlanFree] // max users 50
	err := svc.CheckUserLimit(context.Background(), orgID, ent)
	if !errors.Is(err, domain.ErrPlanLimitReached) {
		t.Fatalf("expected limit reached, got %v", err)
	}
}

func TestCheckUserLimitAllowsBelowCeiling(t *testing.T) {
	svc := NewService(&mockRepo{users: 49})
	ent := domain.PlanCatalog[domain.PlanFree]
	if err := svc.CheckUserLimit(context.Background(), orgID, ent); err != nil {
		t.Fatalf("expected allow, got %v", err)
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

// panicOnCountUsersRepo embeds mockRepo but panics if CountUsers is invoked,
// used to prove CheckUserLimitN short-circuits before hitting the repo.
type panicOnCountUsersRepo struct{ mockRepo }

func (m *panicOnCountUsersRepo) CountUsers(context.Context, uuid.UUID) (int64, error) {
	panic("CountUsers should not be called when n<=0")
}

func TestCheckUserLimitN(t *testing.T) {
	ent := domain.PlanCatalog[domain.PlanFree] // Free plan has a finite LimitMaxUsers
	limit := ent.Limit(domain.LimitMaxUsers)

	t.Run("fits", func(t *testing.T) {
		svc := NewService(&mockRepo{users: limit - 3})
		if err := svc.CheckUserLimitN(context.Background(), orgID, ent, 3); err != nil {
			t.Fatalf("expected allow, got %v", err)
		}
	})

	t.Run("exceeds", func(t *testing.T) {
		svc := NewService(&mockRepo{users: limit - 3})
		err := svc.CheckUserLimitN(context.Background(), orgID, ent, 4)
		if !errors.Is(err, domain.ErrPlanLimitReached) {
			t.Fatalf("expected limit reached, got %v", err)
		}
	})

	t.Run("zero new users always ok", func(t *testing.T) {
		svc := NewService(&panicOnCountUsersRepo{}) // CountUsers must not be called
		if err := svc.CheckUserLimitN(context.Background(), orgID, ent, 0); err != nil {
			t.Fatalf("expected allow, got %v", err)
		}
	})
}
