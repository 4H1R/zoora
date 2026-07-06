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
