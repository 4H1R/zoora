package cache

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ImportResultTTL bounds how long a generated-passwords result file survives.
const ImportResultTTL = 24 * time.Hour

func importResultKey(jobID uuid.UUID) string {
	return "import:result:" + jobID.String()
}

// SetImportResult stores a finished import's result xlsx bytes. Redis is the
// ONLY store for this file — it contains plaintext generated passwords.
func SetImportResult(ctx context.Context, rdb *redis.Client, jobID uuid.UUID, data []byte) error {
	return rdb.Set(ctx, importResultKey(jobID), data, ImportResultTTL).Err()
}

// GetImportResult returns the stored result bytes. Miss/expiry surfaces as
// redis.Nil — callers in feature packages map that to domain.ErrNotFound.
func GetImportResult(ctx context.Context, rdb *redis.Client, jobID uuid.UUID) ([]byte, error) {
	return rdb.Get(ctx, importResultKey(jobID)).Bytes()
}
