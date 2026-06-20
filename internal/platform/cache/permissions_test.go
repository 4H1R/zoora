package cache

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestRolePermissionCacheRoundTripAndInvalidation(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	roleID := uuid.New()
	perms := []string{"users:view", "classes:create"}

	if err := SetRolePermissions(ctx, rdb, roleID, perms); err != nil {
		t.Fatalf("SetRolePermissions() error = %v", err)
	}
	if ttl := server.TTL(rolePermissionKey(roleID)); ttl <= 0 {
		t.Fatalf("cached permissions TTL = %s, want positive TTL", ttl)
	}

	got, err := GetRolePermissions(ctx, rdb, roleID)
	if err != nil {
		t.Fatalf("GetRolePermissions() error = %v", err)
	}
	if len(got) != len(perms) || got[0] != perms[0] || got[1] != perms[1] {
		t.Fatalf("GetRolePermissions() = %#v, want %#v", got, perms)
	}

	if err := InvalidateRolePermissions(ctx, rdb, roleID); err != nil {
		t.Fatalf("InvalidateRolePermissions() error = %v", err)
	}
	if _, err := GetRolePermissions(ctx, rdb, roleID); err == nil {
		t.Fatal("GetRolePermissions() error = nil after invalidation")
	}
}

func TestGetRolePermissionsRejectsCorruptJSON(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	roleID := uuid.New()
	if err := server.Set(rolePermissionKey(roleID), "{not-json"); err != nil {
		t.Fatalf("server.Set() error = %v", err)
	}

	if _, err := GetRolePermissions(ctx, rdb, roleID); err == nil {
		t.Fatal("GetRolePermissions() error = nil for corrupt JSON")
	}
}
