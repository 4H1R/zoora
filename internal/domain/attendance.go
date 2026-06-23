package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "present"
	AttendanceStatusAbsent  AttendanceStatus = "absent"
	AttendanceStatusLate    AttendanceStatus = "late"
	AttendanceStatusExcused AttendanceStatus = "excused"
)

func (s AttendanceStatus) Valid() bool {
	switch s {
	case AttendanceStatusPresent, AttendanceStatusAbsent, AttendanceStatusLate, AttendanceStatusExcused:
		return true
	}
	return false
}

type Attendance struct {
	ID             uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID        `gorm:"type:uuid;not null;index" json:"organization_id"`
	ClassID        uuid.UUID        `gorm:"type:uuid;not null;index" json:"class_id"`
	Class          *Class           `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	ClassSessionID uuid.UUID        `gorm:"type:uuid;not null;index" json:"class_session_id"`
	ClassSession   *ClassSession    `gorm:"foreignKey:ClassSessionID" json:"class_session,omitempty"`
	UserID         uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	User           *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Status         AttendanceStatus `gorm:"type:varchar(20);not null" json:"status"`
	IsAutoMarked   bool             `gorm:"not null;default:false" json:"is_auto_marked"`
	Remarks        string           `json:"remarks"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type CreateAttendanceDTO struct {
	UserID       uuid.UUID        `json:"user_id" binding:"required"`
	Status       AttendanceStatus `json:"status" binding:"required,oneof=present absent late excused"`
	IsAutoMarked bool             `json:"is_auto_marked"`
	Remarks      string           `json:"remarks"`
}

type UpdateAttendanceDTO struct {
	Status  *AttendanceStatus `json:"status" binding:"omitempty,oneof=present absent late excused"`
	Remarks *string           `json:"remarks"`
}

type BulkCreateAttendanceDTO struct {
	Entries []CreateAttendanceDTO `json:"entries" binding:"required,min=1,dive"`
}

type AutoMarkSource string

const (
	AutoMarkSourceLive    AutoMarkSource = "live_room"
	AutoMarkSourceOffline AutoMarkSource = "offline_room"
)

type AutoMarkAttendanceDTO struct {
	Source            AutoMarkSource `json:"source" binding:"required,oneof=live_room offline_room"`
	RoomID            uuid.UUID      `json:"room_id" binding:"required"`
	MinDurationSeconds int           `json:"min_duration_seconds" binding:"gte=0"`
}

type AutoMarkResult struct {
	Marked  int `json:"marked"`
	Skipped int `json:"skipped"`
}

type ListAttendanceQuery struct {
	Status       *AttendanceStatus `form:"status" binding:"omitempty,oneof=present absent late excused"`
	IsAutoMarked *bool             `form:"is_auto_marked"`
	UserID       *uuid.UUID        `form:"-"`
	ListParams   ListParams        `form:"-"`
}

type AdminListAttendanceQuery struct {
	Status         *AttendanceStatus `form:"status" binding:"omitempty,oneof=present absent late excused"`
	IsAutoMarked   *bool             `form:"is_auto_marked"`
	UserID         *uuid.UUID        `form:"-"`
	ClassID        *uuid.UUID        `form:"-"`
	ClassSessionID *uuid.UUID        `form:"-"`
	OrganizationID *uuid.UUID        `form:"-"`
	ListParams     ListParams        `form:"-"`
}

type AdminUpdateAttendanceDTO struct {
	Status  *AttendanceStatus `json:"status" binding:"omitempty,oneof=present absent late excused"`
	Remarks *string           `json:"remarks"`
}

// MyAttendanceSummary counts the caller's records by status.
type MyAttendanceSummary struct {
	Present int `json:"present"`
	Absent  int `json:"absent"`
	Late    int `json:"late"`
	Excused int `json:"excused"`
}

// MyAttendance is the caller's attendance history + summary across classes.
type MyAttendance struct {
	Summary MyAttendanceSummary `json:"summary"`
	Items   []Attendance        `json:"items"`
}

type AttendanceRepository interface {
	Create(ctx context.Context, a *Attendance) error
	FindByID(ctx context.Context, id uuid.UUID) (*Attendance, error)
	Update(ctx context.Context, a *Attendance) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListBySession(ctx context.Context, sessionID uuid.UUID, q ListAttendanceQuery) ([]Attendance, int64, error)
	FindBySessionAndUser(ctx context.Context, sessionID, userID uuid.UUID) (*Attendance, error)
	// ListByUser returns all attendance records for a user across classes,
	// with Class + ClassSession preloaded, newest first.
	ListByUser(ctx context.Context, userID uuid.UUID, p ListParams) ([]Attendance, int64, error)

	// Admin-only.
	AdminList(ctx context.Context, q AdminListAttendanceQuery) ([]Attendance, int64, error)
}

type AttendanceService interface {
	Mark(ctx context.Context, classID, sessionID uuid.UUID, dto CreateAttendanceDTO) (*Attendance, error)
	BulkMark(ctx context.Context, classID, sessionID uuid.UUID, dto BulkCreateAttendanceDTO) ([]Attendance, error)
	AutoMark(ctx context.Context, classID, sessionID uuid.UUID, dto AutoMarkAttendanceDTO) (*AutoMarkResult, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateAttendanceDTO) (*Attendance, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Attendance, error)
	ListBySession(ctx context.Context, classID, sessionID uuid.UUID, q ListAttendanceQuery) ([]Attendance, int64, error)
	ListMine(ctx context.Context, p ListParams) (*MyAttendance, error)

	// Admin surface. Require caller.IsAdmin.
	AdminList(ctx context.Context, q AdminListAttendanceQuery) ([]Attendance, int64, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateAttendanceDTO) (*Attendance, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}
