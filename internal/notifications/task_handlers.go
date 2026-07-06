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

// NewDeliverBotHandler processes notification:deliver-bot tasks (one telegram/
// bale message per delivery row).
func NewDeliverBotHandler(svc domain.NotificationService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var p domain.NotificationDeliverBotPayload
		if err := json.Unmarshal(task.Payload(), &p); err != nil {
			return fmt.Errorf("unmarshaling deliver-bot payload: %v: %w", err, asynq.SkipRetry)
		}
		return svc.DeliverBot(ctx, p.DeliveryID)
	}
}

// NewDeliverSMSHandler processes notification:deliver-sms tasks (one bulk
// provider call per batch of delivery rows).
func NewDeliverSMSHandler(svc domain.NotificationService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var p domain.NotificationDeliverBatchPayload
		if err := json.Unmarshal(task.Payload(), &p); err != nil {
			return fmt.Errorf("unmarshaling deliver-sms payload: %v: %w", err, asynq.SkipRetry)
		}
		return svc.DeliverSMS(ctx, p.NotificationID, p.DeliveryIDs)
	}
}

// NewDeliverPushHandler processes notification:deliver-push tasks (one FCM
// multicast per batch of delivery rows).
func NewDeliverPushHandler(svc domain.NotificationService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var p domain.NotificationDeliverBatchPayload
		if err := json.Unmarshal(task.Payload(), &p); err != nil {
			return fmt.Errorf("unmarshaling deliver-push payload: %v: %w", err, asynq.SkipRetry)
		}
		return svc.DeliverPush(ctx, p.NotificationID, p.DeliveryIDs)
	}
}
