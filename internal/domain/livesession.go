package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LiveRoomStatus string

const (
	LiveRoomStatusCreated  LiveRoomStatus = "created"
	LiveRoomStatusActive   LiveRoomStatus = "active"
	LiveRoomStatusFinished LiveRoomStatus = "finished"
)

func (s LiveRoomStatus) Valid() bool {
	switch s {
	case LiveRoomStatusCreated, LiveRoomStatusActive, LiveRoomStatusFinished:
		return true
	}
	return false
}

type LiveRecordingStatus string

const (
	LiveRecordingStatusStarted   LiveRecordingStatus = "started"
	LiveRecordingStatusCompleted LiveRecordingStatus = "completed"
	LiveRecordingStatusFailed    LiveRecordingStatus = "failed"
)

type LiveRoomConfig struct {
	AllowMicDefault         bool `json:"allow_mic_default"`
	AllowCameraDefault      bool `json:"allow_camera_default"`
	AllowScreenShareDefault bool `json:"allow_screen_share_default"`
	AutoRecord              bool `json:"auto_record"`
	MaxParticipants         int  `json:"max_participants"`
}

func DefaultLiveRoomConfig() LiveRoomConfig {
	return LiveRoomConfig{
		AllowMicDefault:         true,
		AllowCameraDefault:      true,
		AllowScreenShareDefault: false,
		AutoRecord:              false,
		MaxParticipants:         100,
	}
}

type LiveRoom struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ClassSessionID  uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"class_session_id"`
	ClassSession    *ClassSession  `gorm:"foreignKey:ClassSessionID" json:"class_session,omitempty"`
	LiveKitRoomName string         `gorm:"column:livekit_room_name;type:varchar(255);not null;uniqueIndex" json:"livekit_room_name"`
	Status          LiveRoomStatus `gorm:"type:varchar(20);not null;default:'created'" json:"status"`
	Config          LiveRoomConfig `gorm:"type:jsonb;not null;serializer:json" json:"config"`
	ActualStartTime *time.Time     `json:"actual_start_time"`
	ActualEndTime   *time.Time     `json:"actual_end_time"`
	HostLastSeenAt  *time.Time     `json:"host_last_seen_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type LiveParticipant struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	LiveRoomID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"live_room_id"`
	LiveRoom             *LiveRoom  `gorm:"foreignKey:LiveRoomID" json:"live_room,omitempty"`
	UserID               uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User                 *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Identity             string     `gorm:"type:varchar(255);not null" json:"identity"`
	JoinedAt             time.Time  `gorm:"not null" json:"joined_at"`
	LeftAt               *time.Time `json:"left_at"`
	TotalDurationSeconds int        `gorm:"not null;default:0" json:"total_duration_seconds"`
	CreatedAt            time.Time  `json:"created_at"`
}

type LiveRecording struct {
	ID         uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	LiveRoomID uuid.UUID           `gorm:"type:uuid;not null;index" json:"live_room_id"`
	LiveRoom   *LiveRoom           `gorm:"foreignKey:LiveRoomID" json:"live_room,omitempty"`
	EgressID   string              `gorm:"type:varchar(255);not null" json:"egress_id"`
	Status     LiveRecordingStatus `gorm:"type:varchar(20);not null;default:'started'" json:"status"`
	FileURL    string              `gorm:"type:text" json:"file_url"`
	Duration   int                 `gorm:"not null;default:0" json:"duration"`
	Size       int64               `gorm:"not null;default:0" json:"size"`
	StartedAt  time.Time           `gorm:"not null" json:"started_at"`
	EndedAt    *time.Time          `json:"ended_at"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

type CreateLiveRoomDTO struct {
	ClassSessionID uuid.UUID      `json:"class_session_id" binding:"required"`
	Config         LiveRoomConfig `json:"config"`
}

type UpdateLiveRoomConfigDTO struct {
	Config *LiveRoomConfig `json:"config" binding:"required"`
}

type JoinLiveRoomResponse struct {
	Token      string     `json:"token"`
	LiveKitURL string     `json:"livekit_url"`
	Room       *LiveRoom  `json:"room"`
	ChatID     *uuid.UUID `json:"chat_id,omitempty"`
}

type LiveRoomListScope struct {
	All          bool
	TeacherID    *uuid.UUID
	MemberUserID *uuid.UUID
}

type AdminListLiveRoomsQuery struct {
	Status         *LiveRoomStatus `form:"status"`
	IncludeDeleted bool            `form:"include_deleted"`
	ListParams     ListParams      `form:"-"`
}

type LiveRoomRepository interface {
	Create(ctx context.Context, room *LiveRoom) error
	FindByID(ctx context.Context, id uuid.UUID) (*LiveRoom, error)
	FindByClassSessionID(ctx context.Context, sessionID uuid.UUID) (*LiveRoom, error)
	Update(ctx context.Context, room *LiveRoom) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope LiveRoomListScope, p ListParams) ([]LiveRoom, int64, error)
	FindActiveRoomsWithStaleHost(ctx context.Context, staleDuration time.Duration) ([]LiveRoom, error)
	AdminList(ctx context.Context, q AdminListLiveRoomsQuery) ([]LiveRoom, int64, error)
	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*LiveRoom, error)
}

type LiveParticipantRepository interface {
	Create(ctx context.Context, p *LiveParticipant) error
	FindActiveByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*LiveParticipant, error)
	Update(ctx context.Context, p *LiveParticipant) error
	ListByRoom(ctx context.Context, roomID uuid.UUID, p ListParams) ([]LiveParticipant, int64, error)
	ListAllByRoom(ctx context.Context, roomID uuid.UUID) ([]LiveParticipant, error)
	MarkAllLeft(ctx context.Context, roomID uuid.UUID, leftAt time.Time) error
}

type LiveRecordingRepository interface {
	Create(ctx context.Context, r *LiveRecording) error
	FindByID(ctx context.Context, id uuid.UUID) (*LiveRecording, error)
	FindActiveByRoom(ctx context.Context, roomID uuid.UUID) (*LiveRecording, error)
	Update(ctx context.Context, r *LiveRecording) error
	ListByRoom(ctx context.Context, roomID uuid.UUID) ([]LiveRecording, error)
}

type LiveSessionService interface {
	CreateRoom(ctx context.Context, dto CreateLiveRoomDTO) (*LiveRoom, error)
	GetRoom(ctx context.Context, id uuid.UUID) (*LiveRoom, error)
	JoinRoom(ctx context.Context, roomID uuid.UUID) (*JoinLiveRoomResponse, error)
	LeaveRoom(ctx context.Context, roomID uuid.UUID) error
	StartRoom(ctx context.Context, roomID uuid.UUID) (*LiveRoom, error)
	EndRoom(ctx context.Context, roomID uuid.UUID) (*LiveRoom, error)
	UpdateRoomConfig(ctx context.Context, roomID uuid.UUID, dto UpdateLiveRoomConfigDTO) (*LiveRoom, error)
	Heartbeat(ctx context.Context, roomID uuid.UUID) error
	List(ctx context.Context, p ListParams) ([]LiveRoom, int64, error)

	StartRecording(ctx context.Context, roomID uuid.UUID) (*LiveRecording, error)
	StopRecording(ctx context.Context, recordingID uuid.UUID) (*LiveRecording, error)
	ListRecordings(ctx context.Context, roomID uuid.UUID) ([]LiveRecording, error)

	ListParticipants(ctx context.Context, roomID uuid.UUID, p ListParams) ([]LiveParticipant, int64, error)

	AdminList(ctx context.Context, q AdminListLiveRoomsQuery) ([]LiveRoom, int64, error)
	AdminEndRoom(ctx context.Context, roomID uuid.UUID) (*LiveRoom, error)
	AdminHardDelete(ctx context.Context, roomID uuid.UUID) error

	AutoCloseStaleRooms(ctx context.Context) error
}
