package middleware

import (
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// skipLogPaths are high-frequency infra endpoints (health/readiness probes)
// that would otherwise flood the access log with noise.
var skipLogPaths = map[string]struct{}{
	"/healthz": {},
	"/readyz":  {},
}

// sensitiveQueryKeys are query parameters whose values must never reach the
// logs (and by extension the log aggregator). Matched case-insensitively.
var sensitiveQueryKeys = map[string]struct{}{
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"password":      {},
	"secret":        {},
	"signature":     {},
	"sig":           {},
	"api_key":       {},
	"apikey":        {},
	"key":           {},
}

// redactQuery replaces the values of sensitive query parameters with a marker,
// leaving everything else intact. On a parse error it drops the query entirely
// rather than risk logging a raw secret.
func redactQuery(raw string) string {
	if raw == "" {
		return ""
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return "[unparseable]"
	}
	for k := range values {
		if _, bad := sensitiveQueryKeys[strings.ToLower(k)]; bad {
			values.Set(k, "[REDACTED]")
		}
	}
	return values.Encode()
}

func Logging(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if _, skip := skipLogPaths[path]; skip {
			c.Next()
			return
		}

		start := time.Now()
		query := redactQuery(c.Request.URL.RawQuery)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// request_id / org_id / user_id are injected by the logger's context
		// handler from the request context.
		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.Int("body_size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "request completed with errors", attrs...)
		} else if status >= 500 {
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "server error", attrs...)
		} else if status >= 400 {
			logger.LogAttrs(c.Request.Context(), slog.LevelWarn, "client error", attrs...)
		} else {
			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "request completed", attrs...)
		}
	}
}
