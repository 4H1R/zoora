package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PracticeRoom struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	ClassID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"class_id"`
	Class          *Class         `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	ClassSessionID uuid.UUID      `gorm:"type:uuid;not null;index" json:"class_session_id"`
	ClassSession   *ClassSession  `gorm:"foreignKey:ClassSessionID" json:"class_session,omitempty"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User           *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title          string         `gorm:"not null" json:"title"`
	Content        string         `gorm:"type:text;not null;default:''" json:"content"`
	MaxScore       float64        `gorm:"not null;default:0" json:"max_score"`
	StartTime      time.Time      `gorm:"not null" json:"start_time"`
	EndTime        time.Time      `gorm:"not null" json:"end_time"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type PracticeSubmission struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	PracticeRoomID uuid.UUID      `gorm:"type:uuid;not null;index" json:"practice_room_id"`
	PracticeRoom   *PracticeRoom  `gorm:"foreignKey:PracticeRoomID" json:"practice_room,omitempty"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User           *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Content        string         `gorm:"type:text;not null;default:''" json:"content"`
	Score          *float64       `json:"score"`
	TeacherComment string         `gorm:"type:text;not null;default:''" json:"teacher_comment"`
	SubmittedAt    time.Time      `gorm:"not null" json:"submitted_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// --- DTOs ---

type CreatePracticeRoomDTO struct {
	ClassSessionID uuid.UUID `json:"class_session_id" binding:"required"`
	Title          string    `json:"title" binding:"required,min=2"`
	Content        string    `json:"content"`
	MaxScore       float64   `json:"max_score" binding:"gte=0"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
}

type UpdatePracticeRoomDTO struct {
	Title     *string    `json:"title" binding:"omitempty,min=2"`
	Content   *string    `json:"content"`
	MaxScore  *float64   `json:"max_score" binding:"omitempty,gte=0"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

type CreatePracticeSubmissionDTO struct {
	Content string `json:"content"`
}

type GradePracticeSubmissionDTO struct {
	Score          *float64 `json:"score" binding:"omitempty,gte=0"`
	TeacherComment *string  `json:"teacher_comment"`
}

type PracticeRoomListScope struct {
	All            bool
	OrganizationID *uuid.UUID
	OwnerID        *uuid.UUID
	MemberUserID   *uuid.UUID
	IncludeDeleted bool
}

type ListPracticeRoomsQuery struct {
	ClassID        *uuid.UUID `form:"class_id"`
	ClassSessionID *uuid.UUID `form:"class_session_id"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type ListPracticeSubmissionsQuery struct {
	ListParams ListParams `form:"-"`
}

type AdminListPracticeRoomsQuery struct {
	ClassID        *uuid.UUID `form:"class_id"`
	ClassSessionID *uuid.UUID `form:"class_session_id"`
	UserID         *uuid.UUID `form:"user_id"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

// --- Interfaces ---

type PracticeRoomRepository interface {
	Create(ctx context.Context, room *PracticeRoom) error
	FindByID(ctx context.Context, id uuid.UUID) (*PracticeRoom, error)
	Update(ctx context.Context, room *PracticeRoom) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope PracticeRoomListScope, q ListPracticeRoomsQuery) ([]PracticeRoom, int64, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*PracticeRoom, error)
	AdminList(ctx context.Context, q AdminListPracticeRoomsQuery) ([]PracticeRoom, int64, error)
}

type PracticeSubmissionRepository interface {
	Create(ctx context.Context, sub *PracticeSubmission) error
	FindByID(ctx context.Context, id uuid.UUID) (*PracticeSubmission, error)
	Update(ctx context.Context, sub *PracticeSubmission) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*PracticeSubmission, error)
	ListByRoom(ctx context.Context, roomID uuid.UUID, p ListParams) ([]PracticeSubmission, int64, error)
}

type PracticeService interface {
	CreateRoom(ctx context.Context, dto CreatePracticeRoomDTO) (*PracticeRoom, error)
	GetRoom(ctx context.Context, id uuid.UUID) (*PracticeRoom, error)
	UpdateRoom(ctx context.Context, id uuid.UUID, dto UpdatePracticeRoomDTO) (*PracticeRoom, error)
	DeleteRoom(ctx context.Context, id uuid.UUID) error
	ListRooms(ctx context.Context, q ListPracticeRoomsQuery) ([]PracticeRoom, int64, error)

	Submit(ctx context.Context, roomID uuid.UUID, dto CreatePracticeSubmissionDTO) (*PracticeSubmission, error)
	GetSubmission(ctx context.Context, id uuid.UUID) (*PracticeSubmission, error)
	ListSubmissions(ctx context.Context, roomID uuid.UUID, q ListPracticeSubmissionsQuery) ([]PracticeSubmission, int64, error)
	Grade(ctx context.Context, submissionID uuid.UUID, dto GradePracticeSubmissionDTO) (*PracticeSubmission, error)

	AdminList(ctx context.Context, q AdminListPracticeRoomsQuery) ([]PracticeRoom, int64, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}
