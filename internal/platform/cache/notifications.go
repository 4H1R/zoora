package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// notifUnreadTTL is deliberately short: the unread badge is polled every ~30s
// per client, so a brief TTL bounds staleness even if a write-path invalidation
// is missed, while still absorbing the bulk of the repeated polls. Writes
// (mark-read, mark-all-read, fan-out) invalidate explicitly for exactness.
const notifUnreadTTL = 30 * time.Second

func notifUnreadKey(userID uuid.UUID) string {
	return fmt.Sprintf("notif:%s:unread", userID.String())
}

// GetUnreadCount returns the cached unread notification count for a user. Miss
// returns an error — callers fall back to the repository.
func GetUnreadCount(ctx context.Context, rdb *redis.Client, userID uuid.UUID) (int64, error) {
	return rdb.Get(ctx, notifUnreadKey(userID)).Int64()
}

// SetUnreadCount caches a user's unread count.
func SetUnreadCount(ctx context.Context, rdb *redis.Client, userID uuid.UUID, n int64) error {
	return rdb.Set(ctx, notifUnreadKey(userID), n, notifUnreadTTL).Err()
}

// InvalidateUnreadCount drops a user's cached unread count (call after the user
// reads notifications).
func InvalidateUnreadCount(ctx context.Context, rdb *redis.Client, userID uuid.UUID) error {
	return rdb.Del(ctx, notifUnreadKey(userID)).Err()
}

// InvalidateUnreadCounts drops the cached unread count for many users in one
// round-trip (call after fan-out inserts inbox rows for a batch of recipients).
func InvalidateUnreadCounts(ctx context.Context, rdb *redis.Client, userIDs []uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}
	keys := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		keys = append(keys, notifUnreadKey(id))
	}
	return rdb.Del(ctx, keys...).Err()
}
