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
	Source AutoMarkSource `json:"source" binding:"required,oneof=live_room offline_room"`
	// RoomID is required for the offline source. For the live source it is
	// optional: when set, presence is computed from that room alone; when
	// empty, the session's rooms are aggregated.
	RoomID uuid.UUID `json:"room_id" binding:"omitempty"`
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

// ListMyAttendanceQuery filters the caller's own attendance history on
// GET /attendance/me.
type ListMyAttendanceQuery struct {
	Status         *AttendanceStatus `form:"status" binding:"omitempty,oneof=present absent late excused"`
	ClassID        *uuid.UUID        `form:"-"`
	ClassSessionID *uuid.UUID        `form:"-"`
	ListParams     ListParams        `form:"-"`
}

// ListAttendanceMatrixQuery pages/searches/orders the STUDENT (row) axis of
// the attendance matrix. Sessions (columns) are always returned in full,
// ordered by start_time asc.
type ListAttendanceMatrixQuery struct {
	ListParams ListParams `form:"-"`
}

// AttendanceMatrixSession is one column header in the matrix.
type AttendanceMatrixSession struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
}

// AttendanceMatrixCell is one student×session intersection. Absent from the
// student's Cells map means "no record yet" (renders as a dash).
type AttendanceMatrixCell struct {
	ID           uuid.UUID        `json:"id"`
	Status       AttendanceStatus `json:"status"`
	IsAutoMarked bool             `json:"is_auto_marked"`
}

// AttendanceMatrixSummary is the trailing per-student summary column.
// Rate = (present+late) / started_count; started_count counts only sessions
// whose start_time <= now so future sessions don't drag the rate down.
type AttendanceMatrixSummary struct {
	Present      int     `json:"present"`
	Absent       int     `json:"absent"`
	Late         int     `json:"late"`
	Excused      int     `json:"excused"`
	StartedCount int     `json:"started_count"`
	Rate         float64 `json:"rate"`
}

// AttendanceMatrixStudent is one row: the user, their cells keyed by session
// id, and their summary.
type AttendanceMatrixStudent struct {
	UserID  uuid.UUID                          `json:"user_id"`
	User    *User                              `json:"user,omitempty"`
	Cells   map[uuid.UUID]AttendanceMatrixCell `json:"cells"`
	Summary AttendanceMatrixSummary            `json:"summary"`
}

// AttendanceMatrixResult is the full matrix payload. Total/Page/PageSize page
// the student axis only.
type AttendanceMatrixResult struct {
	Sessions []AttendanceMatrixSession `json:"sessions"`
	Students []AttendanceMatrixStudent `json:"students"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
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

// MyAttendanceSummary counts the caller's records by status over the FULL
// filtered set, not just the returned page.
type MyAttendanceSummary struct {
	Present int `json:"present"`
	Absent  int `json:"absent"`
	Late    int `json:"late"`
	Excused int `json:"excused"`
}

// MyAttendance is the caller's attendance history + summary across classes.
// Total/Page/PageSize describe the paged Items slice.
type MyAttendance struct {
	Summary  MyAttendanceSummary `json:"summary"`
	Items    []Attendance        `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

type AttendanceRepository interface {
	Create(ctx context.Context, a *Attendance) error
	FindByID(ctx context.Context, id uuid.UUID) (*Attendance, error)
	Update(ctx context.Context, a *Attendance) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListBySession(ctx context.Context, sessionID uuid.UUID, q ListAttendanceQuery) ([]Attendance, int64, error)
	FindBySessionAndUser(ctx context.Context, sessionID, userID uuid.UUID) (*Attendance, error)
	// ListByUser returns the user's attendance records across classes,
	// filtered by q, with Class + ClassSession preloaded, newest first.
	ListByUser(ctx context.Context, userID uuid.UUID, q ListMyAttendanceQuery) ([]Attendance, int64, error)
	// SummarizeByUser counts the user's records by status over the same
	// class/session scope as ListByUser, ignoring pagination AND the status
	// filter — the summary is a breakdown by status.
	SummarizeByUser(ctx context.Context, userID uuid.UUID, q ListMyAttendanceQuery) (MyAttendanceSummary, error)
	// ListByClassAndUsers returns every attendance row for the class that
	// belongs to one of userIDs. No pagination — caller has already paged the
	// user set. Returns an empty slice when userIDs is empty.
	ListByClassAndUsers(ctx context.Context, classID uuid.UUID, userIDs []uuid.UUID) ([]Attendance, error)

	// Admin-only.
	AdminList(ctx context.Context, q AdminListAttendanceQuery) ([]Attendance, int64, error)
}

type AttendanceService interface {
	Mark(ctx context.Context, classID, sessionID uuid.UUID, dto CreateAttendanceDTO) (*Attendance, error)
	BulkMark(ctx context.Context, classID, sessionID uuid.UUID, dto BulkCreateAttendanceDTO) ([]Attendance, error)
	AutoMark(ctx context.Context, classID, sessionID uuid.UUID, dto AutoMarkAttendanceDTO) (*AutoMarkResult, error)
	// AutoMarkSessionLive runs live auto-mark for a whole session using the org's
	// configured threshold. No caller authz (used by the worker / system).
	AutoMarkSessionLive(ctx context.Context, classID, sessionID uuid.UUID) (*AutoMarkResult, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateAttendanceDTO) (*Attendance, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Attendance, error)
	ListBySession(ctx context.Context, classID, sessionID uuid.UUID, q ListAttendanceQuery) ([]Attendance, int64, error)
	Matrix(ctx context.Context, classID uuid.UUID, q ListAttendanceMatrixQuery) (*AttendanceMatrixResult, error)
	ListMine(ctx context.Context, q ListMyAttendanceQuery) (*MyAttendance, error)

	// Admin surface. Require caller.IsAdmin.
	AdminList(ctx context.Context, q AdminListAttendanceQuery) ([]Attendance, int64, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateAttendanceDTO) (*Attendance, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}
