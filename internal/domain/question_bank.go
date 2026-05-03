package domain

import (
	"context"
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
	Text      string           `gorm:"not null" json:"text"`
	Type      QuestionType     `gorm:"type:varchar(20);not null" json:"type"`
	Options   []QuestionOption `gorm:"type:jsonb;serializer:json" json:"options"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`
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
	Text    string           `json:"text" binding:"required,min=1"`
	Type    QuestionType     `json:"type" binding:"required,oneof=descriptive short_answer choice"`
	Options []QuestionOption `json:"options"`
}

type UpdateQuestionDTO struct {
	Text    *string          `json:"text" binding:"omitempty,min=1"`
	Type    *QuestionType    `json:"type" binding:"omitempty,oneof=descriptive short_answer choice"`
	Options []QuestionOption `json:"options"`
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
	OrganizationID *uuid.UUID `form:"organization_id"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
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
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]Question, error)
	CountByBank(ctx context.Context, bankID uuid.UUID) (int64, error)
	RandomByBank(ctx context.Context, bankID uuid.UUID, count int) ([]Question, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
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
	AdminCreate(ctx context.Context, dto AdminCreateQuestionBankDTO) (*QuestionBank, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateQuestionBankDTO) (*QuestionBank, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
	AdminHardDeleteQuestion(ctx context.Context, id uuid.UUID) error
}
