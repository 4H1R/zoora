package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AIGradingMode selects whether AI scores are applied to the grade or only suggested.
type AIGradingMode string

const (
	AIGradingModeApply   AIGradingMode = "apply"   // write earned_score (never over a manual grade)
	AIGradingModeSuggest AIGradingMode = "suggest" // only write suggested_score + rationale
)

// AIGradingStatus is the lifecycle of a whole "AI grade" run over a quiz.
type AIGradingStatus string

const (
	AIGradingStatusPending   AIGradingStatus = "pending"
	AIGradingStatusRunning   AIGradingStatus = "running"
	AIGradingStatusCompleted AIGradingStatus = "completed"
	AIGradingStatusFailed    AIGradingStatus = "failed"
)

// AIAnswerStatus is the per-answer AI outcome (drives row badges).
const (
	AIAnswerStatusPending = "pending"
	AIAnswerStatusScored  = "scored"
	AIAnswerStatusFailed  = "failed"
)

// GradedBy records how an answer's earned_score was set. Manual is sacred: AI
// never overwrites a manual grade.
const (
	GradedByAI     = "ai"
	GradedByManual = "manual"
)

// AIGradingJob is the durable, pollable state of one teacher-triggered run.
type AIGradingJob struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	QuizID         uuid.UUID       `gorm:"type:uuid;not null;index" json:"quiz_id"`
	CreatedBy      uuid.UUID       `gorm:"type:uuid;not null" json:"created_by"`
	Mode           AIGradingMode   `gorm:"type:varchar(10);not null" json:"mode"`
	Status         AIGradingStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Total          int             `gorm:"not null;default:0" json:"total"`
	Done           int             `gorm:"not null;default:0" json:"done"`
	Failed         int             `gorm:"not null;default:0" json:"failed"`
	Error          string          `gorm:"type:text" json:"error,omitempty"`
	CreatedAt      time.Time       `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName pins the jobs table name for GORM.
func (AIGradingJob) TableName() string { return "ai_grading_jobs" }

// StartAIGradingDTO is the request body for triggering a run.
type StartAIGradingDTO struct {
	Mode  AIGradingMode `json:"mode" binding:"required,oneof=apply suggest"`
	Force bool          `json:"force"` // re-grade answers already AI-graded (never manual)
}

// AIGradingJobRepository persists and advances grading-job rows.
type AIGradingJobRepository interface {
	Create(ctx context.Context, job *AIGradingJob) error
	FindByID(ctx context.Context, id uuid.UUID) (*AIGradingJob, error)
	// IncrementProgress atomically bumps done/failed and, when done+failed >= total,
	// flips status to completed. Safe under concurrent task completion.
	IncrementProgress(ctx context.Context, id uuid.UUID, doneDelta, failedDelta int) error
	MarkRunning(ctx context.Context, id uuid.UUID) error
}
