package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ConnectorType string

const (
	ConnectorTelegram ConnectorType = "telegram"
	ConnectorBale     ConnectorType = "bale"
	ConnectorSMS      ConnectorType = "sms"
	ConnectorPush     ConnectorType = "push"
)

// UserConnector is one delivery endpoint a user linked. Target semantics per
// type: telegram/bale = chat ID, sms = phone (E.164-ish "09..."), push = FCM
// device token (multiple rows allowed — one per device).
type UserConnector struct {
	ID         uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	UserID     uuid.UUID     `gorm:"type:uuid;not null;index" json:"user_id"`
	Type       ConnectorType `gorm:"type:varchar(20);not null" json:"type"`
	Target     string        `gorm:"type:varchar(500);not null" json:"target"`
	VerifiedAt *time.Time    `json:"verified_at,omitempty"`
	Enabled    bool          `gorm:"not null;default:true" json:"enabled"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

func (UserConnector) TableName() string { return "user_connectors" }

// --- DTOs ---

// LinkTokenResponse starts a bot linking flow. DeepLink opens the bot with a
// one-time start token; the worker's poller completes the link.
type LinkTokenResponse struct {
	Token     string    `json:"token"`
	DeepLink  string    `json:"deep_link"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RequestSMSOTPDTO struct {
	Phone string `json:"phone" binding:"required,min=10,max=15"`
}

type VerifySMSOTPDTO struct {
	Code string `json:"code" binding:"required,len=6"`
}

type RegisterPushTokenDTO struct {
	Token string `json:"token" binding:"required,max=500"`
}

type UpdateConnectorDTO struct {
	Enabled *bool `json:"enabled" binding:"required"`
}

// --- interfaces ---

type UserConnectorRepository interface {
	Create(ctx context.Context, c *UserConnector) error
	FindByID(ctx context.Context, id uuid.UUID) (*UserConnector, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]UserConnector, error)
	// ListVerifiedEnabledByUsers returns delivery endpoints for fan-out.
	ListVerifiedEnabledByUsers(ctx context.Context, userIDs []uuid.UUID) ([]UserConnector, error)
	Update(ctx context.Context, c *UserConnector) error
	Delete(ctx context.Context, id uuid.UUID) error
	// DeleteByTypeTarget prunes dead endpoints (e.g. FCM token invalidated).
	DeleteByTypeTarget(ctx context.Context, t ConnectorType, target string) error
}

type ConnectorService interface {
	// CreateLinkToken issues a one-time bot-linking token (telegram|bale).
	CreateLinkToken(ctx context.Context, t ConnectorType) (*LinkTokenResponse, error)
	// CompleteLink is called by the worker bot poller on /start <token>.
	CompleteLink(ctx context.Context, t ConnectorType, token, chatID string) error
	RequestSMSOTP(ctx context.Context, dto RequestSMSOTPDTO) error
	VerifySMSOTP(ctx context.Context, dto VerifySMSOTPDTO) error
	RegisterPushToken(ctx context.Context, dto RegisterPushTokenDTO) (*UserConnector, error)
	List(ctx context.Context) ([]UserConnector, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateConnectorDTO) (*UserConnector, error)
	Unlink(ctx context.Context, id uuid.UUID) error
}

// --- delivery sender ports (implemented in platform/) ---

type BotSender interface {
	SendMessage(ctx context.Context, chatID string, text string) error
}

type SMSSender interface {
	// SendBulk sends one message to many receptors in a single provider call.
	SendBulk(ctx context.Context, phones []string, message string) error
	// SendOTP delivers a verification code via the provider's OTP channel.
	SendOTP(ctx context.Context, phone, code string) error
}

type PushSender interface {
	// SendMulticast returns tokens the provider reports as permanently invalid.
	SendMulticast(ctx context.Context, tokens []string, title, body, link string) (invalidTokens []string, err error)
}
