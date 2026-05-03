package livesessions

import (
	"context"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

func NewAutoCloseHandler(svc domain.LiveSessionService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		return svc.AutoCloseStaleRooms(ctx)
	}
}
