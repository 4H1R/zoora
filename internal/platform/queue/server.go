package queue

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

type Server struct {
	inner  *asynq.Server
	mux    *asynq.ServeMux
	logger *slog.Logger
}

func NewServer(redisURL string, logger *slog.Logger) (*Server, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URI for asynq server: %w", err)
	}

	srv := asynq.NewServer(opts, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			domain.QueueCritical:      6,
			domain.QueueDefault:       3,
			domain.QueueMedia:         2,
			domain.QueueNotifications: 1,
		},
		Logger: NewAsynqLogger(logger),
	})

	mux := asynq.NewServeMux()
	mux.Use(loggingMiddleware(logger))

	return &Server{inner: srv, mux: mux, logger: logger}, nil
}

// loggingMiddleware logs one line per processed task with its type, id, retry
// state, latency, and outcome — the worker-side equivalent of HTTP access logs.
// It also stamps the task id onto the context so any log the handler emits is
// correlated back to the same task.
func loggingMiddleware(logger *slog.Logger) asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			taskID, _ := asynq.GetTaskID(ctx)
			retry, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			queueName, _ := asynq.GetQueueName(ctx)

			ctx = domain.WithTaskID(ctx, taskID)
			start := time.Now()
			err := next.ProcessTask(ctx, task)

			attrs := []any{
				"task_type", task.Type(),
				"queue", queueName,
				"retry", retry,
				"max_retry", maxRetry,
				"latency_ms", time.Since(start).Milliseconds(),
			}
			if err != nil {
				logger.ErrorContext(ctx, "task failed", append(attrs, "error", err)...)
			} else {
				logger.InfoContext(ctx, "task completed", attrs...)
			}
			return err
		})
	}
}

func (s *Server) HandleFunc(pattern string, handler func(ctx context.Context, task *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Run() error {
	s.logger.Info("starting asynq worker server")
	return s.inner.Run(s.mux)
}

func (s *Server) Shutdown() {
	s.logger.Info("shutting down asynq worker server")
	s.inner.Shutdown()
}

type AsynqLogger struct {
	logger *slog.Logger
}

func NewAsynqLogger(logger *slog.Logger) *AsynqLogger {
	return &AsynqLogger{logger: logger}
}

func (l *AsynqLogger) Debug(args ...any) { l.logger.Debug(fmt.Sprint(args...)) }
func (l *AsynqLogger) Info(args ...any)  { l.logger.Info(fmt.Sprint(args...)) }
func (l *AsynqLogger) Warn(args ...any)  { l.logger.Warn(fmt.Sprint(args...)) }
func (l *AsynqLogger) Error(args ...any) { l.logger.Error(fmt.Sprint(args...)) }
func (l *AsynqLogger) Fatal(args ...any) {
	l.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}
