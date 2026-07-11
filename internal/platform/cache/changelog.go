package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
)

// changelogStatusTTL bounds how long a newly published entry stays invisible in
// the badge. Published entries change very rarely, so a per-user snapshot with a
// short TTL is safe; the user's own "mark seen" invalidates explicitly so the
// badge clears immediately.
const changelogStatusTTL = 60 * time.Second

func changelogStatusKey(userID uuid.UUID) string {
	return fmt.Sprintf("changelog:status:%s", userID.String())
}

// GetChangelogStatus returns the cached changelog status for a user. Miss or
// decode failure returns an error — callers fall back to recomputing.
func GetChangelogStatus(ctx context.Context, rdb *redis.Client, userID uuid.UUID) (*domain.ChangelogStatus, error) {
	data, err := rdb.Get(ctx, changelogStatusKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}
	var st domain.ChangelogStatus
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("unmarshaling cached changelog status: %w", err)
	}
	return &st, nil
}

// SetChangelogStatus caches a user's computed changelog status.
func SetChangelogStatus(ctx context.Context, rdb *redis.Client, userID uuid.UUID, st *domain.ChangelogStatus) error {
	data, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("marshaling changelog status: %w", err)
	}
	return rdb.Set(ctx, changelogStatusKey(userID), data, changelogStatusTTL).Err()
}

// InvalidateChangelogStatus drops a user's cached status (call after the user
// marks the changelog seen).
func InvalidateChangelogStatus(ctx context.Context, rdb *redis.Client, userID uuid.UUID) error {
	return rdb.Del(ctx, changelogStatusKey(userID)).Err()
}
