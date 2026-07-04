package queue

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

type Client struct {
	inner     *asynq.Client
	inspector *asynq.Inspector
	logger    *slog.Logger
}

func NewClient(redisURL string, logger *slog.Logger) (*Client, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URI for asynq: %w", err)
	}

	client := asynq.NewClient(opts)
	inspector := asynq.NewInspector(opts)
	logger.Info("asynq client initialized")

	return &Client{inner: client, inspector: inspector, logger: logger}, nil
}

func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	info, err := c.inner.Enqueue(task, opts...)
	if err != nil {
		c.logger.Error("failed to enqueue task", "type", task.Type(), "error", err)
		return nil, fmt.Errorf("enqueuing task %s: %w", task.Type(), err)
	}
	c.logger.Info("task enqueued", "type", task.Type(), "id", info.ID)
	return info, nil
}

// Cancel removes a pending/scheduled task by ID from the given queue. A task
// that has already run (or never existed) is treated as success — callers use
// this to best-effort defuse a scheduled task, so a missing task is the goal
// state, not an error.
func (c *Client) Cancel(queue, taskID string) error {
	err := c.inspector.DeleteTask(queue, taskID)
	if err == nil || errors.Is(err, asynq.ErrTaskNotFound) || errors.Is(err, asynq.ErrQueueNotFound) {
		return nil
	}
	c.logger.Error("failed to cancel task", "queue", queue, "id", taskID, "error", err)
	return fmt.Errorf("canceling task %s: %w", taskID, err)
}

func (c *Client) Close() error {
	if err := c.inspector.Close(); err != nil {
		c.logger.Error("failed to close asynq inspector", "error", err)
	}
	return c.inner.Close()
}
