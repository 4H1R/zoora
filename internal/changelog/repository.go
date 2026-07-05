package changelog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.ChangelogRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, e *domain.ChangelogEntry) error {
	if err := database.DB(ctx, r.db).Create(e).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("changelog.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, e *domain.ChangelogEntry) error {
	// Full-row save; callers load-mutate-save so all columns are set.
	if err := database.DB(ctx, r.db).Save(e).Error; err != nil {
		return fmt.Errorf("changelog.repository.Update: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ChangelogEntry, error) {
	var e domain.ChangelogEntry
	if err := database.DB(ctx, r.db).First(&e, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("changelog.repository.FindByID: %w", err)
	}
	return &e, nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	res := database.DB(ctx, r.db).Delete(&domain.ChangelogEntry{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("changelog.repository.Delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) ListPublished(ctx context.Context, limit, offset int) ([]domain.ChangelogEntry, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.ChangelogEntry{}).
		Where("published_at IS NOT NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("changelog.repository.ListPublished count: %w", err)
	}
	var items []domain.ChangelogEntry
	if err := base.Order("published_at DESC, id DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("changelog.repository.ListPublished: %w", err)
	}
	return items, total, nil
}

func (r *repository) AdminList(ctx context.Context, limit, offset int) ([]domain.ChangelogEntry, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.ChangelogEntry{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("changelog.repository.AdminList count: %w", err)
	}
	var items []domain.ChangelogEntry
	if err := base.Order("created_at DESC, id DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("changelog.repository.AdminList: %w", err)
	}
	return items, total, nil
}

func (r *repository) LatestPublished(ctx context.Context) (*domain.ChangelogEntry, error) {
	var e domain.ChangelogEntry
	err := database.DB(ctx, r.db).
		Where("published_at IS NOT NULL").
		Order("published_at DESC, id DESC").First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // no published entries yet — not an error
	}
	if err != nil {
		return nil, fmt.Errorf("changelog.repository.LatestPublished: %w", err)
	}
	return &e, nil
}

func (r *repository) CountUnseen(ctx context.Context, since *time.Time) (int64, error) {
	q := database.DB(ctx, r.db).Model(&domain.ChangelogEntry{}).
		Where("published_at IS NOT NULL")
	if since != nil {
		q = q.Where("published_at > ?", *since)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return 0, fmt.Errorf("changelog.repository.CountUnseen: %w", err)
	}
	return n, nil
}

func (r *repository) LatestMajorUnseen(ctx context.Context, since *time.Time) (*domain.ChangelogEntry, error) {
	q := database.DB(ctx, r.db).
		Where("published_at IS NOT NULL AND is_major = true")
	if since != nil {
		q = q.Where("published_at > ?", *since)
	}
	var e domain.ChangelogEntry
	err := q.Order("published_at DESC, id DESC").First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("changelog.repository.LatestMajorUnseen: %w", err)
	}
	return &e, nil
}

func (r *repository) GetLastSeen(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	var u domain.User
	if err := database.DB(ctx, r.db).Select("changelog_last_seen_at").
		First(&u, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("changelog.repository.GetLastSeen: %w", err)
	}
	return u.ChangelogLastSeenAt, nil
}

func (r *repository) UpdateLastSeen(ctx context.Context, userID uuid.UUID, t time.Time) error {
	res := database.DB(ctx, r.db).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("changelog_last_seen_at", t)
	if res.Error != nil {
		return fmt.Errorf("changelog.repository.UpdateLastSeen: %w", res.Error)
	}
	return nil
}
