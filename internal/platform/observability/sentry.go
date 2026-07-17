package observability

import (
	"log/slog"
	"time"

	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
)

// InitSentry wires optional Sentry error reporting. It is safe to call
// unconditionally: with no SENTRY_DSN configured it logs that reporting is
// disabled and returns a no-op flush plus an empty handler slice, so the API
// runs exactly as before until keys are added later. A bad DSN degrades to
// disabled rather than crashing the process.
//
// The returned handlers MUST be registered AFTER the panic-recovery middleware:
// Sentry runs with Repanic=true so it captures the panic and then re-panics for
// Recovery to write the client response.
func InitSentry(cfg *config.Config, logger *slog.Logger) (flush func(), handlers []gin.HandlerFunc) {
	noop := func() {}

	if cfg.SentryDSN == "" {
		logger.Info("sentry disabled (no SENTRY_DSN set)")
		return noop, nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.Environment,
		Release:          cfg.SentryRelease,
		EnableTracing:    cfg.SentryTracesSampleRate > 0,
		TracesSampleRate: cfg.SentryTracesSampleRate,
	}); err != nil {
		// A bad DSN must never take the process down — degrade to disabled.
		logger.Error("sentry init failed; continuing without error reporting", "error", err)
		return noop, nil
	}

	logger.Info("sentry enabled",
		"environment", cfg.Environment,
		"tracing", cfg.SentryTracesSampleRate > 0,
	)

	flush = func() { sentry.Flush(2 * time.Second) }
	handlers = []gin.HandlerFunc{
		sentrygin.New(sentrygin.Options{Repanic: true}),
		scopeTagger(),
	}
	return flush, handlers
}

// scopeTagger stamps the request correlation id onto the Sentry scope so an
// event links back to the structured log line (which already carries
// user_id/org_id). Runs only when Sentry is enabled (it is part of the handler
// slice InitSentry returns).
func scopeTagger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if hub := sentrygin.GetHubFromContext(c); hub != nil {
			if rid := domain.RequestIDFromCtx(c.Request.Context()); rid != "" {
				hub.Scope().SetTag("request_id", rid)
			}
		}
		c.Next()
	}
}
