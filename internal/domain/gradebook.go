package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type GradebookColumnType string

const (
	GradebookColumnAutoAttendance GradebookColumnType = "auto_attendance"
	GradebookColumnAutoPractice   GradebookColumnType = "auto_practice"
	GradebookColumnAutoQuiz       GradebookColumnType = "auto_quiz"
	GradebookColumnManualGrade    GradebookColumnType = "manual_grade"
	GradebookColumnManualAttendance GradebookColumnType = "manual_attendance"
	GradebookColumnManualText     GradebookColumnType = "manual_text"
)

func (t GradebookColumnType) Valid() bool {
	switch t {
	case GradebookColumnAutoAttendance, GradebookColumnAutoPractice, GradebookColumnAutoQuiz,
		GradebookColumnManualGrade, GradebookColumnManualAttendance, GradebookColumnManualText:
		return true
	}
	return false
}

func (t GradebookColumnType) IsAuto() bool {
	switch t {
	case GradebookColumnAutoAttendance, GradebookColumnAutoPractice, GradebookColumnAutoQuiz:
		return true
	}
	return false
}

type GradebookColumn struct {
	ID         uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ClassID    uuid.UUID           `gorm:"type:uuid;not null;index" json:"class_id"`
	Title      string              `gorm:"not null" json:"title"`
	Type       GradebookColumnType `gorm:"type:varchar(30);not null" json:"type"`
	SourceID   *uuid.UUID          `gorm:"type:uuid" json:"source_id,omitempty"`
	OrderIndex int                 `gorm:"not null;default:0" json:"order_index"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

type GradebookCell struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ColumnID  uuid.UUID `gorm:"type:uuid;not null;index" json:"column_id"`
	StudentID uuid.UUID `gorm:"type:uuid;not null;index" json:"student_id"`
	Value     string    `gorm:"type:text;not null;default:''" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- DTOs ---

type CreateGradebookColumnDTO struct {
	Title      string              `json:"title" binding:"required,min=1"`
	Type       GradebookColumnType `json:"type" binding:"required,oneof=auto_attendance auto_practice auto_quiz manual_grade manual_attendance manual_text"`
	SourceID   *uuid.UUID          `json:"source_id"`
	OrderIndex int                 `json:"order_index" binding:"gte=0"`
}

type UpdateGradebookColumnDTO struct {
	Title      *string `json:"title" binding:"omitempty,min=1"`
	OrderIndex *int    `json:"order_index" binding:"omitempty,gte=0"`
}

type UpsertGradebookCellDTO struct {
	StudentID uuid.UUID `json:"student_id" binding:"required"`
	Value     string    `json:"value" binding:"required"`
}

// GradebookMatrixRow represents one student's row in the gradebook grid.
type GradebookMatrixRow struct {
	StudentID uuid.UUID         `json:"student_id"`
	Student   *User             `json:"student,omitempty"`
	Cells     map[string]string `json:"cells"` // column_id -> value
}

// GradebookMatrix is the full gradebook response.
type GradebookMatrix struct {
	Columns []GradebookColumn    `json:"columns"`
	Rows    []GradebookMatrixRow `json:"rows"`
}

// ListGradebookColumnsQuery is the query for GET /classes/:id/gradebook/columns.
type ListGradebookColumnsQuery struct {
	Type       *GradebookColumnType `form:"type" binding:"omitempty,oneof=auto_attendance auto_practice auto_quiz manual_grade manual_attendance manual_text"`
	ListParams ListParams           `form:"-"`
}

// --- Interfaces ---

type GradebookColumnRepository interface {
	Create(ctx context.Context, col *GradebookColumn) error
	FindByID(ctx context.Context, id uuid.UUID) (*GradebookColumn, error)
	Update(ctx context.Context, col *GradebookColumn) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByClass(ctx context.Context, classID uuid.UUID, q ListGradebookColumnsQuery) ([]GradebookColumn, int64, error)
	ListAllByClass(ctx context.Context, classID uuid.UUID) ([]GradebookColumn, error)
}

type GradebookCellRepository interface {
	Upsert(ctx context.Context, cell *GradebookCell) error
	ListByColumns(ctx context.Context, columnIDs []uuid.UUID) ([]GradebookCell, error)
}

type GradebookService interface {
	CreateColumn(ctx context.Context, classID uuid.UUID, dto CreateGradebookColumnDTO) (*GradebookColumn, error)
	UpdateColumn(ctx context.Context, columnID uuid.UUID, dto UpdateGradebookColumnDTO) (*GradebookColumn, error)
	DeleteColumn(ctx context.Context, columnID uuid.UUID) error
	UpsertCell(ctx context.Context, classID, columnID uuid.UUID, dto UpsertGradebookCellDTO) (*GradebookCell, error)
	ListColumns(ctx context.Context, classID uuid.UUID, q ListGradebookColumnsQuery) ([]GradebookColumn, int64, error)
	GetMatrix(ctx context.Context, classID uuid.UUID) (*GradebookMatrix, error)
}
