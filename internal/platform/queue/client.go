package queue

import (
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

type Client struct {
	inner  *asynq.Client
	logger *slog.Logger
}

func NewClient(redisURL string, logger *slog.Logger) (*Client, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URI for asynq: %w", err)
	}

	client := asynq.NewClient(opts)
	logger.Info("asynq client initialized")

	return &Client{inner: client, logger: logger}, nil
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

func (c *Client) Close() error {
	return c.inner.Close()
}
