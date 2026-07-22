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

func (r *repository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.Attendance{})
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
	if err := r.baseQuery(ctx).
		Preload("User").
		Preload("Class").
		Preload("ClassSession").
		First(&a, "id = ?", id).Error; err != nil {
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

// byUserScope applies the class/session scoping shared by ListByUser and
// SummarizeByUser so both always agree on the scoped set. The status filter
// is deliberately NOT applied here: the summary is a breakdown BY status, so
// it must ignore any active status filter.
func (r *repository) byUserScope(ctx context.Context, userID uuid.UUID, q domain.ListMyAttendanceQuery) *gorm.DB {
	base := r.baseQuery(ctx).Where("user_id = ?", userID)
	if q.ClassID != nil {
		base = base.Where("class_id = ?", *q.ClassID)
	}
	if q.ClassSessionID != nil {
		base = base.Where("class_session_id = ?", *q.ClassSessionID)
	}
	return base
}

func (r *repository) ListByUser(ctx context.Context, userID uuid.UUID, q domain.ListMyAttendanceQuery) ([]domain.Attendance, int64, error) {
	base := r.byUserScope(ctx, userID, q).
		Preload("Class").
		Preload("ClassSession")
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	var items []domain.Attendance
	total, err := listparams.Paginate(base, q.ListParams, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("attendance.repository.ListByUser: %w", err)
	}
	return items, total, nil
}

func (r *repository) SummarizeByUser(ctx context.Context, userID uuid.UUID, q domain.ListMyAttendanceQuery) (domain.MyAttendanceSummary, error) {
	var rows []struct {
		Status domain.AttendanceStatus
		Count  int
	}
	if err := r.byUserScope(ctx, userID, q).
		Select("status, COUNT(*) AS count").
		Group("status").
		Scan(&rows).Error; err != nil {
		return domain.MyAttendanceSummary{}, fmt.Errorf("attendance.repository.SummarizeByUser: %w", err)
	}
	var s domain.MyAttendanceSummary
	for _, row := range rows {
		switch row.Status {
		case domain.AttendanceStatusPresent:
			s.Present = row.Count
		case domain.AttendanceStatusAbsent:
			s.Absent = row.Count
		case domain.AttendanceStatusLate:
			s.Late = row.Count
		case domain.AttendanceStatusExcused:
			s.Excused = row.Count
		}
	}
	return s, nil
}

func (r *repository) ListBySession(ctx context.Context, sessionID uuid.UUID, q domain.ListAttendanceQuery) ([]domain.Attendance, int64, error) {
	base := r.baseQuery(ctx).Preload("User").Where("class_session_id = ?", sessionID)
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	if q.IsAutoMarked != nil {
		base = base.Where("is_auto_marked = ?", *q.IsAutoMarked)
	}
	var items []domain.Attendance
	total, err := listparams.Paginate(base, q.ListParams, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("attendance.repository.ListBySession: %w", err)
	}
	return items, total, nil
}

func (r *repository) AdminList(ctx context.Context, q domain.AdminListAttendanceQuery) ([]domain.Attendance, int64, error) {
	base := r.baseQuery(ctx).Preload("User").Preload("Class").Preload("ClassSession")
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	if q.IsAutoMarked != nil {
		base = base.Where("is_auto_marked = ?", *q.IsAutoMarked)
	}
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	if q.ClassID != nil {
		base = base.Where("class_id = ?", *q.ClassID)
	}
	if q.ClassSessionID != nil {
		base = base.Where("class_session_id = ?", *q.ClassSessionID)
	}
	if q.OrganizationID != nil {
		base = base.Where("organization_id = ?", *q.OrganizationID)
	}
	var items []domain.Attendance
	total, err := listparams.Paginate(base, q.ListParams, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("attendance.repository.AdminList: %w", err)
	}
	return items, total, nil
}

func (r *repository) FindBySessionAndUser(ctx context.Context, sessionID, userID uuid.UUID) (*domain.Attendance, error) {
	var a domain.Attendance
	if err := r.baseQuery(ctx).
		Where("class_session_id = ? AND user_id = ?", sessionID, userID).
		First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("attendance.repository.FindBySessionAndUser: %w", err)
	}
	return &a, nil
}

func (r *repository) ListByClassAndUsers(ctx context.Context, classID uuid.UUID, userIDs []uuid.UUID) ([]domain.Attendance, error) {
	if len(userIDs) == 0 {
		return []domain.Attendance{}, nil
	}
	var items []domain.Attendance
	if err := r.baseQuery(ctx).
		Where("class_id = ? AND user_id IN ?", classID, userIDs).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("attendance.repository.ListByClassAndUsers: %w", err)
	}
	return items, nil
}
