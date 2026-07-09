package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LiveRoomChatStatus string

const (
	LiveRoomChatStatusActive   LiveRoomChatStatus = "active"
	LiveRoomChatStatusArchived LiveRoomChatStatus = "archived"
)

func (s LiveRoomChatStatus) Valid() bool {
	return s == LiveRoomChatStatusActive || s == LiveRoomChatStatusArchived
}

type LiveRoomMessageType string

const (
	LiveRoomMessageTypeText   LiveRoomMessageType = "text"
	LiveRoomMessageTypeFile   LiveRoomMessageType = "file"
	LiveRoomMessageTypeSystem LiveRoomMessageType = "system"
)

func (t LiveRoomMessageType) Valid() bool {
	switch t {
	case LiveRoomMessageTypeText, LiveRoomMessageTypeFile, LiveRoomMessageTypeSystem:
		return true
	}
	return false
}

type LiveRoomChat struct {
	ID          uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Name        string             `gorm:"type:varchar(255);not null" json:"name"`
	Description string             `gorm:"type:text;not null;default:''" json:"description"`
	LiveRoomID  uuid.UUID          `gorm:"type:uuid;not null;index" json:"live_room_id"`
	Status      LiveRoomChatStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (LiveRoomChat) TableName() string { return "liveroom_chats" }

type LiveRoomMessage struct {
	ID              uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ChatID          uuid.UUID           `gorm:"type:uuid;not null;index" json:"chat_id"`
	SenderID        *uuid.UUID          `gorm:"type:uuid;index" json:"sender_id"`
	Sender          *User               `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ParentMessageID *uuid.UUID          `gorm:"type:uuid;index" json:"parent_message_id"`
	MessageType     LiveRoomMessageType `gorm:"type:varchar(20);not null;default:'text'" json:"message_type"`
	Content         string              `gorm:"type:text;not null;default:''" json:"content"`
	Attachments     json.RawMessage     `gorm:"type:jsonb;not null;default:'[]'" json:"attachments"`
	IsEdited        bool                `gorm:"not null;default:false" json:"is_edited"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}

func (LiveRoomMessage) TableName() string { return "liveroom_messages" }

type CreateChatDTO struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
	LiveRoomID  string `json:"live_room_id" binding:"required,uuid"`
}

type UpdateChatDTO struct {
	Name        *string             `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string             `json:"description" binding:"omitempty,max=1000"`
	Status      *LiveRoomChatStatus `json:"status" binding:"omitempty"`
}

type SendMessageDTO struct {
	ParentMessageID *string             `json:"parent_message_id" binding:"omitempty,uuid"`
	MessageType     LiveRoomMessageType `json:"message_type" binding:"required"`
	Content         string              `json:"content" binding:"required,min=1,max=10000"`
	Attachments     *json.RawMessage    `json:"attachments"`
}

type UpdateMessageDTO struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}

type ListMessagesQuery struct {
	ParentMessageID *uuid.UUID `form:"-"`
	ListParams      ListParams `form:"-"`
}

type ListChatsQuery struct {
	LiveRoomID *uuid.UUID          `form:"-"`
	Status     *LiveRoomChatStatus `form:"status"`
	ListParams ListParams          `form:"-"`
}

type LiveRoomChatRepository interface {
	Create(ctx context.Context, chat *LiveRoomChat) error
	FindByID(ctx context.Context, id uuid.UUID) (*LiveRoomChat, error)
	Update(ctx context.Context, chat *LiveRoomChat) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, q ListChatsQuery) ([]LiveRoomChat, int64, error)
	FindByRoom(ctx context.Context, liveRoomID uuid.UUID) (*LiveRoomChat, error)
}

type LiveRoomMessageRepository interface {
	Create(ctx context.Context, msg *LiveRoomMessage) error
	FindByID(ctx context.Context, id uuid.UUID) (*LiveRoomMessage, error)
	Update(ctx context.Context, msg *LiveRoomMessage) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, chatID uuid.UUID, q ListMessagesQuery) ([]LiveRoomMessage, int64, error)
}

type LiveRoomChatService interface {
	CreateChat(ctx context.Context, dto CreateChatDTO) (*LiveRoomChat, error)
	GetChat(ctx context.Context, id uuid.UUID) (*LiveRoomChat, error)
	UpdateChat(ctx context.Context, id uuid.UUID, dto UpdateChatDTO) (*LiveRoomChat, error)
	DeleteChat(ctx context.Context, id uuid.UUID) error
	ListChats(ctx context.Context, q ListChatsQuery) ([]LiveRoomChat, int64, error)

	SendMessage(ctx context.Context, chatID uuid.UUID, dto SendMessageDTO) (*LiveRoomMessage, error)
	GetMessage(ctx context.Context, id uuid.UUID) (*LiveRoomMessage, error)
	UpdateMessage(ctx context.Context, id uuid.UUID, dto UpdateMessageDTO) (*LiveRoomMessage, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error
	ListMessages(ctx context.Context, chatID uuid.UUID, q ListMessagesQuery) ([]LiveRoomMessage, int64, error)

	// FindChatByRoom and ArchiveByRoom back the live-room lifecycle (join/end
	// room) in the livesessions package; they carry no caller-based authz since
	// the caller has already been authorized to act on the room itself.
	FindChatByRoom(ctx context.Context, liveRoomID uuid.UUID) (*LiveRoomChat, error)
	ArchiveByRoom(ctx context.Context, liveRoomID uuid.UUID) error
}
