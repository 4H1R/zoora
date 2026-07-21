package customfields_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/customfields"
	"github.com/4H1R/zoora/internal/domain"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) CreateDefinition(ctx context.Context, d *domain.UserCustomFieldDefinition) error {
	return m.Called(ctx, d).Error(0)
}

func (m *mockRepo) UpdateDefinition(ctx context.Context, d *domain.UserCustomFieldDefinition) error {
	return m.Called(ctx, d).Error(0)
}

func (m *mockRepo) FindDefinitionByID(ctx context.Context, id uuid.UUID) (*domain.UserCustomFieldDefinition, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.UserCustomFieldDefinition), a.Error(1)
}

func (m *mockRepo) ListDefinitions(ctx context.Context, orgID uuid.UUID, inclArch bool) ([]domain.UserCustomFieldDefinition, error) {
	a := m.Called(ctx, orgID, inclArch)
	return a.Get(0).([]domain.UserCustomFieldDefinition), a.Error(1)
}

func (m *mockRepo) CountActiveDefinitions(ctx context.Context, orgID uuid.UUID) (int64, error) {
	a := m.Called(ctx, orgID)
	return a.Get(0).(int64), a.Error(1)
}

func (m *mockRepo) GetUserCustomFields(ctx context.Context, userID uuid.UUID) (map[string]any, uuid.UUID, error) {
	a := m.Called(ctx, userID)
	return a.Get(0).(map[string]any), a.Get(1).(uuid.UUID), a.Error(2)
}

func (m *mockRepo) SetUserCustomFields(ctx context.Context, userID uuid.UUID, v map[string]any) error {
	return m.Called(ctx, userID, v).Error(0)
}

func (m *mockRepo) CountUsersWithFieldValue(ctx context.Context, orgID, fieldID uuid.UUID, vt string, excl uuid.UUID) (int64, error) {
	a := m.Called(ctx, orgID, fieldID, vt, excl)
	return a.Get(0).(int64), a.Error(1)
}

func (m *mockRepo) HasDuplicateFieldValues(ctx context.Context, orgID, fieldID uuid.UUID) (bool, error) {
	a := m.Called(ctx, orgID, fieldID)
	return a.Get(0).(bool), a.Error(1)
}

func managerCtx(orgID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgID,
		Permissions: []string{string(domain.PermCustomFieldsManage), string(domain.PermUsersUpdate)},
	})
}

func TestCreateDefinitionEnforcesCap(t *testing.T) {
	orgID := uuid.New()
	repo := &mockRepo{}
	repo.On("CountActiveDefinitions", mock.Anything, orgID).Return(int64(domain.MaxActiveCustomFieldsPerOrg), nil)
	svc := customfields.NewService(repo, nil)

	_, err := svc.CreateDefinition(managerCtx(orgID), domain.CreateCustomFieldDefinitionDTO{
		Label: "X", FieldType: domain.CustomFieldTypeText,
	})
	require.ErrorIs(t, err, domain.ErrCustomFieldLimitReached)
}

func TestCreateDefinitionRequiresPermission(t *testing.T) {
	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgID})
	svc := customfields.NewService(&mockRepo{}, nil)

	_, err := svc.CreateDefinition(ctx, domain.CreateCustomFieldDefinitionDTO{Label: "X", FieldType: domain.CustomFieldTypeText})
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdateDefinitionRejectsUniqueToggleWithDuplicates(t *testing.T) {
	orgID := uuid.New()
	fieldID := uuid.New()
	existing := &domain.UserCustomFieldDefinition{ID: fieldID, OrganizationID: orgID, FieldType: domain.CustomFieldTypeText}
	repo := &mockRepo{}
	repo.On("FindDefinitionByID", mock.Anything, fieldID).Return(existing, nil)
	repo.On("HasDuplicateFieldValues", mock.Anything, orgID, fieldID).Return(true, nil)
	svc := customfields.NewService(repo, nil)

	trueVal := true
	_, err := svc.UpdateDefinition(managerCtx(orgID), fieldID, domain.UpdateCustomFieldDefinitionDTO{IsUnique: &trueVal})
	require.ErrorIs(t, err, domain.ErrCustomFieldDuplicateValue)
}

func TestSetUserValuesValidatesAndMerges(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	textField := uuid.New()
	repo := &mockRepo{}
	defs := []domain.UserCustomFieldDefinition{
		{ID: textField, OrganizationID: orgID, FieldType: domain.CustomFieldTypeText, Label: "Student ID", IsUnique: true},
	}
	repo.On("GetUserCustomFields", mock.Anything, userID).Return(map[string]any{}, orgID, nil)
	repo.On("ListDefinitions", mock.Anything, orgID, false).Return(defs, nil)
	repo.On("CountUsersWithFieldValue", mock.Anything, orgID, textField, "12345", userID).Return(int64(0), nil)
	repo.On("SetUserCustomFields", mock.Anything, userID, mock.Anything).Return(nil)
	svc := customfields.NewService(repo, nil)

	out, err := svc.SetUserValues(managerCtx(orgID), userID, domain.SetUserCustomFieldsDTO{
		Values: map[string]any{textField.String(): "12345"},
	})
	require.NoError(t, err)
	require.Equal(t, "12345", out[textField.String()])
	repo.AssertCalled(t, "SetUserCustomFields", mock.Anything, userID, mock.Anything)
}

func TestSetUserValuesRejectsDuplicateUnique(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	textField := uuid.New()
	repo := &mockRepo{}
	defs := []domain.UserCustomFieldDefinition{
		{ID: textField, OrganizationID: orgID, FieldType: domain.CustomFieldTypeText, Label: "Student ID", IsUnique: true},
	}
	repo.On("GetUserCustomFields", mock.Anything, userID).Return(map[string]any{}, orgID, nil)
	repo.On("ListDefinitions", mock.Anything, orgID, false).Return(defs, nil)
	repo.On("CountUsersWithFieldValue", mock.Anything, orgID, textField, "12345", userID).Return(int64(1), nil)
	svc := customfields.NewService(repo, nil)

	_, err := svc.SetUserValues(managerCtx(orgID), userID, domain.SetUserCustomFieldsDTO{
		Values: map[string]any{textField.String(): "12345"},
	})
	require.ErrorIs(t, err, domain.ErrCustomFieldDuplicateValue)
}

func TestSetUserValuesRejectsUnknownFieldKey(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{}
	repo.On("GetUserCustomFields", mock.Anything, userID).Return(map[string]any{}, orgID, nil)
	repo.On("ListDefinitions", mock.Anything, orgID, false).Return([]domain.UserCustomFieldDefinition{}, nil)
	svc := customfields.NewService(repo, nil)

	_, err := svc.SetUserValues(managerCtx(orgID), userID, domain.SetUserCustomFieldsDTO{
		Values: map[string]any{uuid.New().String(): "x"},
	})
	require.Error(t, err)
}

func TestSetUserValuesRejectsCrossOrgUser(t *testing.T) {
	orgID := uuid.New()
	otherOrg := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{}
	repo.On("GetUserCustomFields", mock.Anything, userID).Return(map[string]any{}, otherOrg, nil)
	svc := customfields.NewService(repo, nil)

	_, err := svc.SetUserValues(managerCtx(orgID), userID, domain.SetUserCustomFieldsDTO{
		Values: map[string]any{uuid.New().String(): "x"},
	})
	require.ErrorIs(t, err, domain.ErrForbidden)
}
