package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

const recipientInsertBatchSize = 1000

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.NotificationRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, n *domain.Notification) error {
	if err := database.DB(ctx, r.db).Create(n).Error; err != nil {
		return fmt.Errorf("notifications.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	var n domain.Notification
	if err := database.DB(ctx, r.db).First(&n, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("notifications.repository.FindByID: %w", err)
	}
	return &n, nil
}

func (r *repository) CreateRecipients(ctx context.Context, recipients []domain.NotificationRecipient) error {
	if len(recipients) == 0 {
		return nil
	}
	err := database.DB(ctx, r.db).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(recipients, recipientInsertBatchSize).Error
	if err != nil {
		return fmt.Errorf("notifications.repository.CreateRecipients: %w", err)
	}
	return nil
}

func (r *repository) ListInbox(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.NotificationInboxItem, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.NotificationRecipient{}).
		Joins("JOIN notifications ON notifications.id = notification_recipients.notification_id").
		Where("notification_recipients.user_id = ?", userID)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("notifications.repository.ListInbox count: %w", err)
	}
	var items []domain.NotificationInboxItem
	if err := base.
		Select("notifications.*, notification_recipients.read_at").
		Order("notification_recipients.created_at DESC, notifications.id DESC").
		Limit(limit).Offset(offset).
		Scan(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("notifications.repository.ListInbox: %w", err)
	}
	return items, total, nil
}

func (r *repository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	err := database.DB(ctx, r.db).Model(&domain.NotificationRecipient{}).
		Where("user_id = ? AND read_at IS NULL", userID).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("notifications.repository.CountUnread: %w", err)
	}
	return n, nil
}

func (r *repository) MarkRead(ctx context.Context, notificationID, userID uuid.UUID, t time.Time) error {
	res := database.DB(ctx, r.db).Model(&domain.NotificationRecipient{}).
		Where("notification_id = ? AND user_id = ? AND read_at IS NULL", notificationID, userID).
		Update("read_at", t)
	if res.Error != nil {
		return fmt.Errorf("notifications.repository.MarkRead: %w", res.Error)
	}
	// Row missing entirely → not the caller's notification.
	if res.RowsAffected == 0 {
		var count int64
		if err := database.DB(ctx, r.db).Model(&domain.NotificationRecipient{}).
			Where("notification_id = ? AND user_id = ?", notificationID, userID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("notifications.repository.MarkRead exists: %w", err)
		}
		if count == 0 {
			return domain.ErrNotFound
		}
	}
	return nil
}

func (r *repository) MarkAllRead(ctx context.Context, userID uuid.UUID, t time.Time) error {
	err := database.DB(ctx, r.db).Model(&domain.NotificationRecipient{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", t).Error
	if err != nil {
		return fmt.Errorf("notifications.repository.MarkAllRead: %w", err)
	}
	return nil
}

func (r *repository) ListBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Notification, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Notification{}).
		Where("sender_id = ?", senderID)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("notifications.repository.ListBySender count: %w", err)
	}
	var items []domain.Notification
	if err := base.Order("created_at DESC, id DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("notifications.repository.ListBySender: %w", err)
	}
	return items, total, nil
}

func (r *repository) CountBySenderSince(ctx context.Context, senderID uuid.UUID, since time.Time) (int64, error) {
	var n int64
	err := database.DB(ctx, r.db).Model(&domain.Notification{}).
		Where("sender_id = ? AND created_at >= ?", senderID, since).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("notifications.repository.CountBySenderSince: %w", err)
	}
	return n, nil
}

// activeUsers scopes to enabled, non-deleted accounts.
func (r *repository) activeUsers(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.User{}).
		Where("deleted_at IS NULL AND disabled_at IS NULL")
}

func (r *repository) ListAllActiveUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := r.activeUsers(ctx).Pluck("id", &ids).Error; err != nil {
		return nil, fmt.Errorf("notifications.repository.ListAllActiveUserIDs: %w", err)
	}
	return ids, nil
}

func (r *repository) ListUserIDsByOrg(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := r.activeUsers(ctx).
		Where("organization_id = ?", orgID).Pluck("id", &ids).Error; err != nil {
		return nil, fmt.Errorf("notifications.repository.ListUserIDsByOrg: %w", err)
	}
	return ids, nil
}

func (r *repository) ListUserIDsByClass(ctx context.Context, classID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := database.DB(ctx, r.db).Model(&domain.ClassMember{}).
		Joins("JOIN users ON users.id = class_members.user_id AND users.deleted_at IS NULL AND users.disabled_at IS NULL").
		Where("class_members.class_id = ?", classID).
		Pluck("class_members.user_id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("notifications.repository.ListUserIDsByClass: %w", err)
	}
	return ids, nil
}

func (r *repository) ListUserIDsByRole(ctx context.Context, roleID uuid.UUID, orgID *uuid.UUID) ([]uuid.UUID, error) {
	q := r.activeUsers(ctx).Where("role_id = ?", roleID)
	if orgID != nil {
		q = q.Where("organization_id = ?", *orgID)
	}
	var ids []uuid.UUID
	if err := q.Pluck("id", &ids).Error; err != nil {
		return nil, fmt.Errorf("notifications.repository.ListUserIDsByRole: %w", err)
	}
	return ids, nil
}

func (r *repository) CountActiveUsersByIDs(ctx context.Context, ids []uuid.UUID, orgID *uuid.UUID) (int64, error) {
	q := r.activeUsers(ctx).Where("id IN ?", ids)
	if orgID != nil {
		q = q.Where("organization_id = ?", *orgID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return 0, fmt.Errorf("notifications.repository.CountActiveUsersByIDs: %w", err)
	}
	return n, nil
}

func (r *repository) CountUsersInClassesOwnedBy(ctx context.Context, ids []uuid.UUID, ownerID uuid.UUID) (int64, error) {
	var n int64
	err := database.DB(ctx, r.db).Model(&domain.ClassMember{}).
		Joins("JOIN classes ON classes.id = class_members.class_id").
		Where("classes.user_id = ? AND class_members.user_id IN ?", ownerID, ids).
		Distinct("class_members.user_id").Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("notifications.repository.CountUsersInClassesOwnedBy: %w", err)
	}
	return n, nil
}

func (r *repository) RoleExistsInScope(ctx context.Context, roleID uuid.UUID, orgID *uuid.UUID) (bool, error) {
	q := database.DB(ctx, r.db).Model(&domain.Role{}).Where("id = ?", roleID)
	if orgID != nil {
		q = q.Where("organization_id = ? OR is_preset = true", *orgID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, fmt.Errorf("notifications.repository.RoleExistsInScope: %w", err)
	}
	return n > 0, nil
}

func (r *repository) CreateDeliveries(ctx context.Context, deliveries []domain.NotificationDelivery) error {
	if len(deliveries) == 0 {
		return nil
	}
	err := database.DB(ctx, r.db).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(deliveries, recipientInsertBatchSize).Error
	if err != nil {
		return fmt.Errorf("notifications.repository.CreateDeliveries: %w", err)
	}
	return nil
}

func (r *repository) ListDeliveriesByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.NotificationDelivery, error) {
	var items []domain.NotificationDelivery
	if err := database.DB(ctx, r.db).Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("notifications.repository.ListDeliveriesByIDs: %w", err)
	}
	return items, nil
}

func (r *repository) ListPendingDeliveries(ctx context.Context, notificationID uuid.UUID, channel domain.ConnectorType) ([]domain.NotificationDelivery, error) {
	var items []domain.NotificationDelivery
	err := database.DB(ctx, r.db).
		Where("notification_id = ? AND channel = ? AND status = ?", notificationID, channel, domain.DeliveryPending).
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("notifications.repository.ListPendingDeliveries: %w", err)
	}
	return items, nil
}

func (r *repository) MarkDeliveries(ctx context.Context, ids []uuid.UUID, status domain.NotificationDeliveryStatus, errMsg *string, sentAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	updates := map[string]any{"status": status, "error": errMsg}
	if status == domain.DeliverySent {
		updates["sent_at"] = sentAt
	}
	err := database.DB(ctx, r.db).Model(&domain.NotificationDelivery{}).
		Where("id IN ?", ids).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("notifications.repository.MarkDeliveries: %w", err)
	}
	return nil
}

func (r *repository) CountRecipients(ctx context.Context, notificationID uuid.UUID) (int64, error) {
	var n int64
	err := database.DB(ctx, r.db).Model(&domain.NotificationRecipient{}).
		Where("notification_id = ?", notificationID).Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("notifications.repository.CountRecipients: %w", err)
	}
	return n, nil
}

func (r *repository) DeliveryReport(ctx context.Context, notificationID uuid.UUID) ([]domain.NotificationChannelReport, error) {
	type row struct {
		Channel domain.ConnectorType
		Status  domain.NotificationDeliveryStatus
		N       int64
	}
	var rows []row
	err := database.DB(ctx, r.db).Model(&domain.NotificationDelivery{}).
		Select("channel, status, COUNT(*) as n").
		Where("notification_id = ?", notificationID).
		Group("channel, status").Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("notifications.repository.DeliveryReport: %w", err)
	}
	byChannel := map[domain.ConnectorType]*domain.NotificationChannelReport{}
	order := []domain.ConnectorType{}
	for _, rw := range rows {
		rep, ok := byChannel[rw.Channel]
		if !ok {
			rep = &domain.NotificationChannelReport{Channel: rw.Channel}
			byChannel[rw.Channel] = rep
			order = append(order, rw.Channel)
		}
		switch rw.Status {
		case domain.DeliveryPending:
			rep.Pending = rw.N
		case domain.DeliverySent:
			rep.Sent = rw.N
		case domain.DeliveryFailed:
			rep.Failed = rw.N
		}
	}
	out := make([]domain.NotificationChannelReport, 0, len(order))
	for _, ch := range order {
		out = append(out, *byChannel[ch])
	}
	return out, nil
}
