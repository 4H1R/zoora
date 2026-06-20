package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNewReturnsUsableLogger(t *testing.T) {
	for _, isDevelopment := range []bool{true, false} {
		l := New(isDevelopment)
		if l == nil {
			t.Fatalf("New(%v) returned nil", isDevelopment)
		}
	}
}

func TestLoggerLevelsMatchEnvironmentHandlers(t *testing.T) {
	var devBuf bytes.Buffer
	dev := slog.New(slog.NewTextHandler(&devBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dev.Debug("debug-message")
	if !strings.Contains(devBuf.String(), "debug-message") {
		t.Fatal("debug-level text handler did not write debug message")
	}

	var prodBuf bytes.Buffer
	prod := slog.New(slog.NewJSONHandler(&prodBuf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if prod.Handler().Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("production-style JSON handler should not enable debug level")
	}
	prod.Info("info-message")
	if !strings.Contains(prodBuf.String(), `"msg":"info-message"`) {
		t.Fatalf("JSON handler output = %q, want info message", prodBuf.String())
	}
}
