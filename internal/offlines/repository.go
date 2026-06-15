package offlines

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

func NewRoomRepository(db *gorm.DB) domain.OfflineRoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.OfflineRoom{})
}

func (r *roomRepository) Create(ctx context.Context, room *domain.OfflineRoom) error {
	if err := database.DB(ctx, r.db).Create(room).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("offlines.roomRepository.Create: %w", err)
	}
	return nil
}

func (r *roomRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.OfflineRoom, error) {
	var room domain.OfflineRoom
	if err := r.baseQuery(ctx).
		Preload("Creator").
		Preload("Class").
		Preload("ClassSession").
		First(&room, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("offlines.roomRepository.FindByID: %w", err)
	}
	return &room, nil
}

func (r *roomRepository) Update(ctx context.Context, room *domain.OfflineRoom) error {
	result := database.DB(ctx, r.db).Save(room)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("offlines.roomRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.OfflineRoom{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("offlines.roomRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List applies a role-resolved scope produced by the service. OwnerID and
// MemberUserID are OR'd when both are set, so a teacher who is also enrolled
// in another class sees both sets in one query.
func (r *roomRepository) List(ctx context.Context, scope domain.OfflineRoomListScope, q domain.ListOfflineRoomsQuery) ([]domain.OfflineRoom, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.OfflineRoom{}).
		Preload("Creator").
		Preload("Class").
		Preload("ClassSession")
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if scope.OrganizationID != nil {
		base = base.Where("organization_id = ?", *scope.OrganizationID)
	}
	if !scope.All {
		switch {
		case scope.OwnerID != nil && scope.MemberUserID != nil:
			base = base.Where(
				"creator_id = ? OR class_id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.OwnerID, *scope.MemberUserID,
			)
		case scope.OwnerID != nil:
			base = base.Where("creator_id = ?", *scope.OwnerID)
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
	var rooms []domain.OfflineRoom
	total, err := listparams.Paginate(base, q.ListParams, &rooms)
	if err != nil {
		return nil, 0, fmt.Errorf("offlines.roomRepository.List: %w", err)
	}
	return rooms, total, nil
}

func (r *roomRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	result := r.baseQuery(ctx).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	if result.Error != nil {
		return fmt.Errorf("offlines.roomRepository.IncrementViewCount: %w", result.Error)
	}
	return nil
}

func (r *roomRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.OfflineRoom{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("offlines.roomRepository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.OfflineRoom, error) {
	var room domain.OfflineRoom
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.OfflineRoom{}).First(&room, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("offlines.roomRepository.FindByIDIncludingDeleted: %w", err)
	}
	return &room, nil
}

func (r *roomRepository) AdminList(ctx context.Context, q domain.AdminListOfflineRoomsQuery) ([]domain.OfflineRoom, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.OfflineRoom{}).
		Preload("Creator").
		Preload("Class").
		Preload("ClassSession")
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.ClassID != nil {
		base = base.Where("class_id = ?", *q.ClassID)
	}
	if q.ClassSessionID != nil {
		base = base.Where("class_session_id = ?", *q.ClassSessionID)
	}
	if q.CreatorID != nil {
		base = base.Where("creator_id = ?", *q.CreatorID)
	}
	var rooms []domain.OfflineRoom
	total, err := listparams.Paginate(base, q.ListParams, &rooms)
	if err != nil {
		return nil, 0, fmt.Errorf("offlines.roomRepository.AdminList: %w", err)
	}
	return rooms, total, nil
}
