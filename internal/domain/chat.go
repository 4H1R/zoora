package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatStatus string

const (
	ChatStatusActive   ChatStatus = "active"
	ChatStatusArchived ChatStatus = "archived"
)

func (s ChatStatus) Valid() bool {
	switch s {
	case ChatStatusActive, ChatStatusArchived:
		return true
	}
	return false
}

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

func (t MessageType) Valid() bool {
	switch t {
	case MessageTypeText, MessageTypeFile, MessageTypeSystem:
		return true
	}
	return false
}

type ChatMemberRole string

const (
	ChatMemberRoleAdmin    ChatMemberRole = "admin"
	ChatMemberRoleMember   ChatMemberRole = "member"
	ChatMemberRoleReadOnly ChatMemberRole = "read_only"
)

func (r ChatMemberRole) Valid() bool {
	switch r {
	case ChatMemberRoleAdmin, ChatMemberRoleMember, ChatMemberRoleReadOnly:
		return true
	}
	return false
}

type Chat struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text;not null;default:''" json:"description"`
	ModelType   string         `gorm:"type:varchar(100);not null" json:"model_type"`
	ModelID     uuid.UUID      `gorm:"type:uuid;not null" json:"model_id"`
	Status      ChatStatus     `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type ChatMember struct {
	ID       uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ChatID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"chat_id"`
	Chat     *Chat          `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	UserID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User     *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role     ChatMemberRole `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	JoinedAt time.Time      `gorm:"not null" json:"joined_at"`
}

type Message struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ChatID          uuid.UUID       `gorm:"type:uuid;not null;index" json:"chat_id"`
	Chat            *Chat           `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	SenderID        *uuid.UUID      `gorm:"type:uuid;index" json:"sender_id"`
	Sender          *User           `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ParentMessageID *uuid.UUID      `gorm:"type:uuid;index" json:"parent_message_id"`
	ParentMessage   *Message        `gorm:"foreignKey:ParentMessageID" json:"parent_message,omitempty"`
	MessageType     MessageType     `gorm:"type:varchar(20);not null;default:'text'" json:"message_type"`
	Content         string          `gorm:"type:text;not null;default:''" json:"content"`
	Attachments     json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"attachments"`
	EmojiCounts     json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"emoji_counts"`
	IsEdited        bool            `gorm:"not null;default:false" json:"is_edited"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}

type MessageReaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;index" json:"message_id"`
	Message   *Message  `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Emoji     string    `gorm:"type:varchar(32);not null" json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// DTOs

type CreateChatDTO struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
	ModelType   string `json:"model_type" binding:"required,max=100"`
	ModelID     string `json:"model_id" binding:"required,uuid"`
}

type UpdateChatDTO struct {
	Name        *string     `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string     `json:"description" binding:"omitempty,max=1000"`
	Status      *ChatStatus `json:"status" binding:"omitempty"`
}

type AddChatMemberDTO struct {
	UserID string         `json:"user_id" binding:"required,uuid"`
	Role   ChatMemberRole `json:"role" binding:"required"`
}

type SendMessageDTO struct {
	ParentMessageID *string     `json:"parent_message_id" binding:"omitempty,uuid"`
	MessageType     MessageType `json:"message_type" binding:"required"`
	Content         string      `json:"content" binding:"required,min=1,max=10000"`
	Attachments     *json.RawMessage `json:"attachments"`
}

type UpdateMessageDTO struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}

type ToggleReactionDTO struct {
	Emoji string `json:"emoji" binding:"required,min=1,max=32"`
}

type ListMessagesQuery struct {
	ParentMessageID *uuid.UUID `form:"parent_message_id"`
	ListParams      ListParams `form:"-"`
}

type ListChatsQuery struct {
	ModelType  string     `form:"model_type"`
	ModelID    *uuid.UUID `form:"model_id"`
	Status     *ChatStatus `form:"status"`
	ListParams ListParams  `form:"-"`
}

// Repositories

type ChatRepository interface {
	Create(ctx context.Context, chat *Chat) error
	FindByID(ctx context.Context, id uuid.UUID) (*Chat, error)
	Update(ctx context.Context, chat *Chat) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, q ListChatsQuery) ([]Chat, int64, error)
	FindByModel(ctx context.Context, modelType string, modelID uuid.UUID) (*Chat, error)
}

type ChatMemberRepository interface {
	Create(ctx context.Context, member *ChatMember) error
	FindByChatAndUser(ctx context.Context, chatID, userID uuid.UUID) (*ChatMember, error)
	Delete(ctx context.Context, chatID, userID uuid.UUID) error
	ListByChat(ctx context.Context, chatID uuid.UUID) ([]ChatMember, error)
}

type MessageRepository interface {
	Create(ctx context.Context, msg *Message) error
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	Update(ctx context.Context, msg *Message) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, chatID uuid.UUID, q ListMessagesQuery) ([]Message, int64, error)
}

type MessageReactionRepository interface {
	Create(ctx context.Context, r *MessageReaction) error
	Delete(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
	FindByMessageAndUser(ctx context.Context, messageID, userID uuid.UUID, emoji string) (*MessageReaction, error)
	CountByMessage(ctx context.Context, messageID uuid.UUID) (map[string]int, error)
}

// Services

type ChatService interface {
	CreateChat(ctx context.Context, dto CreateChatDTO) (*Chat, error)
	GetChat(ctx context.Context, id uuid.UUID) (*Chat, error)
	UpdateChat(ctx context.Context, id uuid.UUID, dto UpdateChatDTO) (*Chat, error)
	DeleteChat(ctx context.Context, id uuid.UUID) error
	ListChats(ctx context.Context, q ListChatsQuery) ([]Chat, int64, error)

	AddMember(ctx context.Context, chatID uuid.UUID, dto AddChatMemberDTO) (*ChatMember, error)
	RemoveMember(ctx context.Context, chatID, userID uuid.UUID) error
	ListMembers(ctx context.Context, chatID uuid.UUID) ([]ChatMember, error)

	SendMessage(ctx context.Context, chatID uuid.UUID, dto SendMessageDTO) (*Message, error)
	GetMessage(ctx context.Context, id uuid.UUID) (*Message, error)
	UpdateMessage(ctx context.Context, id uuid.UUID, dto UpdateMessageDTO) (*Message, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error
	ListMessages(ctx context.Context, chatID uuid.UUID, q ListMessagesQuery) ([]Message, int64, error)

	ToggleReaction(ctx context.Context, messageID uuid.UUID, dto ToggleReactionDTO) (*Message, error)

	FindChatByModel(ctx context.Context, modelType string, modelID uuid.UUID) (*Chat, error)
	ArchiveByModel(ctx context.Context, modelType string, modelID uuid.UUID) error
}
