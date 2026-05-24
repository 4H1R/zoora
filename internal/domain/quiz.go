package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizRuleType string

const (
	QuizRuleTypeManual QuizRuleType = "manual"
	QuizRuleTypeRandom QuizRuleType = "random"
)

func (t QuizRuleType) Valid() bool {
	switch t {
	case QuizRuleTypeManual, QuizRuleTypeRandom:
		return true
	}
	return false
}

type Quiz struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User             *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ClassID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"class_id"`
	Class            *Class         `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	Title            string         `gorm:"not null" json:"title"`
	Description      string         `json:"description"`
	DurationMinutes  int            `gorm:"not null" json:"duration_minutes"`
	TotalScore       float64        `gorm:"not null;default:0" json:"total_score"`
	NoBackNavigation bool           `gorm:"not null;default:false" json:"no_back_navigation"`
	ShuffleQuestions bool           `gorm:"not null;default:false" json:"shuffle_questions"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type QuizRule struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	QuizID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"quiz_id"`
	Quiz        *Quiz        `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	Type        QuizRuleType `gorm:"type:varchar(20);not null" json:"type"`
	BankID      *uuid.UUID   `gorm:"type:uuid" json:"bank_id,omitempty"`
	Bank        *QuestionBank `gorm:"foreignKey:BankID" json:"bank,omitempty"`
	QuestionIDs []uuid.UUID  `gorm:"type:jsonb;serializer:json" json:"question_ids,omitempty"`
	Count       int          `gorm:"not null;default:0" json:"count"`
	IsDynamic   bool         `gorm:"not null;default:false" json:"is_dynamic"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type QuizRoom struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	QuizID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"quiz_id"`
	Quiz           *Quiz      `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	ClassSessionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"class_session_id"`
	ClassSession   *ClassSession `gorm:"foreignKey:ClassSessionID" json:"class_session,omitempty"`
	StartedAt      *time.Time `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// --- DTOs ---

type CreateQuizDTO struct {
	ClassID          uuid.UUID `json:"class_id" binding:"required"`
	Title            string    `json:"title" binding:"required,min=2"`
	Description      string    `json:"description"`
	DurationMinutes  int       `json:"duration_minutes" binding:"required,gt=0"`
	NoBackNavigation bool      `json:"no_back_navigation"`
	ShuffleQuestions bool      `json:"shuffle_questions"`
}

type UpdateQuizDTO struct {
	Title            *string `json:"title" binding:"omitempty,min=2"`
	Description      *string `json:"description"`
	DurationMinutes  *int    `json:"duration_minutes" binding:"omitempty,gt=0"`
	NoBackNavigation *bool   `json:"no_back_navigation"`
	ShuffleQuestions *bool   `json:"shuffle_questions"`
}

type CreateQuizRuleDTO struct {
	Type        QuizRuleType `json:"type" binding:"required,oneof=manual random"`
	BankID      *uuid.UUID   `json:"bank_id"`
	QuestionIDs []uuid.UUID  `json:"question_ids"`
	Count       int          `json:"count" binding:"gte=0"`
	IsDynamic   bool         `json:"is_dynamic"`
}

type UpdateQuizRuleDTO struct {
	Type        *QuizRuleType `json:"type" binding:"omitempty,oneof=manual random"`
	BankID      *uuid.UUID    `json:"bank_id"`
	QuestionIDs []uuid.UUID   `json:"question_ids"`
	Count       *int          `json:"count" binding:"omitempty,gte=0"`
	IsDynamic   *bool         `json:"is_dynamic"`
}

type CreateQuizRoomDTO struct {
	ClassSessionID uuid.UUID  `json:"class_session_id" binding:"required"`
	StartedAt      *time.Time `json:"started_at" binding:"required"`
	EndedAt        *time.Time `json:"ended_at" binding:"required"`
}

func (d CreateQuizRoomDTO) Validate() error {
	if d.StartedAt == nil || d.EndedAt == nil {
		return NewValidationError(map[string]string{"window": "started_at and ended_at are required"})
	}
	if !d.EndedAt.After(*d.StartedAt) {
		return NewValidationError(map[string]string{"window": "ended_at must be after started_at"})
	}
	return nil
}

type ListQuizzesQuery struct {
	ClassID        *uuid.UUID `form:"-"`
	ClassSessionID *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type AdminListQuizzesQuery struct {
	ClassID        *uuid.UUID `form:"-"`
	ClassSessionID *uuid.UUID `form:"-"`
	UserID         *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type ListQuizRulesQuery struct {
	ListParams ListParams `form:"-"`
}

type ListQuizRoomsQuery struct {
	QuizID     *uuid.UUID `form:"-"`
	ListParams ListParams `form:"-"`
}

// --- Submission ---

type SubmissionStatus string

const (
	SubmissionStatusInProgress SubmissionStatus = "in_progress"
	SubmissionStatusSubmitted  SubmissionStatus = "submitted"
	SubmissionStatusGraded     SubmissionStatus = "graded"
)

func (s SubmissionStatus) Valid() bool {
	switch s {
	case SubmissionStatusInProgress, SubmissionStatusSubmitted, SubmissionStatusGraded:
		return true
	}
	return false
}

type SubmissionAnswer struct {
	QuestionID        uuid.UUID `json:"question_id"`
	SelectedOptionIDs []string  `json:"selected_option_ids,omitempty"`
	Value             string    `json:"value,omitempty"`
	EarnedScore       float64   `json:"earned_score"`
	SpentSeconds      int       `json:"spent_seconds"`
}

type QuizSubmission struct {
	ID          uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	QuizID      uuid.UUID          `gorm:"type:uuid;not null;index" json:"quiz_id"`
	Quiz        *Quiz              `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	UserID      uuid.UUID          `gorm:"type:uuid;not null;index" json:"user_id"`
	User        *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Status      SubmissionStatus   `gorm:"type:varchar(20);not null;default:'in_progress'" json:"status"`
	Answers     []SubmissionAnswer `gorm:"type:jsonb;serializer:json" json:"answers"`
	TotalScore  float64            `gorm:"not null;default:0" json:"total_score"`
	StartedAt   time.Time          `gorm:"not null" json:"started_at"`
	SubmittedAt *time.Time         `json:"submitted_at"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// IsRoomOpen returns true when the quiz room window contains now.
// StartedAt/EndedAt define the scheduled availability window.
// A nil EndedAt is treated as open-ended (manual close).
func (r *QuizRoom) IsRoomOpen() bool {
	return r.IsRoomOpenAt(time.Now())
}

func (r *QuizRoom) IsRoomOpenAt(t time.Time) bool {
	if r.StartedAt == nil {
		return false
	}
	if t.Before(*r.StartedAt) {
		return false
	}
	if r.EndedAt != nil && !t.Before(*r.EndedAt) {
		return false
	}
	return true
}

type StartQuizSubmissionDTO struct {
	QuizRoomID uuid.UUID `json:"quiz_room_id" binding:"required"`
}

type SubmitAnswerDTO struct {
	QuestionID        uuid.UUID `json:"question_id" binding:"required"`
	SelectedOptionIDs []string  `json:"selected_option_ids"`
	Value             string    `json:"value"`
	SpentSeconds      int       `json:"spent_seconds" binding:"gte=0"`
}

type SubmitQuizDTO struct {
	Answers []SubmitAnswerDTO `json:"answers" binding:"required,dive"`
}

type GradeAnswerDTO struct {
	QuestionID  uuid.UUID `json:"question_id" binding:"required"`
	EarnedScore float64   `json:"earned_score" binding:"gte=0"`
}

type GradeSubmissionDTO struct {
	Grades []GradeAnswerDTO `json:"grades" binding:"required,dive"`
}

type ListSubmissionsQuery struct {
	UserID     *uuid.UUID `form:"-"`
	Status     *string    `form:"status"`
	ListParams ListParams `form:"-"`
}

// SubmissionGracePeriod is the extra time allowed beyond duration_minutes
// before a submission is rejected outright.
const SubmissionGracePeriod = 30 // seconds

// --- Scoping ---

type QuizListScope struct {
	All            bool
	OrganizationID *uuid.UUID
	OwnerID        *uuid.UUID
	MemberUserID   *uuid.UUID
	ClassID        *uuid.UUID
	ClassSessionID *uuid.UUID
	IncludeDeleted bool
}

// --- Interfaces ---

type QuizRepository interface {
	Create(ctx context.Context, quiz *Quiz) error
	FindByID(ctx context.Context, id uuid.UUID) (*Quiz, error)
	Update(ctx context.Context, quiz *Quiz) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope QuizListScope, p ListParams) ([]Quiz, int64, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*Quiz, error)
	AdminList(ctx context.Context, q AdminListQuizzesQuery) ([]Quiz, int64, error)
}

type QuizRuleRepository interface {
	Create(ctx context.Context, rule *QuizRule) error
	FindByID(ctx context.Context, id uuid.UUID) (*QuizRule, error)
	Update(ctx context.Context, rule *QuizRule) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByQuiz(ctx context.Context, quizID uuid.UUID, p ListParams) ([]QuizRule, int64, error)
}

type QuizRoomRepository interface {
	Create(ctx context.Context, room *QuizRoom) error
	FindByID(ctx context.Context, id uuid.UUID) (*QuizRoom, error)
	Update(ctx context.Context, room *QuizRoom) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByQuiz(ctx context.Context, quizID uuid.UUID, p ListParams) ([]QuizRoom, int64, error)
	ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]QuizRoom, error)
	FindOpenByQuizID(ctx context.Context, quizID uuid.UUID) (*QuizRoom, error)
}

type QuizSubmissionRepository interface {
	Create(ctx context.Context, sub *QuizSubmission) error
	FindByID(ctx context.Context, id uuid.UUID) (*QuizSubmission, error)
	Update(ctx context.Context, sub *QuizSubmission) error
	FindByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*QuizSubmission, error)
	ListByQuiz(ctx context.Context, quizID uuid.UUID, q ListSubmissionsQuery) ([]QuizSubmission, int64, error)
}

type QuizService interface {
	Create(ctx context.Context, dto CreateQuizDTO) (*Quiz, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Quiz, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateQuizDTO) (*Quiz, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, q ListQuizzesQuery) ([]Quiz, int64, error)

	CreateRule(ctx context.Context, quizID uuid.UUID, dto CreateQuizRuleDTO) (*QuizRule, error)
	GetRule(ctx context.Context, id uuid.UUID) (*QuizRule, error)
	UpdateRule(ctx context.Context, id uuid.UUID, dto UpdateQuizRuleDTO) (*QuizRule, error)
	DeleteRule(ctx context.Context, id uuid.UUID) error
	ListRules(ctx context.Context, quizID uuid.UUID, q ListQuizRulesQuery) ([]QuizRule, int64, error)

	CreateRoom(ctx context.Context, quizID uuid.UUID, dto CreateQuizRoomDTO) (*QuizRoom, error)
	GetRoom(ctx context.Context, id uuid.UUID) (*QuizRoom, error)
	StartRoom(ctx context.Context, id uuid.UUID) (*QuizRoom, error)
	EndRoom(ctx context.Context, id uuid.UUID) (*QuizRoom, error)
	ListRooms(ctx context.Context, quizID uuid.UUID, q ListQuizRoomsQuery) ([]QuizRoom, int64, error)

	ListQuestionsForTaking(ctx context.Context, quizID uuid.UUID) ([]Question, error)
	StartSubmission(ctx context.Context, quizID uuid.UUID, dto StartQuizSubmissionDTO) (*QuizSubmission, error)
	SubmitQuiz(ctx context.Context, submissionID uuid.UUID, dto SubmitQuizDTO) (*QuizSubmission, error)
	GetSubmission(ctx context.Context, id uuid.UUID) (*QuizSubmission, error)
	ListSubmissions(ctx context.Context, quizID uuid.UUID, q ListSubmissionsQuery) ([]QuizSubmission, int64, error)
	GradeSubmission(ctx context.Context, id uuid.UUID, dto GradeSubmissionDTO) (*QuizSubmission, error)

	AdminList(ctx context.Context, q AdminListQuizzesQuery) ([]Quiz, int64, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}
