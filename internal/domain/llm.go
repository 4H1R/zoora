package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// LLMRole is the message author role in an LLM request.
type LLMRole string

const (
	LLMRoleSystem LLMRole = "system"
	LLMRoleUser   LLMRole = "user"
)

// LLMMessage is one turn in a conversation sent to the model.
type LLMMessage struct {
	Role    LLMRole
	Content string
}

// LLMRequest is a provider-agnostic generation request. System holds trusted
// instructions; Messages hold (untrusted) user content — keeping them in
// separate channels is the primary prompt-injection defense.
type LLMRequest struct {
	System      string
	Messages    []LLMMessage
	JSONMode    bool   // ask the provider for structured JSON output when supported
	JSONSchema  string // optional JSON-schema string for native structured output; "" = none
	MaxTokens   int
	Temperature float32

	// Metering context — recorded by the platform layer on every call.
	Feature        string
	OrganizationID uuid.UUID
}

// LLMUsage reports token accounting returned by the provider.
type LLMUsage struct {
	PromptTokens     int
	CompletionTokens int
}

// LLMResponse is the provider-agnostic result of a generation.
type LLMResponse struct {
	Text  string
	Usage LLMUsage
	Model string
}

// LLM is the pluggable connector every AI feature calls. One implementation is
// active per deploy (selected by config); adapters translate to each provider.
type LLM interface {
	Generate(ctx context.Context, req LLMRequest) (LLMResponse, error)
}

// AIUsageEvent is one metered LLM call, written for per-tenant cost visibility.
type AIUsageEvent struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID   uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Feature          string    `gorm:"type:varchar(40);not null" json:"feature"`
	Provider         string    `gorm:"type:varchar(20);not null" json:"provider"`
	Model            string    `gorm:"type:varchar(60);not null" json:"model"`
	PromptTokens     int       `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens int       `gorm:"not null;default:0" json:"completion_tokens"`
	CostMicros       int64     `gorm:"not null;default:0" json:"cost_micros"`
	CreatedAt        time.Time `gorm:"not null;default:now()" json:"created_at"`
}

// TableName pins the metering table name for GORM.
func (AIUsageEvent) TableName() string { return "ai_usage_events" }

// AIUsageRecorder persists metering events. Implemented in internal/ai.
type AIUsageRecorder interface {
	Record(ctx context.Context, ev AIUsageEvent) error
}
