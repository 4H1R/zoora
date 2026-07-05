package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/4H1R/zoora/internal/domain"
)

// contextHandler enriches every record with correlation fields pulled from the
// context (request id, task id, org id, user id) so callers never thread them
// through individual log calls. Only logs made via the *Context slog methods
// (or LogAttrs) carry a real context and get enriched.
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := domain.RequestIDFromCtx(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	if id := domain.TaskIDFromCtx(ctx); id != "" {
		r.AddAttrs(slog.String("task_id", id))
	}
	if c, ok := domain.CallerFromCtx(ctx); ok {
		r.AddAttrs(slog.String("user_id", c.UserID.String()))
		if c.OrgID != nil {
			r.AddAttrs(slog.String("org_id", c.OrgID.String()))
		}
	} else if hc, ok := domain.HostContextFromCtx(ctx); ok && hc.OrgID != nil {
		r.AddAttrs(slog.String("org_id", hc.OrgID.String()))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs and WithGroup must re-wrap the derived handler; otherwise the
// embedded slog.Handler's promoted methods would return a bare handler and
// silently drop context enrichment for any logger built via With/WithGroup.
func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{Handler: h.Handler.WithGroup(name)}
}

func parseLevel(level string, isDevelopment bool) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
	if isDevelopment {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

// New builds the application logger: a human-readable text handler in
// development, JSON otherwise, wrapped so context correlation fields are added
// automatically. level overrides the default threshold (debug in development,
// info otherwise); pass "" to keep the default.
func New(isDevelopment bool, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level, isDevelopment)}

	var base slog.Handler
	if isDevelopment {
		base = slog.NewTextHandler(os.Stdout, opts)
	} else {
		base = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(contextHandler{Handler: base})
}
