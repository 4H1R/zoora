package customfields

import (
	"context"
	"encoding/json"
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

func NewRepository(db *gorm.DB) domain.CustomFieldRepository {
	return &repository{db: db}
}

func (r *repository) defQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.UserCustomFieldDefinition{})
}

func (r *repository) CreateDefinition(ctx context.Context, def *domain.UserCustomFieldDefinition) error {
	if err := database.DB(ctx, r.db).Create(def).Error; err != nil {
		return fmt.Errorf("customfields.repository.CreateDefinition: %w", err)
	}
	return nil
}

func (r *repository) UpdateDefinition(ctx context.Context, def *domain.UserCustomFieldDefinition) error {
	res := database.DB(ctx, r.db).Save(def)
	if res.Error != nil {
		return fmt.Errorf("customfields.repository.UpdateDefinition: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) FindDefinitionByID(ctx context.Context, id uuid.UUID) (*domain.UserCustomFieldDefinition, error) {
	var def domain.UserCustomFieldDefinition
	if err := r.defQuery(ctx).First(&def, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("customfields.repository.FindDefinitionByID: %w", err)
	}
	return &def, nil
}

func (r *repository) ListDefinitions(ctx context.Context, orgID uuid.UUID, includeArchived bool) ([]domain.UserCustomFieldDefinition, error) {
	q := r.defQuery(ctx).Where("organization_id = ?", orgID)
	if !includeArchived {
		q = q.Where("archived_at IS NULL")
	}
	var out []domain.UserCustomFieldDefinition
	if err := q.Order("position ASC, created_at ASC").Find(&out).Error; err != nil {
		return nil, fmt.Errorf("customfields.repository.ListDefinitions: %w", err)
	}
	return out, nil
}

func (r *repository) CountActiveDefinitions(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var n int64
	if err := r.defQuery(ctx).
		Where("organization_id = ? AND archived_at IS NULL", orgID).
		Count(&n).Error; err != nil {
		return 0, fmt.Errorf("customfields.repository.CountActiveDefinitions: %w", err)
	}
	return n, nil
}

func (r *repository) GetUserCustomFields(ctx context.Context, userID uuid.UUID) (map[string]any, uuid.UUID, error) {
	var user domain.User
	if err := database.DB(ctx, r.db).
		Select("id", "organization_id", "custom_fields").
		First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, uuid.Nil, domain.ErrNotFound
		}
		return nil, uuid.Nil, fmt.Errorf("customfields.repository.GetUserCustomFields: %w", err)
	}
	values := map[string]any{}
	if len(user.CustomFields) > 0 {
		if err := json.Unmarshal(user.CustomFields, &values); err != nil {
			return nil, uuid.Nil, fmt.Errorf("customfields.repository.GetUserCustomFields decode: %w", err)
		}
	}
	orgID := uuid.Nil
	if user.OrganizationID != nil {
		orgID = *user.OrganizationID
	}
	return values, orgID, nil
}

func (r *repository) SetUserCustomFields(ctx context.Context, userID uuid.UUID, values map[string]any) error {
	raw, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("customfields.repository.SetUserCustomFields encode: %w", err)
	}
	res := database.DB(ctx, r.db).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("custom_fields", json.RawMessage(raw))
	if res.Error != nil {
		return fmt.Errorf("customfields.repository.SetUserCustomFields: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) CountUsersWithFieldValue(ctx context.Context, orgID, fieldID uuid.UUID, valueText string, excludeUserID uuid.UUID) (int64, error) {
	q := database.DB(ctx, r.db).Model(&domain.User{}).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Where("custom_fields ->> ? = ?", fieldID.String(), valueText)
	if excludeUserID != uuid.Nil {
		q = q.Where("id <> ?", excludeUserID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return 0, fmt.Errorf("customfields.repository.CountUsersWithFieldValue: %w", err)
	}
	return n, nil
}

// HasDuplicateFieldValues uses raw SQL because GORM's `?` placeholder collides
// with the jsonb key-exists operator; jsonb_exists avoids the ambiguity.
func (r *repository) HasDuplicateFieldValues(ctx context.Context, orgID, fieldID uuid.UUID) (bool, error) {
	var n int64
	err := database.DB(ctx, r.db).
		Raw(`SELECT COUNT(*) FROM (
			SELECT custom_fields ->> ? AS v
			FROM users
			WHERE organization_id = ? AND deleted_at IS NULL
			  AND jsonb_exists(custom_fields, ?)
			GROUP BY 1
			HAVING COUNT(*) > 1
		) dups`, fieldID.String(), orgID, fieldID.String()).
		Scan(&n).Error
	if err != nil {
		return false, fmt.Errorf("customfields.repository.HasDuplicateFieldValues: %w", err)
	}
	return n > 0, nil
}
