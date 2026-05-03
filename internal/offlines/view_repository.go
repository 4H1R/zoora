package offlines

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type viewRepository struct {
	db *gorm.DB
}

func NewViewRepository(db *gorm.DB) domain.OfflineRoomViewRepository {
	return &viewRepository{db: db}
}

func (r *viewRepository) Create(ctx context.Context, v *domain.OfflineRoomView) error {
	if err := database.DB(ctx, r.db).Create(v).Error; err != nil {
		return fmt.Errorf("offlines.viewRepository.Create: %w", err)
	}
	return nil
}

func (r *viewRepository) ListByRoom(ctx context.Context, roomID uuid.UUID) ([]domain.OfflineRoomView, error) {
	var views []domain.OfflineRoomView
	if err := database.DB(ctx, r.db).
		Where("offline_room_id = ?", roomID).
		Order("viewed_at DESC").
		Find(&views).Error; err != nil {
		return nil, fmt.Errorf("offlines.viewRepository.ListByRoom: %w", err)
	}
	return views, nil
}

func (r *viewRepository) ListDistinctUsersByRoom(ctx context.Context, roomID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	if err := database.DB(ctx, r.db).
		Model(&domain.OfflineRoomView{}).
		Where("offline_room_id = ?", roomID).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, fmt.Errorf("offlines.viewRepository.ListDistinctUsersByRoom: %w", err)
	}
	return userIDs, nil
}
