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
	S3Endpoint         string        `env:"S3_ENDPOINT,required"`
	S3Bucket           string        `env:"S3_BUCKET,required"`
	S3AccessKey        string        `env:"S3_ACCESS_KEY,required"`
	S3SecretKey        string        `env:"S3_SECRET_KEY,required"`
	S3Region           string        `env:"S3_REGION"          envDefault:"us-east-1"`
	JWTSecret          string        `env:"JWT_SECRET,required"`
	JWTExpiry          time.Duration `env:"JWT_EXPIRY"         envDefault:"24h"`
	Environment        string        `env:"ENVIRONMENT"        envDefault:"development"`
	CORSAllowedOrigins []string      `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*"`
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
