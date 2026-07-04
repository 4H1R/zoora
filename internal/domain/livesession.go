package domain

import (
	"context"
	"encoding/json"
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

type ParticipantRole string

const (
	ParticipantRoleHost      ParticipantRole = "host"
	ParticipantRolePresenter ParticipantRole = "presenter"
	ParticipantRoleViewer    ParticipantRole = "viewer"
)

func (r ParticipantRole) Valid() bool {
	switch r {
	case ParticipantRoleHost, ParticipantRolePresenter, ParticipantRoleViewer:
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
	MaxParticipants int `json:"max_participants"`
}

func DefaultLiveRoomConfig() LiveRoomConfig {
	return LiveRoomConfig{
		MaxParticipants: 100,
	}
}

type LiveRoom struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ClassSessionID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"class_session_id"`
	ClassSession       *ClassSession  `gorm:"foreignKey:ClassSessionID" json:"class_session,omitempty"`
	Name               string         `gorm:"type:varchar(255);not null;default:''" json:"name"`
	LiveKitRoomName    string         `gorm:"column:livekit_room_name;type:varchar(255);not null;uniqueIndex" json:"livekit_room_name"`
	Status             LiveRoomStatus `gorm:"type:varchar(20);not null;default:'created'" json:"status"`
	Config             LiveRoomConfig `gorm:"type:jsonb;not null;serializer:json" json:"config"`
	ScheduledStartTime *time.Time     `json:"scheduled_start_time"`
	ActualStartTime    *time.Time     `json:"actual_start_time"`
	ActualEndTime      *time.Time     `json:"actual_end_time"`
	HostLastSeenAt     *time.Time     `json:"host_last_seen_at"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

type LiveParticipant struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	LiveRoomID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"live_room_id"`
	LiveRoom             *LiveRoom  `gorm:"foreignKey:LiveRoomID" json:"live_room,omitempty"`
	UserID               uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User                 *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Identity             string          `gorm:"type:varchar(255);not null" json:"identity"`
	Role                 ParticipantRole `gorm:"type:varchar(20);not null;default:'viewer'" json:"role"`
	HandRaisedAt         *time.Time      `json:"hand_raised_at"`
	JoinedAt             time.Time       `gorm:"not null" json:"joined_at"`
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
	ClassSessionID     uuid.UUID      `json:"class_session_id" binding:"required"`
	Name               string         `json:"name" binding:"required,max=255"`
	ScheduledStartTime *time.Time     `json:"scheduled_start_time"`
	Config             LiveRoomConfig `json:"config"`
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

// LiveRoomListScope is the role-resolved filter set the repository understands.
// All short-circuits role scoping but typed filters (status/class/etc.) still apply.
type LiveRoomListScope struct {
	All            bool
	OrganizationID *uuid.UUID
	TeacherID      *uuid.UUID
	MemberUserID   *uuid.UUID
	Status         *LiveRoomStatus
	ClassID        *uuid.UUID
	ClassSessionID *uuid.UUID
	IncludeDeleted bool
}

// ListLiveRoomsQuery is the query for GET /live-rooms. Typed filters sit
// alongside the embedded ListParams populated by the handler after white-listing.
type ListLiveRoomsQuery struct {
	Status         *LiveRoomStatus `form:"status"`
	ClassID        *uuid.UUID      `form:"-"`
	ClassSessionID *uuid.UUID      `form:"-"`
	IncludeDeleted bool            `form:"include_deleted"`
	ListParams     ListParams      `form:"-"`
}

// AdminListLiveRoomsQuery is the query for GET /admin/live-rooms.
type AdminListLiveRoomsQuery struct {
	Status         *LiveRoomStatus `form:"status"`
	UserID         *uuid.UUID      `form:"-"`
	ClassID        *uuid.UUID      `form:"-"`
	ClassSessionID *uuid.UUID      `form:"-"`
	IncludeDeleted bool            `form:"include_deleted"`
	ListParams     ListParams      `form:"-"`
}

// ListLiveParticipantsQuery is the query for GET /live-rooms/:id/participants.
type ListLiveParticipantsQuery struct {
	ActiveOnly *bool      `form:"active_only"`
	UserID     *uuid.UUID `form:"-"`
	ListParams ListParams `form:"-"`
}

// ListLiveRecordingsQuery is the query for GET /live-rooms/:id/recordings.
type ListLiveRecordingsQuery struct {
	Status     *LiveRecordingStatus `form:"status"`
	ListParams ListParams           `form:"-"`
}

type SetParticipantRoleDTO struct {
	Role ParticipantRole `json:"role" binding:"required"`
}

type MuteParticipantDTO struct {
	TrackSID string `json:"track_sid" binding:"required"`
	Muted    bool   `json:"muted"`
}

type SetHandDTO struct {
	Raised bool `json:"raised"`
}

type LiveRoomRepository interface {
	Create(ctx context.Context, room *LiveRoom) error
	FindByID(ctx context.Context, id uuid.UUID) (*LiveRoom, error)
	Update(ctx context.Context, room *LiveRoom) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope LiveRoomListScope, p ListParams) ([]LiveRoom, int64, error)
	FindActiveRoomsWithStaleHost(ctx context.Context, staleDuration time.Duration) ([]LiveRoom, error)
	FindByLiveKitRoomName(ctx context.Context, name string) (*LiveRoom, error)
	ListByClassSession(ctx context.Context, sessionID uuid.UUID) ([]LiveRoom, error)
	AdminList(ctx context.Context, q AdminListLiveRoomsQuery) ([]LiveRoom, int64, error)
	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*LiveRoom, error)
}

type LiveParticipantRepository interface {
	Create(ctx context.Context, p *LiveParticipant) error
	FindActiveByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*LiveParticipant, error)
	GetActiveParticipant(ctx context.Context, roomID uuid.UUID, identity string) (*LiveParticipant, error)
	Update(ctx context.Context, p *LiveParticipant) error
	UpdateParticipantRole(ctx context.Context, roomID uuid.UUID, identity string, role ParticipantRole) error
	SetHandRaised(ctx context.Context, roomID uuid.UUID, identity string, raised bool) error
	ListByRoom(ctx context.Context, roomID uuid.UUID, q ListLiveParticipantsQuery) ([]LiveParticipant, int64, error)
	ListAllByRoom(ctx context.Context, roomID uuid.UUID) ([]LiveParticipant, error)
	MarkAllLeft(ctx context.Context, roomID uuid.UUID, leftAt time.Time) error
}

type LiveRecordingRepository interface {
	Create(ctx context.Context, r *LiveRecording) error
	FindByID(ctx context.Context, id uuid.UUID) (*LiveRecording, error)
	FindActiveByRoom(ctx context.Context, roomID uuid.UUID) (*LiveRecording, error)
	Update(ctx context.Context, r *LiveRecording) error
	ListByRoom(ctx context.Context, roomID uuid.UUID, q ListLiveRecordingsQuery) ([]LiveRecording, int64, error)
}

// LiveWhiteboard persists the tldraw snapshot for a live room so that
// late-joiners and refreshers can load the current board state.
type LiveWhiteboard struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	LiveRoomID uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex" json:"live_room_id"`
	Snapshot   json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"snapshot" swaggertype:"object"`
	UpdatedAt  time.Time       `json:"updated_at"`
	CreatedAt  time.Time       `json:"created_at"`
}

// SaveWhiteboardDTO is the request body for PUT /live-rooms/:id/whiteboard.
type SaveWhiteboardDTO struct {
	Snapshot json.RawMessage `json:"snapshot" binding:"required" swaggertype:"object"`
}

// LiveWhiteboardRepository persists whiteboard snapshots keyed by live_room_id.
type LiveWhiteboardRepository interface {
	Get(ctx context.Context, roomID uuid.UUID) (*LiveWhiteboard, error)
	Upsert(ctx context.Context, roomID uuid.UUID, snapshot json.RawMessage) (*LiveWhiteboard, error)
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
	// List returns rooms visible to the caller under the RBAC hierarchy:
	// super-admin / view_any sees all, teacher sees own classes' rooms,
	// student sees rooms in classes they are enrolled in.
	List(ctx context.Context, q ListLiveRoomsQuery) ([]LiveRoom, int64, error)

	StartRecording(ctx context.Context, roomID uuid.UUID) (*LiveRecording, error)
	StopRecording(ctx context.Context, recordingID uuid.UUID) (*LiveRecording, error)
	ListRecordings(ctx context.Context, roomID uuid.UUID, q ListLiveRecordingsQuery) ([]LiveRecording, int64, error)

	ListParticipants(ctx context.Context, roomID uuid.UUID, q ListLiveParticipantsQuery) ([]LiveParticipant, int64, error)

	SetParticipantRole(ctx context.Context, roomID uuid.UUID, identity string, dto SetParticipantRoleDTO) (*LiveParticipant, error)
	MuteParticipant(ctx context.Context, roomID uuid.UUID, identity string, dto MuteParticipantDTO) error
	SetHand(ctx context.Context, roomID uuid.UUID, dto SetHandDTO) (*LiveParticipant, error)
	SetParticipantHand(ctx context.Context, roomID uuid.UUID, identity string, dto SetHandDTO) (*LiveParticipant, error)

	// GetWhiteboard returns the current snapshot for the room. Any participant
	// (viewer or above) may read it. Returns an empty board if none saved yet.
	GetWhiteboard(ctx context.Context, roomID uuid.UUID) (*LiveWhiteboard, error)
	// SaveWhiteboard persists a snapshot. Only hosts and presenters may write.
	SaveWhiteboard(ctx context.Context, roomID uuid.UUID, dto SaveWhiteboardDTO) (*LiveWhiteboard, error)

	AdminList(ctx context.Context, q AdminListLiveRoomsQuery) ([]LiveRoom, int64, error)
	AdminEndRoom(ctx context.Context, roomID uuid.UUID) (*LiveRoom, error)
	AdminHardDelete(ctx context.Context, roomID uuid.UUID) error

	AutoCloseStaleRooms(ctx context.Context) error

	// OnLiveKitEvent reacts to a verified LiveKit webhook. The HTTP handler
	// verifies the signature and passes only the event type and LiveKit room
	// name so this interface stays free of LiveKit types. Drives the
	// event-driven, no-host auto-close (grace-period timer arm/cancel) and
	// finalizes rooms LiveKit tears down on its own empty_timeout.
	OnLiveKitEvent(ctx context.Context, eventType, livekitRoomName string) error
	// CloseRoomIfNoHost is the delayed task target: it closes the room only if
	// LiveKit confirms no host is still present, so a host returning within the
	// grace window keeps the room alive.
	CloseRoomIfNoHost(ctx context.Context, roomID uuid.UUID) error
}
