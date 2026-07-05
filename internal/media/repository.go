package media

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

func NewRepository(db *gorm.DB) domain.MediaRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, m *domain.Media) error {
	if err := database.DB(ctx, r.db).Create(m).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("media.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	var m domain.Media
	if err := database.DB(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("media.repository.FindByID: %w", err)
	}
	return &m, nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Media{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("media.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) ListFolders(ctx context.Context, orgID uuid.UUID) ([]domain.MediaFolder, error) {
	var folders []domain.MediaFolder
	err := database.DB(ctx, r.db).
		Model(&domain.Media{}).
		Select("model_type, COUNT(*) AS file_count, COALESCE(SUM(size), 0) AS total_size").
		Where("organization_id = ?", orgID).
		Group("model_type").
		Order("model_type ASC").
		Scan(&folders).Error
	if err != nil {
		return nil, fmt.Errorf("media.repository.ListFolders: %w", err)
	}
	return folders, nil
}

func (r *repository) ListFiles(ctx context.Context, orgID uuid.UUID, modelType string, p domain.ListParams) ([]domain.Media, int64, error) {
	base := database.DB(ctx, r.db).
		Model(&domain.Media{}).
		Where("organization_id = ? AND model_type = ?", orgID, modelType)
	var items []domain.Media
	total, err := listparams.Paginate(base, p, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("media.repository.ListFiles: %w", err)
	}
	return items, total, nil
}

func (r *repository) ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]domain.Media, error) {
	q := database.DB(ctx, r.db).
		Where("model_type = ? AND model_id = ?", modelType, modelID).
		Order("order_column ASC, created_at ASC")
	if collection != "" {
		q = q.Where("collection_name = ?", collection)
	}
	var items []domain.Media
	if err := q.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("media.repository.ListByModel: %w", err)
	}
	return items, nil
}
