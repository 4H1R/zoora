package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const permissionCacheTTL = 5 * time.Minute

func rolePermissionKey(roleID uuid.UUID) string {
	return fmt.Sprintf("role:%s:permissions", roleID.String())
}

func GetRolePermissions(ctx context.Context, rdb *redis.Client, roleID uuid.UUID) ([]string, error) {
	data, err := rdb.Get(ctx, rolePermissionKey(roleID)).Bytes()
	if err != nil {
		return nil, err
	}

	var perms []string
	if err := json.Unmarshal(data, &perms); err != nil {
		return nil, fmt.Errorf("unmarshaling cached role permissions: %w", err)
	}
	return perms, nil
}

func SetRolePermissions(ctx context.Context, rdb *redis.Client, roleID uuid.UUID, perms []string) error {
	data, err := json.Marshal(perms)
	if err != nil {
		return fmt.Errorf("marshaling role permissions: %w", err)
	}
	return rdb.Set(ctx, rolePermissionKey(roleID), data, permissionCacheTTL).Err()
}

func InvalidateRolePermissions(ctx context.Context, rdb *redis.Client, roleID uuid.UUID) error {
	return rdb.Del(ctx, rolePermissionKey(roleID)).Err()
}
