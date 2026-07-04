package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port               string        `env:"PORT"               envDefault:"8080"`
	DatabaseURL        string        `env:"DATABASE_URL,required"`
	RedisURL           string        `env:"REDIS_URL,required"`
	LiveKitHost        string        `env:"LIVEKIT_HOST,required"`
	LiveKitPublicURL   string        `env:"LIVEKIT_PUBLIC_URL"`
	LiveKitAPIKey      string        `env:"LIVEKIT_API_KEY,required"`
	LiveKitSecret      string        `env:"LIVEKIT_API_SECRET,required"`
	// LiveRoomHostGracePeriod is how long a live room may stay open after its
	// last host leaves before it is auto-closed. Drives both the webhook-driven
	// delayed close task and the periodic safety-net sweep.
	LiveRoomHostGracePeriod time.Duration `env:"LIVE_ROOM_HOST_GRACE_PERIOD" envDefault:"15m"`
	S3Endpoint         string        `env:"S3_ENDPOINT,required"`
	// S3PublicEndpoint is the browser-facing host used to sign upload/download
	// URLs. The SDK client talks to S3Endpoint (internal, e.g. http://rustfs:9000)
	// so boot-time calls don't depend on the public TLS edge. Falls back to
	// S3Endpoint when unset (dev, where the two are the same host).
	S3PublicEndpoint   string        `env:"S3_PUBLIC_ENDPOINT"`
	S3Bucket           string        `env:"S3_BUCKET,required"`
	S3AccessKey        string        `env:"S3_ACCESS_KEY,required"`
	S3SecretKey        string        `env:"S3_SECRET_KEY,required"`
	S3Region           string        `env:"S3_REGION"          envDefault:"us-east-1"`
	JWTSecret          string        `env:"JWT_SECRET,required"`
	JWTExpiry          time.Duration `env:"JWT_EXPIRY"         envDefault:"24h"`
	Environment        string        `env:"ENVIRONMENT"        envDefault:"development"`
	CORSAllowedOrigins []string      `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*"`
	// BaseDomain is the apex the app is served under. Host parsing strips this
	// suffix to recover the tenant/admin subdomain label. Dev: "localhost".
	BaseDomain string `env:"BASE_DOMAIN" envDefault:"localhost"`
	// AdminSubdomain is the reserved label that routes to the platform-admin scope.
	AdminSubdomain string `env:"ADMIN_SUBDOMAIN" envDefault:"admin"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
