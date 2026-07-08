package billing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// asynqClient is the subset of the platform queue client billing needs. It is
// satisfied by *queue.Client (Enqueue(task, opts...)).
type asynqClient interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// queueEnqueuer adapts the Asynq client into the service's enqueuer interface,
// enqueuing the invoice PDF-generation task after a payment is settled.
type queueEnqueuer struct {
	client asynqClient
}

func NewQueueEnqueuer(client asynqClient) *queueEnqueuer {
	return &queueEnqueuer{client: client}
}

func (q *queueEnqueuer) EnqueuePDF(ctx context.Context, invoiceID uuid.UUID) error {
	payload, err := json.Marshal(domain.InvoiceGeneratePDFPayload{InvoiceID: invoiceID})
	if err != nil {
		return fmt.Errorf("billing.queue.EnqueuePDF marshal: %w", err)
	}
	task := asynq.NewTask(domain.TypeInvoiceGeneratePDF, payload)
	if _, err := q.client.Enqueue(task, asynq.MaxRetry(5)); err != nil {
		return fmt.Errorf("billing.queue.EnqueuePDF: %w", err)
	}
	return nil
}

var _ enqueuer = (*queueEnqueuer)(nil)
