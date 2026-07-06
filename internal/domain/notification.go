package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NotificationCategory string

const (
	NotificationCategorySystem   NotificationCategory = "system"
	NotificationCategoryOrg      NotificationCategory = "org"
	NotificationCategoryClass    NotificationCategory = "class"
	NotificationCategoryReminder NotificationCategory = "reminder" // reserved for the future Bell feature
)

type NotificationAudienceType string

const (
	AudienceAll   NotificationAudienceType = "all"
	AudienceOrg   NotificationAudienceType = "org"
	AudienceClass NotificationAudienceType = "class"
	AudienceRole  NotificationAudienceType = "role"
	AudienceUsers NotificationAudienceType = "users"
)

// NotificationAudience is the targeting spec stored as JSONB on the
// notification. The worker resolves it to concrete user IDs at fan-out time,
// so late joiners are NOT retro-notified (resolution is send-time).
// For AudienceRole, OrgID narrows the role to one org; nil = platform-wide
// (superAdmin only — service enforces).
type NotificationAudience struct {
	Type    NotificationAudienceType `json:"type"`
	OrgID   *uuid.UUID               `json:"org_id,omitempty"`
	ClassID *uuid.UUID               `json:"class_id,omitempty"`
	RoleID  *uuid.UUID               `json:"role_id,omitempty"`
	UserIDs []uuid.UUID              `json:"user_ids,omitempty"`
}

func (a NotificationAudience) Validate() error {
	switch a.Type {
	case AudienceAll:
		return nil
	case AudienceOrg:
		if a.OrgID == nil {
			return NewValidationError(map[string]string{"audience.org_id": "required for org audience"})
		}
	case AudienceClass:
		if a.ClassID == nil {
			return NewValidationError(map[string]string{"audience.class_id": "required for class audience"})
		}
	case AudienceRole:
		if a.RoleID == nil {
			return NewValidationError(map[string]string{"audience.role_id": "required for role audience"})
		}
	case AudienceUsers:
		if len(a.UserIDs) == 0 {
			return NewValidationError(map[string]string{"audience.user_ids": "required for users audience"})
		}
	default:
		return NewValidationError(map[string]string{"audience.type": "must be one of all, org, class, role, users"})
	}
	return nil
}

// Value / Scan make the audience a JSONB column.
func (a NotificationAudience) Value() (driver.Value, error) { return json.Marshal(a) }

func (a *NotificationAudience) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		if s, ok := value.(string); ok {
			b = []byte(s)
		} else {
			return fmt.Errorf("scanning NotificationAudience: unexpected type %T", value)
		}
	}
	return json.Unmarshal(b, a)
}

// Notification is one sent message. SenderID is nullable so system-generated
// notifications (future Bell reminders) have no human sender.
type Notification struct {
	ID             uuid.UUID            `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	SenderID       *uuid.UUID           `gorm:"type:uuid;index" json:"sender_id,omitempty"`
	OrganizationID *uuid.UUID           `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	Category       NotificationCategory `gorm:"type:varchar(20);not null" json:"category"`
	Title          string               `gorm:"type:varchar(255);not null" json:"title"`
	Body           string               `gorm:"type:text;not null" json:"body"`
	ActionURL      *string              `gorm:"type:varchar(500)" json:"action_url,omitempty"`
	Audience       NotificationAudience `gorm:"type:jsonb;not null" json:"audience"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

func (Notification) TableName() string { return "notifications" }

// NotificationRecipient is one user's inbox row (write fan-out).
type NotificationRecipient struct {
	NotificationID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"notification_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (NotificationRecipient) TableName() string { return "notification_recipients" }

type NotificationDeliveryStatus string

const (
	DeliveryPending NotificationDeliveryStatus = "pending"
	DeliverySent    NotificationDeliveryStatus = "sent"
	DeliveryFailed  NotificationDeliveryStatus = "failed"
)

// NotificationDelivery snapshots one channel send. Target is copied from the
// connector at fan-out time so later unlinking can't break in-flight sends.
type NotificationDelivery struct {
	ID             uuid.UUID                  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	NotificationID uuid.UUID                  `gorm:"type:uuid;not null;index" json:"notification_id"`
	UserID         uuid.UUID                  `gorm:"type:uuid;not null" json:"user_id"`
	Channel        ConnectorType              `gorm:"type:varchar(20);not null" json:"channel"`
	Target         string                     `gorm:"type:varchar(500);not null" json:"target"`
	Status         NotificationDeliveryStatus `gorm:"type:varchar(10);not null;default:'pending'" json:"status"`
	Error          *string                    `json:"error,omitempty"`
	SentAt         *time.Time                 `json:"sent_at,omitempty"`
	CreatedAt      time.Time                  `json:"created_at"`
}

func (NotificationDelivery) TableName() string { return "notification_deliveries" }

// NotificationDeliveryReport aggregates delivery outcomes for the sender UI.
type NotificationDeliveryReport struct {
	Recipients int64                       `json:"recipients"`
	Channels   []NotificationChannelReport `json:"channels"`
}

type NotificationChannelReport struct {
	Channel ConnectorType `json:"channel"`
	Pending int64         `json:"pending"`
	Sent    int64         `json:"sent"`
	Failed  int64         `json:"failed"`
}

// --- DTOs ---

type NotificationAudienceDTO struct {
	Type    string      `json:"type" binding:"required,oneof=all org class role users"`
	OrgID   *uuid.UUID  `json:"org_id"`
	ClassID *uuid.UUID  `json:"class_id"`
	RoleID  *uuid.UUID  `json:"role_id"`
	UserIDs []uuid.UUID `json:"user_ids" binding:"omitempty,max=500"`
}

type SendNotificationDTO struct {
	Title     string                  `json:"title" binding:"required,max=255"`
	Body      string                  `json:"body" binding:"required,max=4000"`
	ActionURL *string                 `json:"action_url" binding:"omitempty,max=500"`
	Audience  NotificationAudienceDTO `json:"audience" binding:"required"`
}

// NotificationInboxItem is a notification joined with the caller's read state.
type NotificationInboxItem struct {
	Notification
	ReadAt *time.Time `json:"read_at,omitempty"`
}

// NotificationStatus drives the header bell badge.
type NotificationStatus struct {
	UnreadCount int64 `json:"unread_count"`
}

// --- interfaces ---

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	// CreateRecipients bulk-inserts inbox rows, ignoring conflicts so fan-out
	// retries are idempotent.
	CreateRecipients(ctx context.Context, recipients []NotificationRecipient) error
	ListInbox(ctx context.Context, userID uuid.UUID, limit, offset int) ([]NotificationInboxItem, int64, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkRead(ctx context.Context, notificationID, userID uuid.UUID, t time.Time) error
	MarkAllRead(ctx context.Context, userID uuid.UUID, t time.Time) error
	ListBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]Notification, int64, error)
	// CountBySenderSince powers the per-sender send rate limit.
	CountBySenderSince(ctx context.Context, senderID uuid.UUID, since time.Time) (int64, error)

	// Audience-resolution primitives (active = not deleted, not disabled).
	ListAllActiveUserIDs(ctx context.Context) ([]uuid.UUID, error)
	ListUserIDsByOrg(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error)
	ListUserIDsByClass(ctx context.Context, classID uuid.UUID) ([]uuid.UUID, error)
	// ListUserIDsByRole optionally narrows to one org (nil = platform-wide).
	ListUserIDsByRole(ctx context.Context, roleID uuid.UUID, orgID *uuid.UUID) ([]uuid.UUID, error)
	// CountActiveUsersByIDs validates a users-audience: with non-nil orgID all
	// ids must belong to that org to count.
	CountActiveUsersByIDs(ctx context.Context, ids []uuid.UUID, orgID *uuid.UUID) (int64, error)
	// CountUsersInClassesOwnedBy counts how many of ids are members of at
	// least one class owned by ownerID (teacher users-audience validation).
	CountUsersInClassesOwnedBy(ctx context.Context, ids []uuid.UUID, ownerID uuid.UUID) (int64, error)
	// RoleExistsInScope reports whether the role exists and is visible to the
	// org: role.organization_id = orgID OR role is a global preset. Nil orgID
	// (superAdmin) matches any role.
	RoleExistsInScope(ctx context.Context, roleID uuid.UUID, orgID *uuid.UUID) (bool, error)

	// CreateDeliveries bulk-inserts pending delivery rows, ignoring conflicts
	// so fan-out retries are idempotent. Rows get IDs populated.
	CreateDeliveries(ctx context.Context, deliveries []NotificationDelivery) error
	ListDeliveriesByIDs(ctx context.Context, ids []uuid.UUID) ([]NotificationDelivery, error)
	ListPendingDeliveries(ctx context.Context, notificationID uuid.UUID, channel ConnectorType) ([]NotificationDelivery, error)
	MarkDeliveries(ctx context.Context, ids []uuid.UUID, status NotificationDeliveryStatus, errMsg *string, sentAt time.Time) error
	CountRecipients(ctx context.Context, notificationID uuid.UUID) (int64, error)
	DeliveryReport(ctx context.Context, notificationID uuid.UUID) ([]NotificationChannelReport, error)
}

type NotificationService interface {
	Send(ctx context.Context, dto SendNotificationDTO) (*Notification, error)
	ListInbox(ctx context.Context, p ListParams) ([]NotificationInboxItem, int64, error)
	Status(ctx context.Context) (*NotificationStatus, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	MarkAllRead(ctx context.Context) error
	ListSent(ctx context.Context, p ListParams) ([]Notification, int64, error)
	// Fanout is the worker entry: resolves the audience and inserts recipients.
	Fanout(ctx context.Context, notificationID uuid.UUID) error
	// Report returns the delivery report; sender or superAdmin only.
	Report(ctx context.Context, notificationID uuid.UUID) (*NotificationDeliveryReport, error)
	// DeliverBot / DeliverSMS / DeliverPush are worker task entries.
	DeliverBot(ctx context.Context, deliveryID uuid.UUID) error
	DeliverSMS(ctx context.Context, notificationID uuid.UUID, deliveryIDs []uuid.UUID) error
	DeliverPush(ctx context.Context, notificationID uuid.UUID, deliveryIDs []uuid.UUID) error
}
