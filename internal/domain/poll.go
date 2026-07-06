package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PollOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Poll struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	UserID              uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User                *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ModelType           string         `gorm:"type:varchar(100);not null" json:"model_type"`
	ModelID             uuid.UUID      `gorm:"type:uuid;not null" json:"model_id"`
	Name                string         `gorm:"not null" json:"name"`
	AllowedAnswersCount int            `gorm:"not null;default:1" json:"allowed_answers_count"`
	Options             []PollOption   `gorm:"type:jsonb;serializer:json" json:"options"`
	ClosedAt            *time.Time     `json:"closed_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// IsClosed reports whether the poll has been closed and no longer accepts answers.
func (p *Poll) IsClosed() bool {
	return p.ClosedAt != nil
}

type PollAnswer struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PollID    uuid.UUID `gorm:"type:uuid;not null;index" json:"poll_id"`
	Poll      *Poll     `gorm:"foreignKey:PollID" json:"poll,omitempty"`
	Option    string    `gorm:"not null" json:"option"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePollDTO struct {
	ModelType           string       `json:"model_type" binding:"required,max=100"`
	ModelID             uuid.UUID    `json:"model_id" binding:"required"`
	Name                string       `json:"name" binding:"required,min=2"`
	AllowedAnswersCount int          `json:"allowed_answers_count" binding:"required,gte=1"`
	Options             []PollOption `json:"options" binding:"required,min=2,dive"`
}

type UpdatePollDTO struct {
	Name                *string      `json:"name" binding:"omitempty,min=2"`
	AllowedAnswersCount *int         `json:"allowed_answers_count" binding:"omitempty,gte=1"`
	Options             []PollOption `json:"options" binding:"omitempty,min=2,dive"`
}

type AnswerPollDTO struct {
	Options []string `json:"options" binding:"required,min=1"`
}

type ListPollsQuery struct {
	ModelType      *string    `form:"model_type"`
	ModelID        *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type AdminListPollsQuery struct {
	UserID         *uuid.UUID `form:"-"`
	ModelType      *string    `form:"model_type"`
	ModelID        *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type ListPollAnswersQuery struct {
	UserID     *uuid.UUID `form:"-"`
	ListParams ListParams `form:"-"`
}

type PollResults struct {
	PollID uuid.UUID      `json:"poll_id"`
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
}

type PollListScope struct {
	AllOrgs        bool
	OwnerID        *uuid.UUID
	ModelType      *string
	ModelID        *uuid.UUID
	IncludeDeleted bool
}

type PollRepository interface {
	Create(ctx context.Context, poll *Poll) error
	FindByID(ctx context.Context, id uuid.UUID) (*Poll, error)
	Update(ctx context.Context, poll *Poll) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope PollListScope, p ListParams) ([]Poll, int64, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*Poll, error)
	AdminList(ctx context.Context, q AdminListPollsQuery) ([]Poll, int64, error)

	// CloseByModel marks every open poll owned by the given polymorphic model as
	// closed (closed_at = now). Idempotent: already-closed polls are untouched.
	CloseByModel(ctx context.Context, modelType string, modelID uuid.UUID) error
}

type PollAnswerRepository interface {
	Create(ctx context.Context, answer *PollAnswer) error
	FindByPollAndUser(ctx context.Context, pollID, userID uuid.UUID) ([]PollAnswer, error)
	ListByPoll(ctx context.Context, pollID uuid.UUID, q ListPollAnswersQuery) ([]PollAnswer, int64, error)
	DeleteByPollAndUser(ctx context.Context, pollID, userID uuid.UUID) error
	CountByOption(ctx context.Context, pollID uuid.UUID) (map[string]int, int, error)
}

type PollService interface {
	Create(ctx context.Context, dto CreatePollDTO) (*Poll, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Poll, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdatePollDTO) (*Poll, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, q ListPollsQuery) ([]Poll, int64, error)

	Answer(ctx context.Context, pollID uuid.UUID, dto AnswerPollDTO) ([]PollAnswer, error)
	ListAnswers(ctx context.Context, pollID uuid.UUID, q ListPollAnswersQuery) ([]PollAnswer, int64, error)
	Results(ctx context.Context, pollID uuid.UUID) (*PollResults, error)

	AdminList(ctx context.Context, q AdminListPollsQuery) ([]Poll, int64, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error

	// CloseByModel closes all open polls for a polymorphic owner (e.g. a live
	// room on finish). A system operation — no caller authz required.
	CloseByModel(ctx context.Context, modelType string, modelID uuid.UUID) error
}
