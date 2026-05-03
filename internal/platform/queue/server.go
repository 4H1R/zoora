package queue

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
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
			"critical": 6,
			"default":  3,
		},
		Logger: NewAsynqLogger(logger),
	})

	mux := asynq.NewServeMux()

	return &Server{inner: srv, mux: mux, logger: logger}, nil
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

// AsynqLogger adapts slog.Logger to asynq's Logger interface.
type AsynqLogger struct {
	logger *slog.Logger
}

func NewAsynqLogger(logger *slog.Logger) *AsynqLogger {
	return &AsynqLogger{logger: logger}
}

func (l *AsynqLogger) Debug(args ...interface{}) { l.logger.Debug(fmt.Sprint(args...)) }
func (l *AsynqLogger) Info(args ...interface{})  { l.logger.Info(fmt.Sprint(args...)) }
func (l *AsynqLogger) Warn(args ...interface{})  { l.logger.Warn(fmt.Sprint(args...)) }
func (l *AsynqLogger) Error(args ...interface{}) { l.logger.Error(fmt.Sprint(args...)) }
func (l *AsynqLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}
