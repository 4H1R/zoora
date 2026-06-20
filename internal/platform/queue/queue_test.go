package queue

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNewClientAndServerRejectInvalidRedisURI(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	if _, err := NewClient("://bad", logger); err == nil {
		t.Fatal("NewClient() error = nil for invalid Redis URI")
	}
	if _, err := NewServer("://bad", logger); err == nil {
		t.Fatal("NewServer() error = nil for invalid Redis URI")
	}
}

func TestAsynqLoggerWritesMessages(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAsynqLogger(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	logger.Debug("debug", " ", "message")
	logger.Info("info", " ", "message")
	logger.Warn("warn", " ", "message")
	logger.Error("error", " ", "message")

	got := buf.String()
	for _, want := range []string{"debug message", "info message", "warn message", "error message"} {
		if !strings.Contains(got, want) {
			t.Fatalf("logger output %q does not contain %q", got, want)
		}
	}
}
