package tutorials

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.TutorialRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, tu *domain.Tutorial) error {
	if err := database.DB(ctx, r.db).Create(tu).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("tutorials.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, tu *domain.Tutorial) error {
	// Full-row save; callers load-mutate-save so all columns are set.
	if err := database.DB(ctx, r.db).Save(tu).Error; err != nil {
		return fmt.Errorf("tutorials.repository.Update: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tutorial, error) {
	var tu domain.Tutorial
	if err := database.DB(ctx, r.db).First(&tu, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("tutorials.repository.FindByID: %w", err)
	}
	return &tu, nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	res := database.DB(ctx, r.db).Delete(&domain.Tutorial{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("tutorials.repository.Delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) ListPublished(ctx context.Context) ([]domain.Tutorial, error) {
	var items []domain.Tutorial
	if err := database.DB(ctx, r.db).
		Where("published_at IS NOT NULL").
		Order("position ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("tutorials.repository.ListPublished: %w", err)
	}
	return items, nil
}

func (r *repository) AdminList(ctx context.Context) ([]domain.Tutorial, error) {
	var items []domain.Tutorial
	if err := database.DB(ctx, r.db).
		Order("position ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("tutorials.repository.AdminList: %w", err)
	}
	return items, nil
}

func (r *repository) MaxPosition(ctx context.Context) (int, error) {
	// COALESCE keeps an empty table at 0 rather than a NULL scan error.
	var max int
	if err := database.DB(ctx, r.db).Model(&domain.Tutorial{}).
		Select("COALESCE(MAX(position), 0)").Scan(&max).Error; err != nil {
		return 0, fmt.Errorf("tutorials.repository.MaxPosition: %w", err)
	}
	return max, nil
}

func (r *repository) Reorder(ctx context.Context, ids []uuid.UUID) error {
	err := database.DB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&domain.Tutorial{}).
				Where("id = ?", id).
				Update("position", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("tutorials.repository.Reorder: %w", err)
	}
	return nil
}
