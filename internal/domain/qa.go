package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QA model types (polymorphic attach target). Only live_session is supported today.
const QAModelLiveSession = "live_session"

// QA question lifecycle states.
const (
	QAStatusOpen      = "open"
	QAStatusResolved  = "resolved"
	QAStatusDismissed = "dismissed"
)

// MaxOpenQuestionsPerUser caps how many open questions one user may have on a
// single model at once, to stop a single participant flooding the board.
const MaxOpenQuestionsPerUser = 3

// QAQuestion is an audience question posted against a polymorphic model
// (a live room today). Authorship is always attributed via UserID.
type QAQuestion struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ModelType string         `gorm:"type:varchar(100);not null" json:"model_type"`
	ModelID   uuid.UUID      `gorm:"type:uuid;not null" json:"model_id"`
	Text      string         `gorm:"not null" json:"text"`
	Status    string         `gorm:"type:varchar(20);not null;default:open" json:"status"`
	ClosedAt  *time.Time     `json:"closed_at,omitempty"`
	ClosedBy  *uuid.UUID     `gorm:"type:uuid" json:"closed_by,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// QAVote is one participant's upvote of a question. The UNIQUE(question_id,
// user_id) index makes double-voting impossible at the database level.
type QAVote struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:uq_qa_votes_question_user" json:"question_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:uq_qa_votes_question_user" json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// QAQuestionView is a read projection carrying computed vote data for a given
// viewer. VoteCount is COUNT(*) of qa_votes; VotedByMe is whether the current
// caller has voted. Populated by the repository list query.
type QAQuestionView struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	AuthorName string     `json:"author_name"`
	ModelType  string     `json:"model_type"`
	ModelID    uuid.UUID  `json:"model_id"`
	Text       string     `json:"text"`
	Status     string     `json:"status"`
	ClosedAt   *time.Time `json:"closed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	VoteCount  int        `json:"vote_count"`
	VotedByMe  bool       `json:"voted_by_me"`
}

type CreateQAQuestionDTO struct {
	ModelType string    `json:"model_type" binding:"required,max=100"`
	ModelID   uuid.UUID `json:"model_id" binding:"required"`
	Text      string    `json:"text" binding:"required,min=2,max=500"`
}

type UpdateQAQuestionDTO struct {
	Text string `json:"text" binding:"required,min=2,max=500"`
}

type ListQAQuestionsQuery struct {
	ModelType  *string    `form:"model_type"`
	ModelID    *uuid.UUID `form:"-"`
	Status     *string    `form:"status"`
	ListParams ListParams `form:"-"`
}

type AdminListQAQuestionsQuery struct {
	UserID         *uuid.UUID `form:"-"`
	ModelType      *string    `form:"model_type"`
	ModelID        *uuid.UUID `form:"-"`
	Status         *string    `form:"status"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

// QAListScope holds resolved list filters for the viewer-facing list query.
type QAListScope struct {
	ViewerID  uuid.UUID
	ModelType *string
	ModelID   *uuid.UUID
	Status    *string
}

// QAVoteResult is returned by a vote toggle: the caller's new vote state and the
// question's updated vote count.
type QAVoteResult struct {
	Voted     bool  `json:"voted"`
	VoteCount int64 `json:"vote_count"`
}

type QARepository interface {
	Create(ctx context.Context, q *QAQuestion) error
	FindByID(ctx context.Context, id uuid.UUID) (*QAQuestion, error)
	Update(ctx context.Context, q *QAQuestion) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope QAListScope, p ListParams) ([]QAQuestionView, int64, error)
	CountOpenByUser(ctx context.Context, modelType string, modelID, userID uuid.UUID) (int64, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*QAQuestion, error)
	AdminList(ctx context.Context, q AdminListQAQuestionsQuery) ([]QAQuestion, int64, error)
}

type QAVoteRepository interface {
	Create(ctx context.Context, v *QAVote) error
	Delete(ctx context.Context, questionID, userID uuid.UUID) (bool, error) // returns whether a row was deleted
	Exists(ctx context.Context, questionID, userID uuid.UUID) (bool, error)
	CountByQuestion(ctx context.Context, questionID uuid.UUID) (int64, error)
}

type QAService interface {
	Ask(ctx context.Context, dto CreateQAQuestionDTO) (*QAQuestion, error)
	List(ctx context.Context, q ListQAQuestionsQuery) ([]QAQuestionView, int64, error)
	UpdateText(ctx context.Context, id uuid.UUID, dto UpdateQAQuestionDTO) (*QAQuestion, error)
	Delete(ctx context.Context, id uuid.UUID) error

	ToggleVote(ctx context.Context, id uuid.UUID) (voted bool, count int64, err error)
	Resolve(ctx context.Context, id uuid.UUID) (*QAQuestion, error)
	Dismiss(ctx context.Context, id uuid.UUID) (*QAQuestion, error)
	Reopen(ctx context.Context, id uuid.UUID) (*QAQuestion, error)

	AdminList(ctx context.Context, q AdminListQAQuestionsQuery) ([]QAQuestion, int64, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}

// TableName pins the questions table name.
func (QAQuestion) TableName() string { return "qa_questions" }

// TableName pins the votes table name.
func (QAVote) TableName() string { return "qa_votes" }
