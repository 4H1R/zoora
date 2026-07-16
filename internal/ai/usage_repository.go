package ai

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
)

// UsageRepository persists AI metering events. Implements domain.AIUsageRecorder.
type UsageRepository struct {
	db *gorm.DB
}

// NewUsageRepository builds the metering recorder.
func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

// Record inserts one usage event.
func (r *UsageRepository) Record(ctx context.Context, ev domain.AIUsageEvent) error {
	if err := r.db.WithContext(ctx).Create(&ev).Error; err != nil {
		return fmt.Errorf("recording ai usage: %w", err)
	}
	return nil
}
