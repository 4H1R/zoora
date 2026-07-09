package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TicketType string

const (
	TicketTypeQuestion       TicketType = "question"
	TicketTypeGradeObjection TicketType = "grade_objection"
	TicketTypeOther          TicketType = "other"
)

func (t TicketType) Valid() bool {
	switch t {
	case TicketTypeQuestion, TicketTypeGradeObjection, TicketTypeOther:
		return true
	}
	return false
}

type TicketStatus string

const (
	TicketStatusOpen     TicketStatus = "open"
	TicketStatusAnswered TicketStatus = "answered"
	TicketStatusClosed   TicketStatus = "closed"
)

func (s TicketStatus) Valid() bool {
	switch s {
	case TicketStatusOpen, TicketStatusAnswered, TicketStatusClosed:
		return true
	}
	return false
}

// Ticket is a class-scoped support/objection thread opened by an enrolled
// student and handled by the class teacher (Class.UserID). Membership is
// checked only at creation: tickets survive unenroll, and access afterwards is
// ownership-based (creator or class teacher).
type Ticket struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	ClassID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"class_id"`
	Class          *Class     `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"` // creator
	User           *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title          string     `gorm:"type:varchar(255);not null" json:"title"`
	Type           TicketType `gorm:"type:varchar(20);not null" json:"type"`
	// grade_objection targets: at most one set (DB CHECK); both nil = general.
	QuizRoomID        *uuid.UUID       `gorm:"type:uuid" json:"quiz_room_id,omitempty"`
	QuizRoom          *QuizRoom        `gorm:"foreignKey:QuizRoomID" json:"quiz_room,omitempty"`
	GradebookColumnID *uuid.UUID       `gorm:"type:uuid" json:"gradebook_column_id,omitempty"`
	GradebookColumn   *GradebookColumn `gorm:"foreignKey:GradebookColumnID" json:"gradebook_column,omitempty"`
	Status            TicketStatus     `gorm:"type:varchar(20);not null;default:'open'" json:"status"`
	ClosedAt          *time.Time       `json:"closed_at,omitempty"`
	ClosedBy          *uuid.UUID       `gorm:"type:uuid" json:"closed_by,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`

	// Populated by TicketService.Get only (not a column preload on lists).
	Messages []TicketMessage `gorm:"foreignKey:TicketID" json:"messages,omitempty"`
}

// TicketMessage is one immutable thread entry. No edit/delete: objection
// threads are quasi-audit records.
type TicketMessage struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	TicketID  uuid.UUID       `gorm:"type:uuid;not null;index" json:"ticket_id"`
	UserID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Body      string          `gorm:"type:text;not null" json:"body"`
	MediaIDs  json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"media_ids"`
	CreatedAt time.Time       `json:"created_at"`
}

// ---- DTOs ----

type CreateTicketDTO struct {
	ClassID           string     `json:"class_id" binding:"required,uuid"`
	Title             string     `json:"title" binding:"required,min=2,max=255"`
	Type              TicketType `json:"type" binding:"required,oneof=question grade_objection other"`
	QuizRoomID        *string    `json:"quiz_room_id" binding:"omitempty,uuid"`
	GradebookColumnID *string    `json:"gradebook_column_id" binding:"omitempty,uuid"`
	Body              string     `json:"body" binding:"required,min=1,max=10000"`
	MediaIDs          []string   `json:"media_ids" binding:"omitempty,max=10,dive,uuid"`
}

type AddTicketMessageDTO struct {
	Body     string   `json:"body" binding:"required,min=1,max=10000"`
	MediaIDs []string `json:"media_ids" binding:"omitempty,max=10,dive,uuid"`
}

// ---- Query / scope types ----

type ListTicketsQuery struct {
	ClassID    *uuid.UUID
	Status     *TicketStatus
	Type       *TicketType
	ListParams ListParams
}

// TicketListScope is role-resolved by the service; the repository only
// translates it to SQL. UserID nil = every ticket in the org (platform admin).
// UserID set = tickets the user created OR tickets of classes the user owns.
type TicketListScope struct {
	OrganizationID *uuid.UUID
	UserID         *uuid.UUID
}

// ---- Interfaces ----

type TicketRepository interface {
	Create(ctx context.Context, t *Ticket) error
	// FindByID preloads User, Class, QuizRoom.Quiz, GradebookColumn.
	FindByID(ctx context.Context, id uuid.UUID) (*Ticket, error)
	List(ctx context.Context, scope TicketListScope, q ListTicketsQuery) ([]Ticket, int64, error)
	// SetStatus updates status (+ closed_at/closed_by, nil to clear) and bumps
	// updated_at.
	SetStatus(ctx context.Context, id uuid.UUID, status TicketStatus, closedBy *uuid.UUID, closedAt *time.Time) error
}

type TicketMessageRepository interface {
	Create(ctx context.Context, m *TicketMessage) error
	// ListByTicket returns the full thread oldest-first with User preloaded.
	ListByTicket(ctx context.Context, ticketID uuid.UUID) ([]TicketMessage, error)
}

type TicketService interface {
	Create(ctx context.Context, dto CreateTicketDTO) (*Ticket, error)
	// Get returns the ticket with Messages populated.
	Get(ctx context.Context, id uuid.UUID) (*Ticket, error)
	List(ctx context.Context, q ListTicketsQuery) ([]Ticket, int64, error)
	AddMessage(ctx context.Context, ticketID uuid.UUID, dto AddTicketMessageDTO) (*TicketMessage, error)
	Close(ctx context.Context, ticketID uuid.UUID) (*Ticket, error)
}
