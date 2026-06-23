package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OfflineRoom struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	ClassID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"class_id"`
	Class          *Class         `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	ClassSessionID uuid.UUID      `gorm:"type:uuid;not null;index" json:"class_session_id"`
	ClassSession   *ClassSession  `gorm:"foreignKey:ClassSessionID" json:"class_session,omitempty"`
	CreatorID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"creator_id"`
	Creator        *User          `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Title          string         `gorm:"not null" json:"title"`
	Description    string         `gorm:"type:text;not null;default:''" json:"description"`
	PublishedAt    *time.Time     `json:"published_at"`
	ViewCount      int64          `gorm:"not null;default:0" json:"view_count"`
	Attachments    []Media        `gorm:"-" json:"attachments,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateOfflineRoomDTO struct {
	ClassSessionID uuid.UUID  `json:"class_session_id" binding:"required"`
	Title          string     `json:"title" binding:"required,min=2"`
	Description    string     `json:"description"`
	PublishedAt    *time.Time `json:"published_at"`
}

type UpdateOfflineRoomDTO struct {
	Title       *string    `json:"title" binding:"omitempty,min=2"`
	Description *string    `json:"description"`
	PublishedAt *time.Time `json:"published_at"`
}

// OfflineRoomListScope is the role-resolved view onto offline_rooms that the
// repository understands. The service builds it from the Caller; the repo
// only knows how to translate it into SQL filters.
type OfflineRoomListScope struct {
	All            bool
	OrganizationID *uuid.UUID
	OwnerID        *uuid.UUID
	MemberUserID   *uuid.UUID
}

// UUID query params (class_id, class_session_id, etc.) use form:"-" because
// Gin's form binder cannot decode strings into uuid.UUID's [16]byte underlying
// type. Handlers populate these via httpx.BindUUIDQueries.
type ListOfflineRoomsQuery struct {
	ClassID        *uuid.UUID `form:"-"`
	ClassSessionID *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type AdminListOfflineRoomsQuery struct {
	ClassID        *uuid.UUID `form:"-"`
	ClassSessionID *uuid.UUID `form:"-"`
	CreatorID      *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type OfflineRoomView struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OfflineRoomID  uuid.UUID    `gorm:"type:uuid;not null;index" json:"offline_room_id"`
	OfflineRoom    *OfflineRoom `gorm:"foreignKey:OfflineRoomID" json:"offline_room,omitempty"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	User           *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ViewedAt       time.Time    `gorm:"not null;default:NOW()" json:"viewed_at"`
	DurationSeconds int         `gorm:"not null;default:0" json:"duration_seconds"`
}

type OfflineRoomRepository interface {
	Create(ctx context.Context, room *OfflineRoom) error
	FindByID(ctx context.Context, id uuid.UUID) (*OfflineRoom, error)
	Update(ctx context.Context, room *OfflineRoom) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope OfflineRoomListScope, q ListOfflineRoomsQuery) ([]OfflineRoom, int64, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*OfflineRoom, error)
	AdminList(ctx context.Context, q AdminListOfflineRoomsQuery) ([]OfflineRoom, int64, error)
}

type OfflineRoomViewRepository interface {
	Create(ctx context.Context, v *OfflineRoomView) error
	ListByRoom(ctx context.Context, roomID uuid.UUID) ([]OfflineRoomView, error)
	ListDistinctUsersByRoom(ctx context.Context, roomID uuid.UUID) ([]uuid.UUID, error)
}

type OfflineService interface {
	CreateRoom(ctx context.Context, dto CreateOfflineRoomDTO) (*OfflineRoom, error)
	GetRoom(ctx context.Context, id uuid.UUID) (*OfflineRoom, error)
	UpdateRoom(ctx context.Context, id uuid.UUID, dto UpdateOfflineRoomDTO) (*OfflineRoom, error)
	DeleteRoom(ctx context.Context, id uuid.UUID) error
	ListRooms(ctx context.Context, q ListOfflineRoomsQuery) ([]OfflineRoom, int64, error)

	AdminList(ctx context.Context, q AdminListOfflineRoomsQuery) ([]OfflineRoom, int64, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}
