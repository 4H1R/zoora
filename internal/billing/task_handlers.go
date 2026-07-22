package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// NewGeneratePDFHandler processes invoice:generate-pdf — rendering and storing
// the receipt PDF for a paid invoice.
func NewGeneratePDFHandler(svc domain.BillingService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var p domain.InvoiceGeneratePDFPayload
		if err := json.Unmarshal(task.Payload(), &p); err != nil {
			// Malformed payload is unrecoverable — don't retry.
			return fmt.Errorf("billing generate-pdf: unmarshal: %w: %w", err, asynq.SkipRetry)
		}
		return svc.GeneratePDF(ctx, p.InvoiceID)
	}
}

// NewReminderSweepHandler processes the daily renewal + unpaid-invoice reminder
// sweep.
func NewReminderSweepHandler(svc domain.BillingService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		return svc.RunReminderSweep(ctx, time.Now())
	}
}

// NewExpireSweepHandler processes the sweep that expires stale pending invoices.
func NewExpireSweepHandler(svc domain.BillingService) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		return svc.ExpireStaleInvoices(ctx, time.Now())
	}
}
