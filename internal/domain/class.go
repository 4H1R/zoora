package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ClassSessionType enumerates the allowed kinds of a ClassSession. Values are
// mirrored by a CHECK constraint in the class_sessions migration so bad data
// cannot reach the table even if it bypasses the service layer.
type ClassSessionType string

const (
	ClassSessionTypeLive     ClassSessionType = "live"
	ClassSessionTypeQuiz     ClassSessionType = "quiz"
	ClassSessionTypePractice ClassSessionType = "practice"
)

func (t ClassSessionType) Valid() bool {
	switch t {
	case ClassSessionTypeLive, ClassSessionTypeQuiz, ClassSessionTypePractice:
		return true
	}
	return false
}

// Class is a cohort owned by a single teacher (UserID) inside an organization.
// TotalUsers stores the maximum enrollment capacity (not the current member
// count). 0 means unlimited. Current enrollment is derived from class_members.
type Class struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User        *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	TotalUsers  int            `gorm:"not null;default:0" json:"total_users"` // capacity; 0 = unlimited
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type ClassSession struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ClassID      uuid.UUID        `gorm:"type:uuid;not null;index" json:"class_id"`
	Class        *Class           `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	Name         string           `gorm:"not null" json:"name"`
	Description  string           `json:"description"`
	StartTime    time.Time        `gorm:"not null" json:"start_time"`
	Type         ClassSessionType `gorm:"type:varchar(20);not null" json:"type"`
	IsRecordable bool             `gorm:"not null;default:false" json:"is_recordable"`
	QuizRooms    []QuizRoom       `gorm:"foreignKey:ClassSessionID" json:"quiz_rooms,omitempty"`
	LiveRooms      []LiveRoom       `gorm:"foreignKey:ClassSessionID" json:"live_rooms,omitempty"`
	PracticeRooms  []PracticeRoom   `gorm:"foreignKey:ClassSessionID" json:"practice_rooms,omitempty"`
	OfflineRooms   []OfflineRoom    `gorm:"foreignKey:ClassSessionID" json:"offline_rooms,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"-"`
}

// ClassMember links a user to a class they are enrolled in. The unique
// constraint on (class_id, user_id) in the migration prevents duplicate
// enrollments. No soft-delete: unenrolling is a hard delete.
type ClassMember struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ClassID   uuid.UUID `gorm:"type:uuid;not null;index" json:"class_id"`
	Class     *Class    `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateClassDTO struct {
	Name        string     `json:"name" binding:"required,min=2"`
	Description string     `json:"description"`
	TotalUsers  int        `json:"total_users" binding:"gte=0"` // capacity; 0 = unlimited
	UserID      *uuid.UUID `json:"user_id" binding:"omitempty,uuid4"`
}

type UpdateClassDTO struct {
	Name        *string    `json:"name" binding:"omitempty,min=2"`
	Description *string    `json:"description"`
	TotalUsers  *int       `json:"total_users" binding:"omitempty,gte=0"`
	UserID      *uuid.UUID `json:"user_id" binding:"omitempty,uuid4"`
}

type CreateClassSessionDTO struct {
	Name         string           `json:"name" binding:"required,min=2"`
	Description  string           `json:"description"`
	StartTime    time.Time        `json:"start_time" binding:"required"`
	Type         ClassSessionType `json:"type" binding:"required,oneof=live quiz practice"`
	IsRecordable bool             `json:"is_recordable"`
}

type UpdateClassSessionDTO struct {
	Name         *string           `json:"name" binding:"omitempty,min=2"`
	Description  *string           `json:"description"`
	StartTime    *time.Time        `json:"start_time"`
	Type         *ClassSessionType `json:"type" binding:"omitempty,oneof=live quiz practice"`
	IsRecordable *bool             `json:"is_recordable"`
}

type EnrollClassMemberDTO struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type ClassListScope struct {
	All            bool
	OrganizationID *uuid.UUID
	TeacherID      *uuid.UUID
	MemberUserID   *uuid.UUID
	IncludeDeleted bool
}

// AdminListClassesQuery is the query for GET /admin/classes. Typed filters
// sit alongside the embedded ListParams populated by the handler after
// white-listing.
type AdminListClassesQuery struct {
	UserID         *uuid.UUID `form:"user_id"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

// AdminCreateClassDTO is the body for POST /admin/classes.
type AdminCreateClassDTO struct {
	OrganizationID uuid.UUID `json:"organization_id" binding:"required"`
	UserID         uuid.UUID `json:"user_id" binding:"required"`
	Name           string    `json:"name" binding:"required,min=2"`
	Description    string    `json:"description"`
	TotalUsers     int       `json:"total_users" binding:"gte=0"`
}

// AdminUpdateClassDTO is the body for PUT /admin/classes/:id.
type AdminUpdateClassDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=2"`
	Description *string `json:"description"`
	TotalUsers  *int    `json:"total_users" binding:"omitempty,gte=0"`
}

// ListClassSessionsQuery is the query for GET /classes/:id/sessions.
type ListClassSessionsQuery struct {
	Type           *ClassSessionType `form:"type"`
	IncludeDeleted bool              `form:"include_deleted"`
	ListParams     ListParams        `form:"-"`
}

// ListClassMembersQuery is the query for GET /classes/:id/members.
type ListClassMembersQuery struct {
	ListParams ListParams `form:"-"`
}

type ClassRepository interface {
	Create(ctx context.Context, class *Class) error
	FindByID(ctx context.Context, id uuid.UUID) (*Class, error)
	Update(ctx context.Context, class *Class) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List applies a role-resolved scope produced by the service from the
	// Caller. The repository itself is role-agnostic — it only knows how to
	// translate the scope into SQL filters.
	List(ctx context.Context, scope ClassListScope, p ListParams) ([]Class, int64, error)

	// Admin-only.
	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*Class, error)
	AdminList(ctx context.Context, q AdminListClassesQuery) ([]Class, int64, error)
}

type ClassSessionRepository interface {
	Create(ctx context.Context, session *ClassSession) error
	FindByID(ctx context.Context, id uuid.UUID) (*ClassSession, error)
	Update(ctx context.Context, session *ClassSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByClass(ctx context.Context, classID uuid.UUID, q ListClassSessionsQuery) ([]ClassSession, int64, error)

	// Admin-only.
	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*ClassSession, error)
}

type ClassMemberRepository interface {
	Create(ctx context.Context, m *ClassMember) error
	Delete(ctx context.Context, classID, userID uuid.UUID) error
	Exists(ctx context.Context, classID, userID uuid.UUID) (bool, error)
	CountByClass(ctx context.Context, classID uuid.UUID) (int64, error)
	ListByClass(ctx context.Context, classID uuid.UUID, p ListParams) ([]ClassMember, int64, error)
	ListAllByClass(ctx context.Context, classID uuid.UUID) ([]ClassMember, error)
}

type ClassService interface {
	Create(ctx context.Context, dto CreateClassDTO) (*Class, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Class, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateClassDTO) (*Class, error)
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns classes visible to the caller under the RBAC hierarchy:
	// super-admin sees all, org-admin sees their org, teacher sees their
	// own classes, student sees classes they are enrolled in.
	List(ctx context.Context, p ListParams) ([]Class, int64, error)

	CreateSession(ctx context.Context, classID uuid.UUID, dto CreateClassSessionDTO) (*ClassSession, error)
	GetSession(ctx context.Context, id uuid.UUID) (*ClassSession, error)
	UpdateSession(ctx context.Context, id uuid.UUID, dto UpdateClassSessionDTO) (*ClassSession, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	ListSessions(ctx context.Context, classID uuid.UUID, q ListClassSessionsQuery) ([]ClassSession, int64, error)

	Enroll(ctx context.Context, classID uuid.UUID, dto EnrollClassMemberDTO) (*ClassMember, error)
	Leave(ctx context.Context, classID, userID uuid.UUID) error
	ListMembers(ctx context.Context, classID uuid.UUID, q ListClassMembersQuery) ([]ClassMember, int64, error)

	// Admin surface. Require caller.IsAdmin.
	AdminList(ctx context.Context, q AdminListClassesQuery) ([]Class, int64, error)
	AdminCreate(ctx context.Context, dto AdminCreateClassDTO) (*Class, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateClassDTO) (*Class, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
	AdminHardDeleteSession(ctx context.Context, id uuid.UUID) error
}
