package entitlements

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
)

type Repository interface {
	GetOrgPlan(ctx context.Context, orgID uuid.UUID) (domain.Plan, *time.Time, error)
	CountUsers(ctx context.Context, orgID uuid.UUID) (int64, error)
	SumStorageBytes(ctx context.Context, orgID uuid.UUID) (int64, error)
	CountActiveLiveRooms(ctx context.Context, orgID uuid.UUID) (int64, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) GetOrgPlan(ctx context.Context, orgID uuid.UUID) (domain.Plan, *time.Time, error) {
	var row struct {
		Plan          domain.Plan
		PlanExpiresAt *time.Time
	}
	err := r.db.WithContext(ctx).Table("organizations").
		Select("plan, plan_expires_at").
		Where("id = ? AND deleted_at IS NULL", orgID).
		Take(&row).Error
	if err != nil {
		return "", nil, err
	}
	return row.Plan, row.PlanExpiresAt, nil
}

func (r *repository) CountUsers(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("users").
		Where("organization_id = ? AND deleted_at IS NULL", orgID).Count(&n).Error
	return n, err
}

// SumStorageBytes = media bytes + recording bytes. Recordings live in
// live_recordings (separate table, NOT media rows) so both must be summed for
// the storage quota to see recording usage. media has no soft delete.
func (r *repository) SumStorageBytes(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var mediaSum int64
	if err := r.db.WithContext(ctx).Table("media").
		Where("organization_id = ?", orgID).
		Select("COALESCE(SUM(size), 0)").Scan(&mediaSum).Error; err != nil {
		return 0, err
	}
	// live_recordings has no organization_id — reach the org through the room's
	// class: live_recordings -> live_rooms -> class_sessions -> classes.
	var recSum int64
	if err := r.db.WithContext(ctx).Table("live_recordings").
		Joins("JOIN live_rooms ON live_rooms.id = live_recordings.live_room_id").
		Joins("JOIN class_sessions ON class_sessions.id = live_rooms.class_session_id").
		Joins("JOIN classes ON classes.id = class_sessions.class_id").
		Where("classes.organization_id = ?", orgID).
		Select("COALESCE(SUM(live_recordings.size), 0)").Scan(&recSum).Error; err != nil {
		return 0, err
	}
	return mediaSum + recSum, nil
}

func (r *repository) CountActiveLiveRooms(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("live_rooms").
		Joins("JOIN class_sessions ON class_sessions.id = live_rooms.class_session_id").
		Joins("JOIN classes ON classes.id = class_sessions.class_id").
		Where("classes.organization_id = ? AND live_rooms.status = ? AND live_rooms.deleted_at IS NULL",
			orgID, domain.LiveRoomStatusActive).
		Count(&n).Error
	return n, err
}
