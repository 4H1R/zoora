package customfields

import (
	"context"
	"log/slog"
	"maps"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo   domain.CustomFieldRepository
	logger *slog.Logger
}

func NewService(repo domain.CustomFieldRepository, logger *slog.Logger) domain.CustomFieldService {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{repo: repo, logger: logger}
}

// callerOrg requires the manage permission (admin bypass) and returns the caller's org.
func (s *service) callerOrg(ctx context.Context) (uuid.UUID, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return uuid.Nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermCustomFieldsManage) {
		return uuid.Nil, domain.ErrForbidden
	}
	if caller.OrgID == nil {
		return uuid.Nil, domain.ErrForbidden
	}
	return *caller.OrgID, nil
}

func (s *service) CreateDefinition(ctx context.Context, dto domain.CreateCustomFieldDefinitionDTO) (*domain.UserCustomFieldDefinition, error) {
	orgID, err := s.callerOrg(ctx)
	if err != nil {
		return nil, err
	}
	if !dto.FieldType.Valid() {
		return nil, domain.NewValidationError(map[string]string{"field_type": "unsupported type"})
	}
	if dto.FieldType == domain.CustomFieldTypeSelect && len(dto.Options) == 0 {
		return nil, domain.NewValidationError(map[string]string{"options": "select fields need at least one option"})
	}

	count, err := s.repo.CountActiveDefinitions(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if count >= domain.MaxActiveCustomFieldsPerOrg {
		return nil, domain.ErrCustomFieldLimitReached
	}

	def := &domain.UserCustomFieldDefinition{
		OrganizationID: orgID,
		Label:          dto.Label,
		FieldType:      dto.FieldType,
		Options:        dto.Options,
		IsRequired:     dto.IsRequired,
		IsUnique:       dto.IsUnique,
		VisibleToUser:  dto.VisibleToUser,
		Position:       int(count),
		Description:    dto.Description,
	}
	if def.Options == nil {
		def.Options = []string{}
	}
	if err := s.repo.CreateDefinition(ctx, def); err != nil {
		return nil, err
	}
	return def, nil
}

func (s *service) ListDefinitions(ctx context.Context) ([]domain.UserCustomFieldDefinition, error) {
	orgID, err := s.callerOrg(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListDefinitions(ctx, orgID, false)
}

func (s *service) UpdateDefinition(ctx context.Context, id uuid.UUID, dto domain.UpdateCustomFieldDefinitionDTO) (*domain.UserCustomFieldDefinition, error) {
	orgID, err := s.callerOrg(ctx)
	if err != nil {
		return nil, err
	}
	def, err := s.repo.FindDefinitionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if def.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	if dto.Label != nil {
		def.Label = *dto.Label
	}
	if dto.IsRequired != nil {
		def.IsRequired = *dto.IsRequired
	}
	if dto.VisibleToUser != nil {
		def.VisibleToUser = *dto.VisibleToUser
	}
	if dto.Position != nil {
		def.Position = *dto.Position
	}
	if dto.Description != nil {
		def.Description = dto.Description
	}
	if dto.Options != nil {
		for _, opt := range removedOptions(def.Options, *dto.Options) {
			n, err := s.repo.CountUsersWithFieldValue(ctx, orgID, id, opt, uuid.Nil)
			if err != nil {
				return nil, err
			}
			if n > 0 {
				return nil, domain.ErrCustomFieldOptionInUse
			}
		}
		def.Options = *dto.Options
	}
	if dto.IsUnique != nil && *dto.IsUnique && !def.IsUnique {
		hasDup, err := s.repo.HasDuplicateFieldValues(ctx, orgID, id)
		if err != nil {
			return nil, err
		}
		if hasDup {
			return nil, domain.ErrCustomFieldDuplicateValue
		}
		def.IsUnique = true
	} else if dto.IsUnique != nil {
		def.IsUnique = *dto.IsUnique
	}

	if err := s.repo.UpdateDefinition(ctx, def); err != nil {
		return nil, err
	}
	return def, nil
}

func (s *service) ArchiveDefinition(ctx context.Context, id uuid.UUID) error {
	orgID, err := s.callerOrg(ctx)
	if err != nil {
		return err
	}
	def, err := s.repo.FindDefinitionByID(ctx, id)
	if err != nil {
		return err
	}
	if def.OrganizationID != orgID {
		return domain.ErrForbidden
	}
	if def.ArchivedAt != nil {
		return nil
	}
	now := timeNow()
	def.ArchivedAt = &now
	return s.repo.UpdateDefinition(ctx, def)
}

func (s *service) SetUserValues(ctx context.Context, userID uuid.UUID, dto domain.SetUserCustomFieldsDTO) (map[string]any, error) {
	orgID, err := s.callerOrg(ctx)
	if err != nil {
		return nil, err
	}

	current, userOrg, err := s.repo.GetUserCustomFields(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userOrg != orgID {
		return nil, domain.ErrForbidden
	}

	defs, err := s.repo.ListDefinitions(ctx, orgID, false)
	if err != nil {
		return nil, err
	}
	defByID := make(map[string]domain.UserCustomFieldDefinition, len(defs))
	for _, d := range defs {
		defByID[d.ID.String()] = d
	}

	fieldErrs := map[string]string{}
	merged := make(map[string]any, len(current))
	maps.Copy(merged, current)
	for key, val := range dto.Values {
		def, ok := defByID[key]
		if !ok {
			fieldErrs[key] = "unknown field"
			continue
		}
		if val == nil {
			delete(merged, key)
			continue
		}
		if err := domain.ValidateCustomFieldValue(def, val); err != nil {
			fieldErrs[key] = err.Error()
			continue
		}
		if def.IsUnique {
			text := domain.CustomFieldValueToText(val)
			n, err := s.repo.CountUsersWithFieldValue(ctx, orgID, def.ID, text, userID)
			if err != nil {
				return nil, err
			}
			if n > 0 {
				return nil, domain.ErrCustomFieldDuplicateValue
			}
		}
		merged[key] = val
	}
	if len(fieldErrs) > 0 {
		return nil, domain.NewValidationError(fieldErrs)
	}

	for _, d := range defs {
		if !d.IsRequired {
			continue
		}
		v, ok := merged[d.ID.String()]
		if !ok || isEmptyValue(v) {
			fieldErrs[d.ID.String()] = "required"
		}
	}
	if len(fieldErrs) > 0 {
		return nil, domain.NewValidationError(fieldErrs)
	}

	if err := s.repo.SetUserCustomFields(ctx, userID, merged); err != nil {
		return nil, err
	}
	return merged, nil
}

func (s *service) GetVisibleUserValues(ctx context.Context, userID uuid.UUID) ([]domain.VisibleCustomField, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	values, orgID, err := s.repo.GetUserCustomFields(ctx, userID)
	if err != nil {
		return nil, err
	}
	if orgID == uuid.Nil {
		return []domain.VisibleCustomField{}, nil
	}
	defs, err := s.repo.ListDefinitions(ctx, orgID, false)
	if err != nil {
		return nil, err
	}
	out := make([]domain.VisibleCustomField, 0, len(defs))
	for _, d := range defs {
		if !d.VisibleToUser {
			continue
		}
		out = append(out, domain.VisibleCustomField{
			FieldID:   d.ID,
			Label:     d.Label,
			FieldType: d.FieldType,
			Value:     values[d.ID.String()],
		})
	}
	return out, nil
}

func removedOptions(old, updated []string) []string {
	keep := make(map[string]struct{}, len(updated))
	for _, o := range updated {
		keep[o] = struct{}{}
	}
	var removed []string
	for _, o := range old {
		if _, ok := keep[o]; !ok {
			removed = append(removed, o)
		}
	}
	return removed
}

func isEmptyValue(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	return false
}
