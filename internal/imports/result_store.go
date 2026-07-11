package imports

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

type redisResultStore struct {
	rdb *redis.Client
}

// NewRedisResultStore adapts the platform cache helpers to the ResultStore
// interface, mapping redis.Nil to domain.ErrNotFound.
func NewRedisResultStore(rdb *redis.Client) ResultStore {
	return &redisResultStore{rdb: rdb}
}

func (r *redisResultStore) Set(ctx context.Context, jobID uuid.UUID, data []byte) error {
	return cache.SetImportResult(ctx, r.rdb, jobID, data)
}

func (r *redisResultStore) Get(ctx context.Context, jobID uuid.UUID) ([]byte, error) {
	data, err := cache.GetImportResult(ctx, r.rdb, jobID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return data, nil
}
