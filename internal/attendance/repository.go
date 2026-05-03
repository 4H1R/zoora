package attendance

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.AttendanceRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, a *domain.Attendance) error {
	if err := database.DB(ctx, r.db).Create(a).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("attendance.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Attendance, error) {
	var a domain.Attendance
	if err := database.DB(ctx, r.db).First(&a, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("attendance.repository.FindByID: %w", err)
	}
	return &a, nil
}

func (r *repository) Update(ctx context.Context, a *domain.Attendance) error {
	result := database.DB(ctx, r.db).Save(a)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("attendance.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Attendance{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("attendance.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) ListBySession(ctx context.Context, sessionID uuid.UUID, q domain.ListAttendanceQuery) ([]domain.Attendance, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Attendance{}).
		Where("class_session_id = ?", sessionID)
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	var items []domain.Attendance
	total, err := listparams.Paginate(base, q.ListParams, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("attendance.repository.ListBySession: %w", err)
	}
	return items, total, nil
}

func (r *repository) FindBySessionAndUser(ctx context.Context, sessionID, userID uuid.UUID) (*domain.Attendance, error) {
	var a domain.Attendance
	if err := database.DB(ctx, r.db).
		Where("class_session_id = ? AND user_id = ?", sessionID, userID).
		First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("attendance.repository.FindBySessionAndUser: %w", err)
	}
	return &a, nil
}
