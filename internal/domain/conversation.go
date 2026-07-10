package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PresenceStatus reports whether a user is currently online and, regardless of
// current online state, when they were last seen. It mirrors the chathub
// presence tracker's status across the service boundary so the conversations
// package need not import chathub.
type PresenceStatus struct {
	Online   bool      `json:"online"`
	LastSeen time.Time `json:"last_seen"`
}

type ConversationType string

const (
	ConversationTypeDirect  ConversationType = "direct"
	ConversationTypeGroup   ConversationType = "group"
	ConversationTypeChannel ConversationType = "channel"
)

func (t ConversationType) Valid() bool {
	switch t {
	case ConversationTypeDirect, ConversationTypeGroup, ConversationTypeChannel:
		return true
	}
	return false
}

type ConversationMemberRole string

const (
	ConversationMemberRoleAdmin  ConversationMemberRole = "admin"
	ConversationMemberRoleMember ConversationMemberRole = "member"
)

func (r ConversationMemberRole) Valid() bool {
	return r == ConversationMemberRoleAdmin || r == ConversationMemberRoleMember
}

type Conversation struct {
	ID             uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID        `gorm:"type:uuid;not null;index" json:"organization_id"`
	Type           ConversationType `gorm:"type:varchar(20);not null" json:"type"`
	Name           string           `gorm:"type:varchar(255);not null;default:''" json:"name"`
	Description    string           `gorm:"type:text;not null;default:''" json:"description"`
	AvatarURL      string           `gorm:"type:text;not null;default:''" json:"avatar_url"`
	ColorIndex     int16            `gorm:"not null;default:0" json:"color_index"`
	CreatedBy      *uuid.UUID       `gorm:"type:uuid" json:"created_by"`
	DirectKey      *string          `gorm:"type:text" json:"-"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`

	// Computed / preloaded, not columns.
	Members     []ConversationMember `gorm:"foreignKey:ConversationID" json:"members,omitempty"`
	UnreadCount int64                `gorm:"-" json:"unread_count"`
	LastMessage *ConversationMessage `gorm:"-" json:"last_message,omitempty"`
}

type ConversationMember struct {
	ID                uuid.UUID              `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ConversationID    uuid.UUID              `gorm:"type:uuid;not null;index" json:"conversation_id"`
	UserID            uuid.UUID              `gorm:"type:uuid;not null;index" json:"user_id"`
	User              *User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role              ConversationMemberRole `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	LastReadMessageID *uuid.UUID             `gorm:"type:uuid" json:"last_read_message_id"`
	LastReadAt        *time.Time             `json:"last_read_at"`
	MutedUntil        *time.Time             `json:"muted_until"`
	JoinedAt          time.Time              `gorm:"not null" json:"joined_at"`
}

type ConversationMessage struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ConversationID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"conversation_id"`
	SenderID         *uuid.UUID      `gorm:"type:uuid;index" json:"sender_id"`
	Sender           *User           `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReplyToMessageID *uuid.UUID      `gorm:"type:uuid" json:"reply_to_message_id"`
	Content          string          `gorm:"type:text;not null;default:''" json:"content"`
	IsEdited         bool            `gorm:"not null;default:false" json:"is_edited"`
	IsPinned         bool            `gorm:"not null;default:false" json:"is_pinned"`
	PinnedBy         *uuid.UUID      `gorm:"type:uuid" json:"pinned_by"`
	PinnedAt         *time.Time      `json:"pinned_at"`
	MediaIDs         json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"media_ids"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`

	Reactions map[string]int `gorm:"-" json:"reactions,omitempty"`
}

type ConversationMessageReaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;index" json:"message_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Emoji     string    `gorm:"type:varchar(32);not null" json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversationMention struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;index" json:"message_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ---- DTOs ----

type CreateConversationDTO struct {
	Type        ConversationType `json:"type" binding:"required"`
	Name        string           `json:"name" binding:"omitempty,max=255"`
	Description string           `json:"description" binding:"omitempty,max=1000"`
	ColorIndex  int16            `json:"color_index" binding:"omitempty,min=0,max=6"`
	MemberIDs   []string         `json:"member_ids" binding:"omitempty,max=500,dive,uuid"`
}

type CreateDirectDTO struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

type UpdateConversationDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	AvatarURL   *string `json:"avatar_url" binding:"omitempty,max=1000"`
	ColorIndex  *int16  `json:"color_index" binding:"omitempty,min=0,max=6"`
}

type AddConversationMemberDTO struct {
	UserID string                 `json:"user_id" binding:"required,uuid"`
	Role   ConversationMemberRole `json:"role" binding:"omitempty"`
}

type SendConversationMessageDTO struct {
	ID               *string  `json:"id" binding:"omitempty,uuid"` // client-supplied uuidv7 (idempotency)
	Content          string   `json:"content" binding:"required,min=1,max=10000"`
	ReplyToMessageID *string  `json:"reply_to_message_id" binding:"omitempty,uuid"`
	MentionUserIDs   []string `json:"mentions" binding:"omitempty,max=100,dive,uuid"`
	MediaIDs         []string `json:"media_ids" binding:"omitempty,max=20,dive,uuid"`
}

type UpdateConversationMessageDTO struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}

type ToggleConversationReactionDTO struct {
	Emoji string `json:"emoji" binding:"required,min=1,max=32"`
}

type MarkReadDTO struct {
	MessageID string `json:"message_id" binding:"required,uuid"`
}

// ---- Query / cursor types ----

// MessageCursor selects a keyset window within a conversation.
// Exactly one of Before/After/Around may be set; none = latest page.
type MessageCursor struct {
	Before *uuid.UUID
	After  *uuid.UUID
	Around *uuid.UUID
	Limit  int // clamped 1..100 by service
}

type ListConversationsQuery struct {
	Type       *ConversationType
	ListParams ListParams
}

// ---- Interfaces ----

type ConversationRepository interface {
	Create(ctx context.Context, c *Conversation) error
	FindByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	Update(ctx context.Context, c *Conversation) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindDirect(ctx context.Context, orgID uuid.UUID, directKey string) (*Conversation, error)
	// ListForUser returns conversations the user is a member of, org-scoped, paginated.
	ListForUser(ctx context.Context, orgID, userID uuid.UUID, q ListConversationsQuery) ([]Conversation, int64, error)
	Touch(ctx context.Context, id uuid.UUID) error // bump updated_at on new message
}

type ConversationMemberRepository interface {
	Create(ctx context.Context, m *ConversationMember) error
	CreateMany(ctx context.Context, members []ConversationMember) error
	FindByConversationAndUser(ctx context.Context, convID, userID uuid.UUID) (*ConversationMember, error)
	Delete(ctx context.Context, convID, userID uuid.UUID) error
	ListByConversation(ctx context.Context, convID uuid.UUID) ([]ConversationMember, error)
	ListUserIDs(ctx context.Context, convID uuid.UUID) ([]uuid.UUID, error)
	SetLastRead(ctx context.Context, convID, userID, messageID uuid.UUID, at time.Time) error
	SetMuted(ctx context.Context, convID, userID uuid.UUID, until *time.Time) error
	UnreadCount(ctx context.Context, convID, userID uuid.UUID) (int64, error)
}

type ConversationMessageRepository interface {
	Create(ctx context.Context, m *ConversationMessage) error
	FindByID(ctx context.Context, id uuid.UUID) (*ConversationMessage, error)
	Update(ctx context.Context, m *ConversationMessage) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListWindow(ctx context.Context, convID uuid.UUID, cur MessageCursor) ([]ConversationMessage, error)
	Latest(ctx context.Context, convID uuid.UUID) (*ConversationMessage, error)
	ListPinned(ctx context.Context, convID uuid.UUID) ([]ConversationMessage, error)
	SetPinned(ctx context.Context, id uuid.UUID, pinned bool, by *uuid.UUID, at *time.Time) error
	// SearchInConversation: ILIKE nav search (Phase 3 adds full-text global).
	SearchInConversation(ctx context.Context, convID uuid.UUID, q string, limit int) ([]ConversationMessage, error)
	// SearchGlobal: ranked full-text search across all conversations the user
	// is a member of, org-scoped.
	SearchGlobal(ctx context.Context, orgID, userID uuid.UUID, q string, limit int) ([]ConversationMessage, error)
}

type ConversationReactionRepository interface {
	Create(ctx context.Context, r *ConversationMessageReaction) error
	Delete(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
	FindByMessageAndUser(ctx context.Context, messageID, userID uuid.UUID, emoji string) (*ConversationMessageReaction, error)
	CountByMessage(ctx context.Context, messageID uuid.UUID) (map[string]int, error)
}

type ConversationMentionRepository interface { // Phase 3
	CreateMany(ctx context.Context, messageID uuid.UUID, userIDs []uuid.UUID) error
}

type ConversationService interface {
	CreateGroupOrChannel(ctx context.Context, dto CreateConversationDTO) (*Conversation, error)
	CreateOrGetDirect(ctx context.Context, dto CreateDirectDTO) (*Conversation, error)
	Get(ctx context.Context, id uuid.UUID) (*Conversation, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateConversationDTO) (*Conversation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListForCaller(ctx context.Context, q ListConversationsQuery) ([]Conversation, int64, error)

	AddMember(ctx context.Context, convID uuid.UUID, dto AddConversationMemberDTO) (*ConversationMember, error)
	RemoveMember(ctx context.Context, convID, userID uuid.UUID) error
	ListMembers(ctx context.Context, convID uuid.UUID) ([]ConversationMember, error)
	Leave(ctx context.Context, convID uuid.UUID) error

	SendMessage(ctx context.Context, convID uuid.UUID, dto SendConversationMessageDTO) (*ConversationMessage, error)
	ListMessages(ctx context.Context, convID uuid.UUID, cur MessageCursor) ([]ConversationMessage, error)
	EditMessage(ctx context.Context, msgID uuid.UUID, dto UpdateConversationMessageDTO) (*ConversationMessage, error)
	DeleteMessage(ctx context.Context, msgID uuid.UUID) error
	ToggleReaction(ctx context.Context, msgID uuid.UUID, dto ToggleConversationReactionDTO) (*ConversationMessage, error)

	MarkRead(ctx context.Context, convID uuid.UUID, dto MarkReadDTO) error
	SetMuted(ctx context.Context, convID uuid.UUID, until *time.Time) error

	PinMessage(ctx context.Context, msgID uuid.UUID) error
	UnpinMessage(ctx context.Context, msgID uuid.UUID) error
	ListPinned(ctx context.Context, convID uuid.UUID) ([]ConversationMessage, error)

	// Search performs a global ranked full-text search across all
	// conversations the caller is a member of.
	Search(ctx context.Context, q string, limit int) ([]ConversationMessage, error)
	// SearchInConversation performs an in-conversation ILIKE nav search.
	SearchInConversation(ctx context.Context, convID uuid.UUID, q string, limit int) ([]ConversationMessage, error)

	// Presence returns the online/last-seen status for each requested user id,
	// filtered to users in the caller's organization.
	Presence(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]PresenceStatus, error)
}
