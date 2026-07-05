package media

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// NewCleanupHandler processes media:cleanup — purging a polymorphic collection's
// rows and S3 objects (e.g. live-room slides after a room finishes).
func NewCleanupHandler(svc domain.MediaService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload domain.MediaCleanupPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			// Malformed payload is unrecoverable — don't retry.
			return fmt.Errorf("media cleanup: unmarshal payload: %w: %w", err, asynq.SkipRetry)
		}
		return svc.CleanupByModel(ctx, payload.ModelType, payload.ModelID, payload.CollectionName)
	}
}
