package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port             string `env:"PORT"               envDefault:"8080"`
	DatabaseURL      string `env:"DATABASE_URL,required"`
	RedisURL         string `env:"REDIS_URL,required"`
	LiveKitHost      string `env:"LIVEKIT_HOST,required"`
	LiveKitPublicURL string `env:"LIVEKIT_PUBLIC_URL"`
	LiveKitAPIKey    string `env:"LIVEKIT_API_KEY,required"`
	LiveKitSecret    string `env:"LIVEKIT_API_SECRET,required"`
	// LiveRoomHostGracePeriod is how long a live room may stay open after its
	// last host leaves before it is auto-closed. Drives both the webhook-driven
	// delayed close task and the periodic safety-net sweep.
	LiveRoomHostGracePeriod time.Duration `env:"LIVE_ROOM_HOST_GRACE_PERIOD" envDefault:"15m"`
	S3Endpoint              string        `env:"S3_ENDPOINT,required"`
	// S3PublicEndpoint is the browser-facing host used to sign upload/download
	// URLs. The SDK client talks to S3Endpoint (internal, e.g. http://rustfs:9000)
	// so boot-time calls don't depend on the public TLS edge. Falls back to
	// S3Endpoint when unset (dev, where the two are the same host).
	S3PublicEndpoint string `env:"S3_PUBLIC_ENDPOINT"`
	S3Bucket         string `env:"S3_BUCKET,required"`
	// S3PublicBucket holds anonymously-readable assets (changelog media) served
	// to browsers via permanent, non-expiring URLs. Falls back to S3Bucket when
	// unset, but in that case objects are NOT public — set it in every real env.
	S3PublicBucket string        `env:"S3_PUBLIC_BUCKET"`
	S3AccessKey    string        `env:"S3_ACCESS_KEY,required"`
	S3SecretKey    string        `env:"S3_SECRET_KEY,required"`
	S3Region       string        `env:"S3_REGION"          envDefault:"us-east-1"`
	JWTSecret      string        `env:"JWT_SECRET,required"`
	JWTExpiry      time.Duration `env:"JWT_EXPIRY"         envDefault:"24h"`
	Environment    string        `env:"ENVIRONMENT"        envDefault:"development"`
	// LogLevel overrides the log threshold (debug/info/warn/error). Empty falls
	// back to debug in development, info in production.
	LogLevel           string   `env:"LOG_LEVEL"`
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*"`
	// BaseDomain is the apex the app is served under. Host parsing strips this
	// suffix to recover the tenant/admin subdomain label. Dev: "localhost".
	BaseDomain string `env:"BASE_DOMAIN" envDefault:"localhost"`
	// AdminSubdomain is the reserved label that routes to the platform-admin scope.
	AdminSubdomain string `env:"ADMIN_SUBDOMAIN" envDefault:"admin"`
	// NotificationSendRatePerHour caps how many notifications one non-admin
	// sender may create per hour. 0 disables the limit.
	NotificationSendRatePerHour int `env:"NOTIFICATION_SEND_RATE_PER_HOUR" envDefault:"10"`

	// --- notification connectors (all optional; empty disables the channel) ---
	TelegramBotToken     string `env:"TELEGRAM_BOT_TOKEN"`
	TelegramBotUsername  string `env:"TELEGRAM_BOT_USERNAME"`
	TelegramProxyURL     string `env:"TELEGRAM_PROXY_URL"`
	BaleBotToken         string `env:"BALE_BOT_TOKEN"`
	BaleBotUsername      string `env:"BALE_BOT_USERNAME"`
	BaleProxyURL         string `env:"BALE_PROXY_URL"`
	KavenegarAPIKey      string `env:"KAVENEGAR_API_KEY"`
	KavenegarSender      string `env:"KAVENEGAR_SENDER"`
	KavenegarOTPTemplate string `env:"KAVENEGAR_OTP_TEMPLATE"`
	FCMCredentialsFile   string `env:"FCM_CREDENTIALS_FILE"`

	// --- billing / zarinpal ---
	ZarinpalMerchantID      string `env:"ZARINPAL_MERCHANT_ID"`
	ZarinpalSandbox         bool   `env:"ZARINPAL_SANDBOX" envDefault:"true"`
	ZarinpalCallbackBaseURL string `env:"ZARINPAL_CALLBACK_BASE_URL"`
	// --- invoice issuer (seller block on the PDF) ---
	InvoiceIssuerName       string `env:"INVOICE_ISSUER_NAME" envDefault:"Zoora"`
	InvoiceIssuerEconomicID string `env:"INVOICE_ISSUER_ECONOMIC_ID"`
	InvoiceIssuerAddress    string `env:"INVOICE_ISSUER_ADDRESS"`
	InvoiceIssuerPhone      string `env:"INVOICE_ISSUER_PHONE"`
	AppBaseURL              string `env:"APP_BASE_URL" envDefault:"http://localhost:5173"`
	// Remote Chromium CDP URL (e.g. http://chrome:9222) for headless PDF receipt
	// rendering. Empty launches a local headless Chromium via the exec allocator.
	ChromeRemoteURL string `env:"CHROME_REMOTE_URL"`
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
