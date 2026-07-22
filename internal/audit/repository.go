package audit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.AuditRepository {
	return &repository{db: db}
}

// Create inserts an entry. Uses database.DB(ctx, ...) so it joins the caller's
// transaction when one is present in ctx (the same-tx success path); falls back
// to the base handle for denied/best-effort writes with no tx.
func (r *repository) Create(ctx context.Context, e *domain.AuditEntry) error {
	if err := database.DB(ctx, r.db).Create(e).Error; err != nil {
		return fmt.Errorf("audit.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) List(ctx context.Context, orgID uuid.UUID, q domain.AuditListQuery) ([]domain.AuditEntry, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.AuditEntry{}).Where("organization_id = ?", orgID)
	if q.ActorID != nil {
		base = base.Where("actor_id = ?", *q.ActorID)
	}
	if q.Action != nil {
		base = base.Where("action = ?", *q.Action)
	}
	if q.TargetType != nil {
		base = base.Where("target_type = ?", *q.TargetType)
	}
	if q.TargetID != nil {
		base = base.Where("target_id = ?", *q.TargetID)
	}
	if q.Outcome != nil {
		base = base.Where("outcome = ?", *q.Outcome)
	}
	if q.From != nil {
		base = base.Where("created_at >= ?", *q.From)
	}
	if q.To != nil {
		base = base.Where("created_at <= ?", *q.To)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("audit.repository.List count: %w", err)
	}

	var entries []domain.AuditEntry
	err := base.
		Order("created_at DESC").
		Limit(q.ListParams.Limit()).
		Offset(q.ListParams.Offset()).
		Find(&entries).Error
	if err != nil {
		return nil, 0, fmt.Errorf("audit.repository.List: %w", err)
	}
	return entries, total, nil
}
