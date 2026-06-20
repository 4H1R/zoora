package livekit

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	livekitproto "github.com/livekit/protocol/livekit"

	"github.com/4H1R/zoora/internal/config"
)

func TestNewClientPublicURLFallbackAndOverride(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{
		LiveKitHost:    "ws://livekit:7880",
		LiveKitAPIKey:  "api-key",
		LiveKitSecret:  "secret",
	}

	client := NewClient(cfg, logger)
	if client.Host() != cfg.LiveKitHost {
		t.Fatalf("Host() = %q, want %q", client.Host(), cfg.LiveKitHost)
	}
	if client.PublicURL() != cfg.LiveKitHost {
		t.Fatalf("PublicURL() = %q, want host fallback", client.PublicURL())
	}

	cfg.LiveKitPublicURL = "wss://public.example.test"
	client = NewClient(cfg, logger)
	if client.PublicURL() != cfg.LiveKitPublicURL {
		t.Fatalf("PublicURL() = %q, want configured public URL", client.PublicURL())
	}
}

func TestGenerateTokenCreatesJoinTokenAndRejectsInvalidSecrets(t *testing.T) {
	client := &Client{apiKey: "api-key", apiSecret: strings.Repeat("s", 32)}

	token, err := client.GenerateToken(
		"room-1",
		"user-1",
		"Alice",
		[]livekitproto.TrackSource{livekitproto.TrackSource_MICROPHONE, livekitproto.TrackSource_CAMERA},
		true,
	)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if strings.Count(token, ".") != 2 {
		t.Fatalf("token = %q, want JWT with three segments", token)
	}

	client.apiSecret = ""
	if _, err := client.GenerateToken("room-1", "user-1", "Alice", nil, false); err == nil {
		t.Fatal("GenerateToken() error = nil for empty secret")
	}
}
