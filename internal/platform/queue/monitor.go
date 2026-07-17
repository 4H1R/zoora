package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

// QueueInfoProvider is the slice of Client the health monitor needs. Kept as an
// interface so the handler is testable without a live Redis inspector.
type QueueInfoProvider interface {
	QueueInfos() ([]*asynq.QueueInfo, error)
}

// NewHealthCheckHandler returns an Asynq handler that inspects every queue and
// warns when tasks have accumulated in the archived (dead-letter) or retry sets
// — the signal that jobs are failing past their retry budget and need a human.
// Registered on a periodic schedule; emits a WARN per affected queue plus one
// INFO summary line so the totals are always queryable.
func NewHealthCheckHandler(provider QueueInfoProvider, logger *slog.Logger) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		infos, err := provider.QueueInfos()
		if err != nil {
			return fmt.Errorf("inspecting queues for health check: %w", err)
		}

		var totalArchived, totalRetry int
		for _, info := range infos {
			totalArchived += info.Archived
			totalRetry += info.Retry
			if info.Archived > 0 {
				logger.WarnContext(ctx, "dead-lettered tasks present",
					"queue", info.Queue,
					"archived", info.Archived,
					"retry", info.Retry,
					"failed_today", info.Failed,
				)
			}
		}

		logger.InfoContext(ctx, "queue health",
			"archived_total", totalArchived,
			"retry_total", totalRetry,
		)
		return nil
	}
}
