package imports

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// NewProcessHandler handles TypeImportProcess. Bad payloads skip retry;
// everything else is delegated (the service itself converts processing
// problems into a failed job rather than a retryable task error).
func NewProcessHandler(svc domain.ImportService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload domain.ImportProcessPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("import process: unmarshal payload: %w: %w", err, asynq.SkipRetry)
		}
		return svc.ProcessJob(ctx, payload)
	}
}
