package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// CustomFieldType enumerates the value types a manager can choose for a field.
type CustomFieldType string

const (
	CustomFieldTypeText    CustomFieldType = "text"
	CustomFieldTypeNumber  CustomFieldType = "number"
	CustomFieldTypeDate    CustomFieldType = "date"
	CustomFieldTypeBoolean CustomFieldType = "boolean"
	CustomFieldTypeSelect  CustomFieldType = "select"
)

// MaxActiveCustomFieldsPerOrg caps non-archived definitions per organization.
const MaxActiveCustomFieldsPerOrg = 10

func (t CustomFieldType) Valid() bool {
	switch t {
	case CustomFieldTypeText, CustomFieldTypeNumber, CustomFieldTypeDate,
		CustomFieldTypeBoolean, CustomFieldTypeSelect:
		return true
	default:
		return false
	}
}

// UserCustomFieldDefinition is a manager-defined, org-scoped profile field.
// User values are stored in users.custom_fields keyed by this row's ID.
type UserCustomFieldDefinition struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null;index" json:"organization_id"`
	Label          string          `gorm:"not null" json:"label"`
	FieldType      CustomFieldType `gorm:"column:field_type;not null" json:"field_type"`
	Options        []string        `gorm:"type:jsonb;not null;default:'[]';serializer:json" json:"options"`
	IsRequired     bool            `gorm:"not null;default:false" json:"is_required"`
	IsUnique       bool            `gorm:"not null;default:false" json:"is_unique"`
	VisibleToUser  bool            `gorm:"not null;default:false" json:"visible_to_user"`
	Position       int             `gorm:"not null;default:0" json:"position"`
	Description    *string         `json:"description,omitempty"`
	ArchivedAt     *time.Time      `gorm:"index" json:"archived_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (UserCustomFieldDefinition) TableName() string { return "user_custom_field_definitions" }

// VisibleCustomField is a hydrated {label, type, value} triple returned to a
// member for the fields their org marked visible_to_user. It exists because
// members lack custom_fields:manage and therefore cannot list definitions.
type VisibleCustomField struct {
	FieldID   uuid.UUID       `json:"field_id"`
	Label     string          `json:"label"`
	FieldType CustomFieldType `json:"field_type"`
	Value     any             `json:"value"`
}

// --- DTOs ---

type CreateCustomFieldDefinitionDTO struct {
	Label         string          `json:"label" binding:"required,min=1,max=255"`
	FieldType     CustomFieldType `json:"field_type" binding:"required"`
	Options       []string        `json:"options" binding:"omitempty,dive,min=1"`
	IsRequired    bool            `json:"is_required"`
	IsUnique      bool            `json:"is_unique"`
	VisibleToUser bool            `json:"visible_to_user"`
	Description   *string         `json:"description" binding:"omitempty,max=1000"`
}

// UpdateCustomFieldDefinitionDTO — note FieldType is intentionally absent (immutable).
type UpdateCustomFieldDefinitionDTO struct {
	Label         *string   `json:"label" binding:"omitempty,min=1,max=255"`
	Options       *[]string `json:"options" binding:"omitempty,dive,min=1"`
	IsRequired    *bool     `json:"is_required"`
	IsUnique      *bool     `json:"is_unique"`
	VisibleToUser *bool     `json:"visible_to_user"`
	Position      *int      `json:"position"`
	Description   *string   `json:"description" binding:"omitempty,max=1000"`
}

// SetUserCustomFieldsDTO is a partial merge: only listed keys change.
// A key mapped to JSON null deletes that value.
type SetUserCustomFieldsDTO struct {
	Values map[string]any `json:"values" binding:"required"`
}

// --- value validation ---

// ValidateCustomFieldValue checks a single value against its definition's type/options.
// value is a decoded JSON scalar (string, float64, bool). nil means "clear" and is caller-handled.
func ValidateCustomFieldValue(def UserCustomFieldDefinition, value any) error {
	switch def.FieldType {
	case CustomFieldTypeText:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%w: %q expects text", ErrValidation, def.Label)
		}
	case CustomFieldTypeSelect:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: %q expects a choice", ErrValidation, def.Label)
		}
		if slices.Contains(def.Options, s) {
			return nil
		}
		return fmt.Errorf("%w: %q is not a valid option for %q", ErrValidation, s, def.Label)
	case CustomFieldTypeNumber:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%w: %q expects a number", ErrValidation, def.Label)
		}
	case CustomFieldTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%w: %q expects true/false", ErrValidation, def.Label)
		}
	case CustomFieldTypeDate:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: %q expects a date", ErrValidation, def.Label)
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			if _, err2 := time.Parse(time.RFC3339, s); err2 != nil {
				return fmt.Errorf("%w: %q is not a valid date", ErrValidation, def.Label)
			}
		}
	default:
		return fmt.Errorf("%w: unknown field type %q", ErrValidation, def.FieldType)
	}
	return nil
}

// CustomFieldValueToText renders a value to its canonical text form for uniqueness scans,
// matching Postgres' `custom_fields ->> '<uuid>'` text extraction.
func CustomFieldValueToText(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// --- ports ---

type CustomFieldRepository interface {
	CreateDefinition(ctx context.Context, def *UserCustomFieldDefinition) error
	UpdateDefinition(ctx context.Context, def *UserCustomFieldDefinition) error
	FindDefinitionByID(ctx context.Context, id uuid.UUID) (*UserCustomFieldDefinition, error)
	ListDefinitions(ctx context.Context, orgID uuid.UUID, includeArchived bool) ([]UserCustomFieldDefinition, error)
	CountActiveDefinitions(ctx context.Context, orgID uuid.UUID) (int64, error)

	// GetUserCustomFields returns the raw value bag for a user (may be empty map)
	// plus the user's org (uuid.Nil if the user has none).
	GetUserCustomFields(ctx context.Context, userID uuid.UUID) (map[string]any, uuid.UUID, error)
	// SetUserCustomFields replaces the whole bag for a user.
	SetUserCustomFields(ctx context.Context, userID uuid.UUID, values map[string]any) error
	// CountUsersWithFieldValue counts non-deleted users in org whose value for fieldID
	// (text form) equals valueText, excluding excludeUserID. Powers uniqueness scan.
	CountUsersWithFieldValue(ctx context.Context, orgID, fieldID uuid.UUID, valueText string, excludeUserID uuid.UUID) (int64, error)
	// HasDuplicateFieldValues reports whether two+ non-deleted users in org share a
	// non-null value for fieldID. Powers the unique-toggle-on guard.
	HasDuplicateFieldValues(ctx context.Context, orgID, fieldID uuid.UUID) (bool, error)
}

type CustomFieldService interface {
	CreateDefinition(ctx context.Context, dto CreateCustomFieldDefinitionDTO) (*UserCustomFieldDefinition, error)
	UpdateDefinition(ctx context.Context, id uuid.UUID, dto UpdateCustomFieldDefinitionDTO) (*UserCustomFieldDefinition, error)
	ArchiveDefinition(ctx context.Context, id uuid.UUID) error
	ListDefinitions(ctx context.Context) ([]UserCustomFieldDefinition, error)
	SetUserValues(ctx context.Context, userID uuid.UUID, dto SetUserCustomFieldsDTO) (map[string]any, error)
	// GetVisibleUserValues returns the visible_to_user fields hydrated with the
	// user's values, for a member viewing their own profile (or a manager).
	GetVisibleUserValues(ctx context.Context, userID uuid.UUID) ([]VisibleCustomField, error)
}
