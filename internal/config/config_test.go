package config

import (
	"testing"
	"time"
)

func TestLoadAppliesDefaultsAndParsesDurations(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("JWT_EXPIRY", "45m")
	t.Setenv("ENVIRONMENT", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want default 8080", cfg.Port)
	}
	if cfg.S3Region != "us-east-1" {
		t.Fatalf("S3Region = %q, want default us-east-1", cfg.S3Region)
	}
	if cfg.JWTExpiry != 45*time.Minute {
		t.Fatalf("JWTExpiry = %s, want 45m", cfg.JWTExpiry)
	}
	if cfg.IsDevelopment() {
		t.Fatal("IsDevelopment() = true for production environment")
	}
	if !cfg.IsProduction() {
		t.Fatal("IsProduction() = false for production environment")
	}
}

func TestLoadFailsWhenRequiredValuesAreMissing(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want required env error")
	}
}

func TestLoadCORSAllowedOriginsFromEnv(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := []string{"https://app.example.com", "https://admin.example.com"}
	if len(cfg.CORSAllowedOrigins) != len(want) {
		t.Fatalf("CORSAllowedOrigins = %v, want %v", cfg.CORSAllowedOrigins, want)
	}
	for i := range want {
		if cfg.CORSAllowedOrigins[i] != want[i] {
			t.Fatalf("origin[%d] = %q, want %q", i, cfg.CORSAllowedOrigins[i], want[i])
		}
	}
}

func TestLoadCORSAllowedOriginsDefaultsToWildcard(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "*" {
		t.Fatalf("default CORSAllowedOrigins = %v, want [*]", cfg.CORSAllowedOrigins)
	}
}

func TestLoadReadsLLMConfig(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LLM_PROVIDER", "gemini")
	t.Setenv("LLM_API_KEY", "secret")
	t.Setenv("LLM_MODEL", "gemini-2.0-flash")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LLMProvider != "gemini" || cfg.LLMAPIKey != "secret" || cfg.LLMModel != "gemini-2.0-flash" {
		t.Fatalf("llm fields not parsed: %+v", cfg)
	}
	if cfg.LLMMaxTokens != 512 {
		t.Fatalf("expected default LLMMaxTokens 512, got %d", cfg.LLMMaxTokens)
	}
	if cfg.LLMTimeout != 30*time.Second {
		t.Fatalf("expected default LLMTimeout 30s, got %s", cfg.LLMTimeout)
	}
	if cfg.LLMAIQueueConcurrency != 5 {
		t.Fatalf("expected default concurrency 5, got %d", cfg.LLMAIQueueConcurrency)
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/zoora")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("LIVEKIT_HOST", "ws://localhost:7880")
	t.Setenv("LIVEKIT_API_KEY", "key")
	t.Setenv("LIVEKIT_API_SECRET", "secret")
	t.Setenv("S3_ENDPOINT", "http://localhost:9000")
	t.Setenv("S3_BUCKET", "zoora")
	t.Setenv("S3_ACCESS_KEY", "access")
	t.Setenv("S3_SECRET_KEY", "secret")
	t.Setenv("JWT_SECRET", "jwt-secret")
}
