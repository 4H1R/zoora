package livesessions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

func NewAutoCloseHandler(svc domain.LiveSessionService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		return svc.AutoCloseStaleRooms(ctx)
	}
}

// NewCloseIfNoHostHandler processes the delayed, webhook-armed no-host close.
func NewCloseIfNoHostHandler(svc domain.LiveSessionService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload domain.LiveSessionCloseIfNoHostPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			// Malformed payload is unrecoverable — don't retry.
			return fmt.Errorf("close-if-no-host: unmarshal payload: %w: %w", err, asynq.SkipRetry)
		}
		return svc.CloseRoomIfNoHost(ctx, payload.RoomID)
	}
}
