package practices

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

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) domain.PracticeRoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.PracticeRoom{})
}

func (r *roomRepository) Create(ctx context.Context, room *domain.PracticeRoom) error {
	if err := database.DB(ctx, r.db).Create(room).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("practices.roomRepository.Create: %w", err)
	}
	return nil
}

func (r *roomRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.PracticeRoom, error) {
	var room domain.PracticeRoom
	if err := r.baseQuery(ctx).First(&room, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("practices.roomRepository.FindByID: %w", err)
	}
	return &room, nil
}

func (r *roomRepository) Update(ctx context.Context, room *domain.PracticeRoom) error {
	result := database.DB(ctx, r.db).Save(room)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("practices.roomRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.PracticeRoom{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("practices.roomRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) List(ctx context.Context, scope domain.PracticeRoomListScope, q domain.ListPracticeRoomsQuery) ([]domain.PracticeRoom, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.PracticeRoom{})
	if scope.IncludeDeleted {
		base = base.Unscoped()
	}
	if !scope.All {
		switch {
		case scope.OwnerID != nil && scope.MemberUserID != nil:
			base = base.Where(
				"user_id = ? OR class_id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.OwnerID, *scope.MemberUserID,
			)
		case scope.OwnerID != nil:
			base = base.Where("user_id = ?", *scope.OwnerID)
		case scope.MemberUserID != nil:
			base = base.Where(
				"class_id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.MemberUserID,
			)
		}
	}
	if q.ClassID != nil {
		base = base.Where("class_id = ?", *q.ClassID)
	}
	if q.ClassSessionID != nil {
		base = base.Where("class_session_id = ?", *q.ClassSessionID)
	}
	var rooms []domain.PracticeRoom
	total, err := listparams.Paginate(base, q.ListParams, &rooms)
	if err != nil {
		return nil, 0, fmt.Errorf("practices.roomRepository.List: %w", err)
	}
	return rooms, total, nil
}

func (r *roomRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.PracticeRoom{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("practices.roomRepository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.PracticeRoom, error) {
	var room domain.PracticeRoom
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.PracticeRoom{}).First(&room, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("practices.roomRepository.FindByIDIncludingDeleted: %w", err)
	}
	return &room, nil
}

func (r *roomRepository) AdminList(ctx context.Context, q domain.AdminListPracticeRoomsQuery) ([]domain.PracticeRoom, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.PracticeRoom{})
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.ClassID != nil {
		base = base.Where("class_id = ?", *q.ClassID)
	}
	if q.ClassSessionID != nil {
		base = base.Where("class_session_id = ?", *q.ClassSessionID)
	}
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	var rooms []domain.PracticeRoom
	total, err := listparams.Paginate(base, q.ListParams, &rooms)
	if err != nil {
		return nil, 0, fmt.Errorf("practices.roomRepository.AdminList: %w", err)
	}
	return rooms, total, nil
}

// --- Submission Repository ---

type submissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) domain.PracticeSubmissionRepository {
	return &submissionRepository{db: db}
}

func (r *submissionRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.PracticeSubmission{})
}

func (r *submissionRepository) Create(ctx context.Context, sub *domain.PracticeSubmission) error {
	if err := database.DB(ctx, r.db).Create(sub).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("practices.submissionRepository.Create: %w", err)
	}
	return nil
}

func (r *submissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.PracticeSubmission, error) {
	var sub domain.PracticeSubmission
	if err := r.baseQuery(ctx).First(&sub, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("practices.submissionRepository.FindByID: %w", err)
	}
	return &sub, nil
}

func (r *submissionRepository) Update(ctx context.Context, sub *domain.PracticeSubmission) error {
	result := database.DB(ctx, r.db).Save(sub)
	if result.Error != nil {
		return fmt.Errorf("practices.submissionRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *submissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.PracticeSubmission{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("practices.submissionRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *submissionRepository) FindByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*domain.PracticeSubmission, error) {
	var sub domain.PracticeSubmission
	err := r.baseQuery(ctx).
		Where("practice_room_id = ? AND user_id = ?", roomID, userID).
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("practices.submissionRepository.FindByRoomAndUser: %w", err)
	}
	return &sub, nil
}

func (r *submissionRepository) ListByRoom(ctx context.Context, roomID uuid.UUID, p domain.ListParams) ([]domain.PracticeSubmission, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.PracticeSubmission{}).
		Where("practice_room_id = ?", roomID)
	var subs []domain.PracticeSubmission
	total, err := listparams.Paginate(base, p, &subs)
	if err != nil {
		return nil, 0, fmt.Errorf("practices.submissionRepository.ListByRoom: %w", err)
	}
	return subs, total, nil
}
