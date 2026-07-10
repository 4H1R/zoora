package organizations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// prefixDeleter purges all S3 objects under a key prefix. *storage.Client
// satisfies this via DeleteByPrefix.
type prefixDeleter interface {
	DeleteByPrefix(ctx context.Context, prefix string) error
}

// NewCleanupHandler processes organization:cleanup — deleting every S3 object
// under a deleted org's key prefix (orgs/{org_id}/). Media rows are already
// gone via the DB org FK cascade; this reaps the underlying storage, which no
// FK covers. Idempotent: a retry re-deletes harmlessly.
func NewCleanupHandler(storage prefixDeleter) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload domain.OrganizationCleanupPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			// Malformed payload is unrecoverable — don't retry.
			return fmt.Errorf("org cleanup: unmarshal payload: %w: %w", err, asynq.SkipRetry)
		}
		return storage.DeleteByPrefix(ctx, domain.OrgStoragePrefix(payload.OrganizationID))
	}
}
