package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// NewFanoutHandler processes notification:fanout tasks. Malformed payloads
// skip retry; resolution/insert errors retry with Asynq backoff.
func NewFanoutHandler(svc domain.NotificationService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var p domain.NotificationFanoutPayload
		if err := json.Unmarshal(task.Payload(), &p); err != nil {
			return fmt.Errorf("unmarshaling fanout payload: %v: %w", err, asynq.SkipRetry)
		}
		return svc.Fanout(ctx, p.NotificationID)
	}
}
