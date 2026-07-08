package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestUserCacheRoundTripAndInvalidation(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	userID := uuid.New()
	orgID := uuid.New()
	roleID := uuid.New()
	cu := CachedUser{
		OrganizationID: &orgID,
		RoleID:         &roleID,
		IsAdmin:        false,
		Username:       "jane",
		Name:           "Jane Doe",
	}

	if err := SetUser(ctx, rdb, userID, cu); err != nil {
		t.Fatalf("SetUser() error = %v", err)
	}
	if ttl := server.TTL(userKey(userID)); ttl <= 0 {
		t.Fatalf("cached user TTL = %s, want positive TTL", ttl)
	}

	got, err := GetUser(ctx, rdb, userID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if got.Username != cu.Username || got.Name != cu.Name || got.IsAdmin != cu.IsAdmin {
		t.Fatalf("GetUser() = %#v, want %#v", got, cu)
	}
	if got.OrganizationID == nil || *got.OrganizationID != orgID {
		t.Fatalf("GetUser() OrganizationID = %v, want %v", got.OrganizationID, orgID)
	}
	if got.RoleID == nil || *got.RoleID != roleID {
		t.Fatalf("GetUser() RoleID = %v, want %v", got.RoleID, roleID)
	}

	if err := InvalidateUser(ctx, rdb, userID); err != nil {
		t.Fatalf("InvalidateUser() error = %v", err)
	}
	if _, err := GetUser(ctx, rdb, userID); err == nil {
		t.Fatal("GetUser() error = nil after invalidation")
	}
}

// A disabled user round-trips with its DisabledAt preserved so the auth
// middleware can reject the cached snapshot without a DB hit.
func TestUserCachePreservesDisabledAt(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	userID := uuid.New()
	disabledAt := time.Unix(1_700_000_000, 0).UTC()

	if err := SetUser(ctx, rdb, userID, CachedUser{Username: "locked", DisabledAt: &disabledAt}); err != nil {
		t.Fatalf("SetUser() error = %v", err)
	}

	got, err := GetUser(ctx, rdb, userID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if got.DisabledAt == nil || !got.DisabledAt.Equal(disabledAt) {
		t.Fatalf("GetUser() DisabledAt = %v, want %v", got.DisabledAt, disabledAt)
	}
}

func TestGetUserRejectsCorruptJSON(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	userID := uuid.New()
	if err := server.Set(userKey(userID), "{not-json"); err != nil {
		t.Fatalf("server.Set() error = %v", err)
	}

	if _, err := GetUser(ctx, rdb, userID); err == nil {
		t.Fatal("GetUser() error = nil for corrupt JSON")
	}
}
