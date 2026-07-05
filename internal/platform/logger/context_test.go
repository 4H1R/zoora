package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func TestContextHandlerInjectsCorrelationFields(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	l := slog.New(contextHandler{Handler: base})

	orgID := uuid.New()
	userID := uuid.New()
	ctx := domain.WithRequestID(context.Background(), "req-123")
	ctx = domain.WithTaskID(ctx, "task-456")
	ctx = domain.WithCaller(ctx, domain.Caller{UserID: userID, OrgID: &orgID})

	l.InfoContext(ctx, "hello")

	out := buf.String()
	for _, want := range []string{`"request_id":"req-123"`, `"task_id":"task-456"`, `"user_id":"` + userID.String() + `"`, `"org_id":"` + orgID.String() + `"`} {
		if !strings.Contains(out, want) {
			t.Errorf("log output missing %s\ngot: %s", want, out)
		}
	}
}

func TestContextHandlerSurvivesWith(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	// Derive via With/WithGroup: enrichment must still apply.
	l := slog.New(contextHandler{Handler: base}).With("component", "test").WithGroup("g")

	ctx := domain.WithRequestID(context.Background(), "req-789")
	l.InfoContext(ctx, "hello")

	if !strings.Contains(buf.String(), `"request_id":"req-789"`) {
		t.Errorf("With/WithGroup dropped context enrichment; got: %s", buf.String())
	}
}

func TestContextHandlerNoContextIsClean(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	l := slog.New(contextHandler{Handler: base})

	l.Info("no-ctx")
	if strings.Contains(buf.String(), "request_id") {
		t.Errorf("expected no correlation fields without context, got: %s", buf.String())
	}
}

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		dev  bool
		want slog.Level
	}{
		{"debug", false, slog.LevelDebug},
		{"WARN", false, slog.LevelWarn},
		{"error", true, slog.LevelError},
		{"", true, slog.LevelDebug},
		{"", false, slog.LevelInfo},
		{"garbage", false, slog.LevelInfo},
	}
	for _, c := range cases {
		if got := parseLevel(c.in, c.dev); got != c.want {
			t.Errorf("parseLevel(%q, %v) = %v, want %v", c.in, c.dev, got, c.want)
		}
	}
}
