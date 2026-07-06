package connectors

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.UserConnectorRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, c *domain.UserConnector) error {
	// Upsert on (user_id, type, target): re-linking the same endpoint just
	// re-verifies and re-enables it.
	err := database.DB(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "type"}, {Name: "target"}},
		DoUpdates: clause.AssignmentColumns([]string{"verified_at", "enabled", "updated_at"}),
	}).Create(c).Error
	if err != nil {
		return fmt.Errorf("connectors.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.UserConnector, error) {
	var c domain.UserConnector
	if err := database.DB(ctx, r.db).First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("connectors.repository.FindByID: %w", err)
	}
	return &c, nil
}

func (r *repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserConnector, error) {
	var items []domain.UserConnector
	if err := database.DB(ctx, r.db).
		Where("user_id = ?", userID).
		Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("connectors.repository.ListByUser: %w", err)
	}
	return items, nil
}

func (r *repository) ListVerifiedEnabledByUsers(ctx context.Context, userIDs []uuid.UUID) ([]domain.UserConnector, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var items []domain.UserConnector
	if err := database.DB(ctx, r.db).
		Where("user_id IN ? AND verified_at IS NOT NULL AND enabled = true", userIDs).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("connectors.repository.ListVerifiedEnabledByUsers: %w", err)
	}
	return items, nil
}

func (r *repository) Update(ctx context.Context, c *domain.UserConnector) error {
	if err := database.DB(ctx, r.db).Save(c).Error; err != nil {
		return fmt.Errorf("connectors.repository.Update: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	res := database.DB(ctx, r.db).Delete(&domain.UserConnector{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("connectors.repository.Delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) DeleteByTypeTarget(ctx context.Context, t domain.ConnectorType, target string) error {
	err := database.DB(ctx, r.db).
		Delete(&domain.UserConnector{}, "type = ? AND target = ?", t, target).Error
	if err != nil {
		return fmt.Errorf("connectors.repository.DeleteByTypeTarget: %w", err)
	}
	return nil
}
