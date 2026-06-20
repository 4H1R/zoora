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
