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
	ID           string     `json:"id"`
	Value        string     `json:"value"`
	Score        float64    `json:"score"`
	ImageMediaID *uuid.UUID `json:"image_media_id,omitempty"`

	// SystemImageMediaID references the anti-cheat image rendered by the worker
	// from Value when the owning quiz renders as image. Server-owned: clients
	// cannot set it, and the take endpoint blanks Value while keeping this so the
	// student sees only the image. Nil until the render task completes.
	SystemImageMediaID *uuid.UUID `json:"system_image_media_id,omitempty"`
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

// QuestionOptionPhotosCollection is the media collection name for per-option
// images on choice questions.
const QuestionOptionPhotosCollection = "option-photos"

// QuestionSystemPhotosCollection and QuestionOptionSystemPhotosCollection are
// the media collections for anti-cheat images the worker renders from a
// question's body text and each option's value. Kept distinct from the
// user-uploaded "photos"/"option-photos" collections so the two never collide
// and cleanup can target only the generated set.
const (
	QuestionSystemPhotosCollection       = "system-photos"
	QuestionOptionSystemPhotosCollection = "system-option-photos"
)

// ImageRenderStatus tracks the lifecycle of a question's anti-cheat images.
type ImageRenderStatus string

const (
	ImageRenderStatusNone    ImageRenderStatus = "none"    // not an image question
	ImageRenderStatusPending ImageRenderStatus = "pending" // render task enqueued / running
	ImageRenderStatusReady   ImageRenderStatus = "ready"   // images generated, exam can start
	ImageRenderStatusFailed  ImageRenderStatus = "failed"  // render failed, needs re-save
)

func (s ImageRenderStatus) Valid() bool {
	switch s {
	case ImageRenderStatusNone, ImageRenderStatusPending, ImageRenderStatusReady, ImageRenderStatusFailed:
		return true
	}
	return false
}

// QuestionBankStatus tracks a bank's copy lifecycle. Banks created directly are
// 'ready'; a bank created by redeeming a share code starts as 'copying' while
// the worker clones questions + media, then flips to 'ready' (or 'failed').
type QuestionBankStatus string

const (
	QuestionBankStatusReady   QuestionBankStatus = "ready"
	QuestionBankStatusCopying QuestionBankStatus = "copying"
	QuestionBankStatusFailed  QuestionBankStatus = "failed"
)

type QuestionBank struct {
	ID             uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID          `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string             `gorm:"not null" json:"name"`
	Description    string             `json:"description"`
	Status         QuestionBankStatus `gorm:"type:varchar(20);not null;default:'ready'" json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	DeletedAt      gorm.DeletedAt     `gorm:"index" json:"-"`
}

// QuestionBankShareCode is a bank's redeemable share code: multi-use until it
// expires or is revoked; at most one non-revoked code exists per bank. Redeeming
// clones the bank into the redeemer's org as an independent copy.
type QuestionBankShareCode struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	BankID         uuid.UUID     `gorm:"type:uuid;not null;index" json:"bank_id"`
	Bank           *QuestionBank `gorm:"foreignKey:BankID" json:"-"`
	OrganizationID uuid.UUID     `gorm:"type:uuid;not null" json:"organization_id"`
	Code           string        `gorm:"type:varchar(32);not null;uniqueIndex" json:"code"`
	CreatedBy      uuid.UUID     `gorm:"type:uuid;not null" json:"created_by"`
	ExpiresAt      *time.Time    `json:"expires_at,omitempty"`
	RevokedAt      *time.Time    `json:"-"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// Active reports whether the code can still be redeemed at the given time.
func (c *QuestionBankShareCode) Active(now time.Time) bool {
	if c.RevokedAt != nil {
		return false
	}
	if c.ExpiresAt != nil && c.ExpiresAt.Before(now) {
		return false
	}
	return true
}

type Question struct {
	ID             uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID          `gorm:"type:uuid;not null;index" json:"organization_id"`
	BankID         uuid.UUID          `gorm:"type:uuid;not null;index" json:"bank_id"`
	Bank           *QuestionBank      `gorm:"foreignKey:BankID" json:"bank,omitempty"`
	Text           string             `gorm:"not null" json:"text"`
	Type           QuestionType       `gorm:"type:varchar(20);not null" json:"type"`
	Options        []QuestionOption   `gorm:"type:jsonb;serializer:json" json:"options"`
	Metadata       []QuestionMetadata `gorm:"type:jsonb;serializer:json" json:"metadata"`

	// ModelAnswer is the teacher's reference answer for descriptive questions.
	// Used only to compute the advisory similarity signal shown during manual
	// grading — never affects the score, never sent to students.
	ModelAnswer string `gorm:"not null;default:''" json:"model_answer"`

	// Negative-marking default for this question (Layer 1). choice-only.
	NegativeMarkMode NegativeMarkMode `gorm:"type:varchar(20);not null;default:'none'" json:"negative_mark_mode"`
	NegativeValue    float64          `gorm:"not null;default:0" json:"negative_value"`
	WrongsPerPoint   int              `gorm:"not null;default:0" json:"wrongs_per_point"`

	// MinSeconds is the teacher-declared minimum expected time to answer this
	// question. 0 = no expectation. Advisory only: answers faster than this are
	// flagged for review, never rejected.
	MinSeconds int `gorm:"not null;default:0" json:"min_seconds"`

	// ImageRenderStatus tracks whether this question's anti-cheat images have
	// been rendered. The render decision lives on the QUIZ (Quiz.RenderAsImage);
	// the rendered images are cached here and reused across every quiz that uses
	// the question. A status other than 'none' also marks the question as
	// participating in rendering, so an edit to its text re-renders the images.
	// A quiz cannot start while any of its questions is not yet 'ready'.
	ImageRenderStatus ImageRenderStatus `gorm:"type:varchar(20);not null;default:'none'" json:"image_render_status"`
	// SystemImageMediaID references the rendered body image. Server-owned; nil
	// until the render task completes.
	SystemImageMediaID *uuid.UUID `gorm:"type:uuid" json:"system_image_media_id,omitempty"`
	// SystemImageContentHash fingerprints the text last rendered into the
	// anti-cheat images. The worker skips re-rendering when the current content
	// still hashes to this value and the images are already 'ready', avoiding
	// redundant CPU and S3 writes on re-enqueues. Server-owned; not sent to
	// clients.
	SystemImageContentHash string `gorm:"type:varchar(64);not null;default:''" json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// IsMultiSelectFlag is populated only by the "take" endpoint after answer
	// keys are stripped, so the client can still render multi-select choice
	// questions without seeing per-option scores.
	IsMultiSelectFlag *bool `gorm:"-" json:"is_multi_select,omitempty"`

	// NegativeConfig is the transient, score-free effective negative-marking
	// config attached by the "take" endpoint so the client can show penalties
	// without leaking the answer key. Never persisted.
	NegativeConfig *NegativeMarkConfig `gorm:"-" json:"negative_config,omitempty"`
}

type CreateQuestionBankDTO struct {
	Name        string `json:"name" binding:"required,min=2"`
	Description string `json:"description"`
}

type UpdateQuestionBankDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=2"`
	Description *string `json:"description"`
}

type CreateQuestionDTO struct {
	Text             string             `json:"text" binding:"required,min=1"`
	Type             QuestionType       `json:"type" binding:"required,oneof=descriptive short_answer choice"`
	Options          []QuestionOption   `json:"options"`
	ModelAnswer      string             `json:"model_answer"`
	Metadata         []QuestionMetadata `json:"metadata"`
	NegativeMarkMode NegativeMarkMode   `json:"negative_mark_mode"`
	NegativeValue    float64            `json:"negative_value"`
	WrongsPerPoint   int                `json:"wrongs_per_point"`
	MinSeconds       int                `json:"min_seconds" binding:"omitempty,gte=0"`
}

type UpdateQuestionDTO struct {
	Text             *string            `json:"text" binding:"omitempty,min=1"`
	Type             *QuestionType      `json:"type" binding:"omitempty,oneof=descriptive short_answer choice"`
	Options          []QuestionOption   `json:"options"`
	ModelAnswer      *string            `json:"model_answer"`
	Metadata         []QuestionMetadata `json:"metadata"`
	NegativeMarkMode *NegativeMarkMode  `json:"negative_mark_mode"`
	NegativeValue    *float64           `json:"negative_value"`
	WrongsPerPoint   *int               `json:"wrongs_per_point"`
	MinSeconds       *int               `json:"min_seconds" binding:"omitempty,gte=0"`
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
// Descriptive questions carry a single score-holder option, so the max is its
// score — the point value the grader marks the free-text answer out of.
func (q *Question) MaxScore() float64 {
	if q.Type == QuestionTypeDescriptive {
		var sum float64
		for _, o := range q.Options {
			if o.Score > 0 {
				sum += o.Score
			}
		}
		return sum
	}
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
		if qType != QuestionTypeChoice && o.ImageMediaID != nil {
			return NewValidationError(map[string]string{
				fmt.Sprintf("options[%d].image_media_id", i): "option images are allowed only for choice questions",
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

// GenerateShareCodeDTO creates (or replaces) a bank's share code. A nil
// ExpiresInDays means the code never expires (until revoked).
type GenerateShareCodeDTO struct {
	ExpiresInDays *int `json:"expires_in_days" binding:"omitempty,gte=1,lte=365"`
}

type RedeemShareCodeDTO struct {
	Code string `json:"code" binding:"required,min=4,max=32"`
}

// ShareCodePreview is what a prospective redeemer sees before cloning: enough
// to decide, nothing org-identifying.
type ShareCodePreview struct {
	BankName      string     `json:"bank_name"`
	Description   string     `json:"description"`
	QuestionCount int64      `json:"question_count"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

type AdminCreateQuestionBankDTO struct {
	OrganizationID uuid.UUID `json:"organization_id" binding:"required"`
	Name           string    `json:"name" binding:"required,min=2"`
	Description    string    `json:"description"`
}

type AdminUpdateQuestionBankDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=2"`
	Description *string `json:"description"`
}

type QuestionBankRepository interface {
	Create(ctx context.Context, bank *QuestionBank) error
	FindByID(ctx context.Context, id uuid.UUID) (*QuestionBank, error)
	Update(ctx context.Context, bank *QuestionBank) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, p ListParams) ([]QuestionBank, int64, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*QuestionBank, error)
	AdminList(ctx context.Context, q AdminListQuestionBanksQuery) ([]QuestionBank, int64, error)

	CreateShareCode(ctx context.Context, code *QuestionBankShareCode) error
	FindShareCodeByCode(ctx context.Context, code string) (*QuestionBankShareCode, error)
	// FindActiveShareCodeByBank returns the bank's single non-revoked code
	// (which may still be expired — callers check Active()).
	FindActiveShareCodeByBank(ctx context.Context, bankID uuid.UUID) (*QuestionBankShareCode, error)
	RevokeActiveShareCodesByBank(ctx context.Context, bankID uuid.UUID, at time.Time) error
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

	GenerateShareCode(ctx context.Context, bankID uuid.UUID, dto GenerateShareCodeDTO) (*QuestionBankShareCode, error)
	GetShareCode(ctx context.Context, bankID uuid.UUID) (*QuestionBankShareCode, error)
	RevokeShareCode(ctx context.Context, bankID uuid.UUID) error
	PreviewShareCode(ctx context.Context, code string) (*ShareCodePreview, error)
	// RedeemShareCode clones the code's bank into the caller's org: it creates a
	// 'copying' shell bank, enqueues the copy task, and returns the shell.
	RedeemShareCode(ctx context.Context, dto RedeemShareCodeDTO) (*QuestionBank, error)

	AdminList(ctx context.Context, q AdminListQuestionBanksQuery) ([]QuestionBank, int64, error)
	AdminListQuestions(ctx context.Context, q AdminListQuestionsQuery) ([]Question, int64, error)
	AdminCreate(ctx context.Context, dto AdminCreateQuestionBankDTO) (*QuestionBank, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateQuestionBankDTO) (*QuestionBank, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
	AdminHardDeleteQuestion(ctx context.Context, id uuid.UUID) error
}
