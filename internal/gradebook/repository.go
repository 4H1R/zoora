package gradebook

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type columnRepository struct {
	db *gorm.DB
}

func NewColumnRepository(db *gorm.DB) domain.GradebookColumnRepository {
	return &columnRepository{db: db}
}

func (r *columnRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.GradebookColumn{})
}

func (r *columnRepository) Create(ctx context.Context, col *domain.GradebookColumn) error {
	if err := database.DB(ctx, r.db).Create(col).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("gradebook.columnRepository.Create: %w", err)
	}
	return nil
}

func (r *columnRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.GradebookColumn, error) {
	var col domain.GradebookColumn
	if err := r.baseQuery(ctx).First(&col, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gradebook.columnRepository.FindByID: %w", err)
	}
	return &col, nil
}

func (r *columnRepository) Update(ctx context.Context, col *domain.GradebookColumn) error {
	result := database.DB(ctx, r.db).Save(col)
	if result.Error != nil {
		return fmt.Errorf("gradebook.columnRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *columnRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.GradebookColumn{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("gradebook.columnRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *columnRepository) ListByClass(ctx context.Context, classID uuid.UUID, q domain.ListGradebookColumnsQuery) ([]domain.GradebookColumn, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.GradebookColumn{}).Where("class_id = ?", classID)
	if q.Type != nil {
		base = base.Where("type = ?", *q.Type)
	}
	var cols []domain.GradebookColumn
	total, err := listparams.Paginate(base, q.ListParams, &cols)
	if err != nil {
		return nil, 0, fmt.Errorf("gradebook.columnRepository.ListByClass: %w", err)
	}
	return cols, total, nil
}

func (r *columnRepository) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.GradebookColumn, error) {
	var cols []domain.GradebookColumn
	if err := database.DB(ctx, r.db).
		Where("class_id = ?", classID).
		Order("order_index ASC, created_at ASC").
		Find(&cols).Error; err != nil {
		return nil, fmt.Errorf("gradebook.columnRepository.ListAllByClass: %w", err)
	}
	return cols, nil
}

// --- Cell Repository ---

type cellRepository struct {
	db *gorm.DB
}

func NewCellRepository(db *gorm.DB) domain.GradebookCellRepository {
	return &cellRepository{db: db}
}

func (r *cellRepository) Upsert(ctx context.Context, cell *domain.GradebookCell) error {
	err := database.DB(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "column_id"}, {Name: "student_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).
		Create(cell).Error
	if err != nil {
		return fmt.Errorf("gradebook.cellRepository.Upsert: %w", err)
	}
	return nil
}

func (r *cellRepository) ListByColumns(ctx context.Context, columnIDs []uuid.UUID) ([]domain.GradebookCell, error) {
	if len(columnIDs) == 0 {
		return nil, nil
	}
	var cells []domain.GradebookCell
	if err := database.DB(ctx, r.db).
		Where("column_id IN ?", columnIDs).
		Find(&cells).Error; err != nil {
		return nil, fmt.Errorf("gradebook.cellRepository.ListByColumns: %w", err)
	}
	return cells, nil
}
