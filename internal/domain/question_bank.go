package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestionType string

const (
	QuestionTypeDescriptive QuestionType = "descriptive"
	QuestionTypeShortAnswer QuestionType = "short_answer"
	QuestionTypeChoice      QuestionType = "choice"
)

func (t QuestionType) Valid() bool {
	switch t {
	case QuestionTypeDescriptive, QuestionTypeShortAnswer, QuestionTypeChoice:
		return true
	}
	return false
}

type QuestionOption struct {
	ID    string  `json:"id"`
	Value string  `json:"value"`
	Score float64 `json:"score"`
}

type QuestionMetadataType string

const (
	QuestionMetadataPhoto QuestionMetadataType = "photo"
)

func (t QuestionMetadataType) Valid() bool {
	return t == QuestionMetadataPhoto
}

// QuestionMetadata attaches typed assets to a question. Photo items reference
// rows in the media table (model_type = "question").
type QuestionMetadata struct {
	Type    QuestionMetadataType `json:"type"`
	MediaID uuid.UUID            `json:"media_id"`
}

// QuestionMediaModelType is the polymorphic media association value used for
// question photos.
const QuestionMediaModelType = "question"

// QuestionPhotosCollection is the media collection name for question photos.
const QuestionPhotosCollection = "photos"

type QuestionBank struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Question struct {
	ID             uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID        `gorm:"type:uuid;not null;index" json:"organization_id"`
	BankID         uuid.UUID        `gorm:"type:uuid;not null;index" json:"bank_id"`
	Bank      *QuestionBank    `gorm:"foreignKey:BankID" json:"bank,omitempty"`
	Text      string             `gorm:"not null" json:"text"`
	Type      QuestionType       `gorm:"type:varchar(20);not null" json:"type"`
	Options   []QuestionOption   `gorm:"type:jsonb;serializer:json" json:"options"`
	Metadata  []QuestionMetadata `gorm:"type:jsonb;serializer:json" json:"metadata"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`

	// IsMultiSelectFlag is populated only by the "take" endpoint after answer
	// keys are stripped, so the client can still render multi-select choice
	// questions without seeing per-option scores.
	IsMultiSelectFlag *bool `gorm:"-" json:"is_multi_select,omitempty"`
}

// --- DTOs ---

type CreateQuestionBankDTO struct {
	Name        string `json:"name" binding:"required,min=2"`
	Description string `json:"description"`
}

type UpdateQuestionBankDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=2"`
	Description *string `json:"description"`
}

type CreateQuestionDTO struct {
	Text     string             `json:"text" binding:"required,min=1"`
	Type     QuestionType       `json:"type" binding:"required,oneof=descriptive short_answer choice"`
	Options  []QuestionOption   `json:"options"`
	Metadata []QuestionMetadata `json:"metadata"`
}

type UpdateQuestionDTO struct {
	Text     *string            `json:"text" binding:"omitempty,min=1"`
	Type     *QuestionType      `json:"type" binding:"omitempty,oneof=descriptive short_answer choice"`
	Options  []QuestionOption   `json:"options"`
	Metadata []QuestionMetadata `json:"metadata"`
}

// IsMultiSelect reports whether a choice question has more than one
// positive-score option. Multi-select questions allow several correct picks
// and sum their scores.
func (q *Question) IsMultiSelect() bool {
	if q.Type != QuestionTypeChoice {
		return false
	}
	positives := 0
	for _, o := range q.Options {
		if o.Score > 0 {
			positives++
			if positives > 1 {
				return true
			}
		}
	}
	return false
}

// MaxScore returns the highest option score. Negative-only sets return 0.
// For multi-select choice questions, returns the sum of positive scores.
func (q *Question) MaxScore() float64 {
	if q.IsMultiSelect() {
		var sum float64
		for _, o := range q.Options {
			if o.Score > 0 {
				sum += o.Score
			}
		}
		return sum
	}
	var max float64
	for _, o := range q.Options {
		if o.Score > max {
			max = o.Score
		}
	}
	return max
}

// ValidateQuestionOptions enforces option-count, value-required, and
// negative-score rules per question type.
func ValidateQuestionOptions(qType QuestionType, options []QuestionOption) error {
	if !qType.Valid() {
		return NewValidationError(map[string]string{"type": "invalid question type"})
	}
	switch qType {
	case QuestionTypeChoice:
		if len(options) < 2 {
			return NewValidationError(map[string]string{"options": "choice questions require at least 2 options"})
		}
		positives := 0
		for _, o := range options {
			if o.Score > 0 {
				positives++
			}
		}
		if positives == 0 {
			return NewValidationError(map[string]string{"options": "choice questions require at least one option with a positive score"})
		}
	case QuestionTypeShortAnswer:
		if len(options) < 1 {
			return NewValidationError(map[string]string{"options": "short_answer questions require at least 1 option"})
		}
	case QuestionTypeDescriptive:
		if len(options) < 1 {
			return NewValidationError(map[string]string{"options": "descriptive questions require at least 1 option"})
		}
	}
	for i, o := range options {
		if qType != QuestionTypeChoice && o.Score < 0 {
			return NewValidationError(map[string]string{
				fmt.Sprintf("options[%d].score", i): "negative scores are allowed only for choice questions",
			})
		}
		if qType == QuestionTypeShortAnswer && strings.TrimSpace(o.Value) == "" {
			return NewValidationError(map[string]string{
				fmt.Sprintf("options[%d].value", i): "value is required for short_answer options",
			})
		}
	}
	return nil
}

// ValidateQuestionMetadata enforces structural metadata-item invariants.
// Existence of the referenced media row is verified by the service layer.
func ValidateQuestionMetadata(items []QuestionMetadata) error {
	for i, m := range items {
		if !m.Type.Valid() {
			return NewValidationError(map[string]string{
				fmt.Sprintf("metadata[%d].type", i): "unsupported metadata type",
			})
		}
		if m.MediaID == uuid.Nil {
			return NewValidationError(map[string]string{
				fmt.Sprintf("metadata[%d].media_id", i): "media_id is required",
			})
		}
	}
	return nil
}

type ListQuestionBanksQuery struct {
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type ListQuestionsQuery struct {
	Type           *QuestionType `form:"type"`
	IncludeDeleted bool          `form:"include_deleted"`
	ListParams     ListParams    `form:"-"`
}

type AdminListQuestionBanksQuery struct {
	OrganizationID *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type AdminListQuestionsQuery struct {
	BankID         *uuid.UUID    `form:"-"`
	OrganizationID *uuid.UUID    `form:"-"`
	Type           *QuestionType `form:"type"`
	IncludeDeleted bool          `form:"include_deleted"`
	ListParams     ListParams    `form:"-"`
}

// AdminCreateQuestionBankDTO is the body for POST /admin/question-banks.
type AdminCreateQuestionBankDTO struct {
	OrganizationID uuid.UUID `json:"organization_id" binding:"required"`
	Name           string    `json:"name" binding:"required,min=2"`
	Description    string    `json:"description"`
}

// AdminUpdateQuestionBankDTO is the body for PUT /admin/question-banks/:id.
type AdminUpdateQuestionBankDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=2"`
	Description *string `json:"description"`
}

// --- Interfaces ---

type QuestionBankRepository interface {
	Create(ctx context.Context, bank *QuestionBank) error
	FindByID(ctx context.Context, id uuid.UUID) (*QuestionBank, error)
	Update(ctx context.Context, bank *QuestionBank) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, p ListParams) ([]QuestionBank, int64, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*QuestionBank, error)
	AdminList(ctx context.Context, q AdminListQuestionBanksQuery) ([]QuestionBank, int64, error)
}

type QuestionRepository interface {
	Create(ctx context.Context, question *Question) error
	FindByID(ctx context.Context, id uuid.UUID) (*Question, error)
	Update(ctx context.Context, question *Question) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBank(ctx context.Context, bankID uuid.UUID, q ListQuestionsQuery) ([]Question, int64, error)
	ListAllByBank(ctx context.Context, bankID uuid.UUID) ([]Question, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]Question, error)
	CountByBank(ctx context.Context, bankID uuid.UUID) (int64, error)
	RandomByBank(ctx context.Context, bankID uuid.UUID, count int) ([]Question, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
	AdminList(ctx context.Context, q AdminListQuestionsQuery) ([]Question, int64, error)
}

type QuestionBankService interface {
	Create(ctx context.Context, dto CreateQuestionBankDTO) (*QuestionBank, error)
	GetByID(ctx context.Context, id uuid.UUID) (*QuestionBank, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateQuestionBankDTO) (*QuestionBank, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, p ListParams) ([]QuestionBank, int64, error)

	CreateQuestion(ctx context.Context, bankID uuid.UUID, dto CreateQuestionDTO) (*Question, error)
	GetQuestion(ctx context.Context, id uuid.UUID) (*Question, error)
	UpdateQuestion(ctx context.Context, id uuid.UUID, dto UpdateQuestionDTO) (*Question, error)
	DeleteQuestion(ctx context.Context, id uuid.UUID) error
	ListQuestions(ctx context.Context, bankID uuid.UUID, q ListQuestionsQuery) ([]Question, int64, error)

	AdminList(ctx context.Context, q AdminListQuestionBanksQuery) ([]QuestionBank, int64, error)
	AdminListQuestions(ctx context.Context, q AdminListQuestionsQuery) ([]Question, int64, error)
	AdminCreate(ctx context.Context, dto AdminCreateQuestionBankDTO) (*QuestionBank, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateQuestionBankDTO) (*QuestionBank, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
	AdminHardDeleteQuestion(ctx context.Context, id uuid.UUID) error
}
