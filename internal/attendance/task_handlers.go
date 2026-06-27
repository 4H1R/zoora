package attendance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// NewAutoMarkHandler returns an Asynq handler that runs session-scoped live
// auto-mark using the org's configured threshold. Enqueued when a live room
// finishes.
func NewAutoMarkHandler(svc domain.AttendanceService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload domain.AttendanceAutoMarkPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("attendance auto-mark: unmarshal payload: %w", err)
		}
		_, err := svc.AutoMarkSessionLive(ctx, payload.ClassID, payload.SessionID)
		return err
	}
}
