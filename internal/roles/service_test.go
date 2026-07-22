package roles_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/roles"
)

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}

func (m *mockRoleRepo) FindPresetByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *mockRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *mockRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}

func (m *mockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockRoleRepo) List(ctx context.Context, f domain.RoleFilter) ([]domain.Role, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.Role), args.Error(1)
}

func (m *mockRoleRepo) AdminList(ctx context.Context, f domain.AdminRoleFilter) ([]domain.Role, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.Role), args.Get(1).(int64), args.Error(2)
}

func (m *mockRoleRepo) SetPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.Called(ctx, roleID, permissionIDs).Error(0)
}

func (m *mockRoleRepo) Stats(ctx context.Context, orgID *uuid.UUID) (*domain.RoleStats, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RoleStats), args.Error(1)
}

func (m *mockRoleRepo) GetPermissionNames(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]string), args.Error(1)
}

type mockPermRepo struct{ mock.Mock }

func (m *mockPermRepo) List(ctx context.Context) ([]domain.Permission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *mockPermRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Permission, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]domain.Permission), args.Error(1)
}

type noopTx struct{}

func (noopTx) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// auditSpy captures the records a service emits so tests can assert on them.
type auditSpy struct{ records []domain.AuditRecord }

func (a *auditSpy) Record(_ context.Context, r domain.AuditRecord) error {
	a.records = append(a.records, r)
	return nil
}

func (a *auditSpy) RecordDenied(_ context.Context, _ domain.AuditRecord) error { return nil }

func TestCreateRole(t *testing.T) {
	logger := slog.Default()
	roleRepo := &mockRoleRepo{}
	permRepo := &mockPermRepo{}

	permID := uuid.New()

	caller := domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"users:view"},
		Ent:         domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
	}
	ctx := domain.WithCaller(context.Background(), caller)

	perms := []domain.Permission{{ID: permID, Name: "users:view"}}
	permRepo.On("FindByIDs", ctx, []uuid.UUID{permID}).Return(perms, nil)
	roleRepo.On("Create", ctx, mock.AnythingOfType("*domain.Role")).Return(nil)
	roleRepo.On("SetPermissions", ctx, mock.AnythingOfType("uuid.UUID"), []uuid.UUID{permID}).Return(nil)

	svc := roles.NewService(roleRepo, permRepo, noopTx{}, &auditSpy{}, nil, logger)

	orgID := uuid.New()
	role, err := svc.Create(ctx, domain.CreateRoleDTO{
		OrganizationID: &orgID,
		Name:           "Viewer",
		PermissionIDs:  []uuid.UUID{permID},
	})

	assert.NoError(t, err)
	assert.Equal(t, "Viewer", role.Name)
}

func TestCreateRole_FreePlanRejectsCustomRole(t *testing.T) {
	svc := roles.NewService(&mockRoleRepo{}, &mockPermRepo{}, noopTx{}, &auditSpy{}, nil, slog.Default())
	caller := domain.Caller{UserID: uuid.New(), Ent: domain.PlanCatalog[domain.PlanFree]}
	ctx := domain.WithCaller(context.Background(), caller)

	orgID := uuid.New()
	_, err := svc.Create(ctx, domain.CreateRoleDTO{
		OrganizationID: &orgID,
		Name:           "Viewer",
		PermissionIDs:  []uuid.UUID{uuid.New()},
	})
	assert.ErrorIs(t, err, domain.ErrFeatureNotInPlan)
}

func TestCreateRole_AdminBypassesFeatureGateForPresets(t *testing.T) {
	logger := slog.Default()
	roleRepo := &mockRoleRepo{}
	permRepo := &mockPermRepo{}
	permID := uuid.New()

	// Admin carries Free entitlements but must still create preset roles.
	caller := domain.Caller{UserID: uuid.New(), IsAdmin: true, Ent: domain.PlanCatalog[domain.PlanFree]}
	ctx := domain.WithCaller(context.Background(), caller)

	permRepo.On("FindByIDs", ctx, []uuid.UUID{permID}).Return([]domain.Permission{{ID: permID, Name: "users:view"}}, nil)
	roleRepo.On("Create", ctx, mock.AnythingOfType("*domain.Role")).Return(nil)
	roleRepo.On("SetPermissions", ctx, mock.AnythingOfType("uuid.UUID"), []uuid.UUID{permID}).Return(nil)

	svc := roles.NewService(roleRepo, permRepo, noopTx{}, &auditSpy{}, nil, logger)
	role, err := svc.Create(ctx, domain.CreateRoleDTO{
		IsPreset:      true,
		Name:          "Manager",
		PermissionIDs: []uuid.UUID{permID},
	})
	assert.NoError(t, err)
	assert.Equal(t, "Manager", role.Name)
}

func TestCreateRole_RejectsGrantingUnheldPermission(t *testing.T) {
	roleRepo := &mockRoleRepo{}
	permRepo := &mockPermRepo{}
	permA := uuid.New()
	permC := uuid.New()

	// Caller holds {users:view, users:update} but tries to grant users:delete_any.
	caller := domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"users:view", "users:update"},
		Ent:         domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
	}
	ctx := domain.WithCaller(context.Background(), caller)

	perms := []domain.Permission{
		{ID: permA, Name: "users:view"},
		{ID: permC, Name: "users:delete_any"},
	}
	permRepo.On("FindByIDs", ctx, []uuid.UUID{permA, permC}).Return(perms, nil)

	svc := roles.NewService(roleRepo, permRepo, noopTx{}, &auditSpy{}, nil, slog.Default())

	orgID := uuid.New()
	_, err := svc.Create(ctx, domain.CreateRoleDTO{
		OrganizationID: &orgID,
		Name:           "Escalated",
		PermissionIDs:  []uuid.UUID{permA, permC},
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	roleRepo.AssertNotCalled(t, "Create")
	roleRepo.AssertNotCalled(t, "SetPermissions")
}

func TestCreateRole_AdminBypassesSubsetCheck(t *testing.T) {
	roleRepo := &mockRoleRepo{}
	permRepo := &mockPermRepo{}
	permID := uuid.New()

	// Admin holds no explicit Permissions but may grant anything.
	caller := domain.Caller{UserID: uuid.New(), IsAdmin: true, Ent: domain.PlanCatalog[domain.PlanFree]}
	ctx := domain.WithCaller(context.Background(), caller)

	permRepo.On("FindByIDs", ctx, []uuid.UUID{permID}).Return([]domain.Permission{{ID: permID, Name: "users:delete_any"}}, nil)
	roleRepo.On("Create", ctx, mock.AnythingOfType("*domain.Role")).Return(nil)
	roleRepo.On("SetPermissions", ctx, mock.AnythingOfType("uuid.UUID"), []uuid.UUID{permID}).Return(nil)

	svc := roles.NewService(roleRepo, permRepo, noopTx{}, &auditSpy{}, nil, slog.Default())

	orgID := uuid.New()
	role, err := svc.Create(ctx, domain.CreateRoleDTO{
		OrganizationID: &orgID,
		Name:           "Powerful",
		PermissionIDs:  []uuid.UUID{permID},
	})
	assert.NoError(t, err)
	assert.Equal(t, "Powerful", role.Name)
}

func TestUpdateRole_RejectsGrantingUnheldPermission(t *testing.T) {
	roleRepo := &mockRoleRepo{}
	permRepo := &mockPermRepo{}
	orgID := uuid.New()
	roleID := uuid.New()
	permID := uuid.New()

	// Caller owns the org role but does not hold users:delete_any.
	caller := domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgID,
		Permissions: []string{"users:view"},
	}
	ctx := domain.WithCaller(context.Background(), caller)

	roleRepo.On("FindByID", ctx, roleID).Return(&domain.Role{ID: roleID, OrganizationID: &orgID}, nil)
	permRepo.On("FindByIDs", ctx, []uuid.UUID{permID}).Return([]domain.Permission{{ID: permID, Name: "users:delete_any"}}, nil)

	svc := roles.NewService(roleRepo, permRepo, noopTx{}, &auditSpy{}, nil, slog.Default())

	_, err := svc.Update(ctx, roleID, domain.UpdateRoleDTO{PermissionIDs: []uuid.UUID{permID}})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	roleRepo.AssertNotCalled(t, "Update")
	roleRepo.AssertNotCalled(t, "SetPermissions")
}

func TestUpdateRole_AllowsGrantingHeldPermission(t *testing.T) {
	roleRepo := &mockRoleRepo{}
	permRepo := &mockPermRepo{}
	orgID := uuid.New()
	roleID := uuid.New()
	permID := uuid.New()

	caller := domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgID,
		Permissions: []string{"users:view"},
	}
	ctx := domain.WithCaller(context.Background(), caller)

	roleRepo.On("FindByID", ctx, roleID).Return(&domain.Role{ID: roleID, OrganizationID: &orgID}, nil)
	permRepo.On("FindByIDs", ctx, []uuid.UUID{permID}).Return([]domain.Permission{{ID: permID, Name: "users:view"}}, nil)
	roleRepo.On("Update", ctx, mock.AnythingOfType("*domain.Role")).Return(nil)
	roleRepo.On("SetPermissions", ctx, roleID, []uuid.UUID{permID}).Return(nil)

	svc := roles.NewService(roleRepo, permRepo, noopTx{}, &auditSpy{}, nil, slog.Default())

	role, err := svc.Update(ctx, roleID, domain.UpdateRoleDTO{PermissionIDs: []uuid.UUID{permID}})
	assert.NoError(t, err)
	assert.Equal(t, roleID, role.ID)
	roleRepo.AssertExpectations(t)
}

func TestAdminList_ForcesIncludePreset(t *testing.T) {
	roleRepo := &mockRoleRepo{}
	svc := roles.NewService(roleRepo, &mockPermRepo{}, noopTx{}, &auditSpy{}, nil, slog.Default())

	orgID := uuid.New()
	ctx := context.Background()
	roleRepo.On("AdminList", ctx, mock.MatchedBy(func(f domain.AdminRoleFilter) bool {
		return f.IncludePreset && f.OrganizationID != nil && *f.OrganizationID == orgID
	})).Return([]domain.Role{{Name: domain.PresetRoleManager, IsPreset: true}}, int64(1), nil)

	list, total, err := svc.AdminList(ctx, domain.AdminRoleFilter{OrganizationID: &orgID})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	roleRepo.AssertExpectations(t)
}

func TestDeleteRole_RecordsAudit(t *testing.T) {
	roleRepo := &mockRoleRepo{}
	spy := &auditSpy{}
	orgID := uuid.New()
	roleID := uuid.New()

	caller := domain.Caller{UserID: uuid.New(), OrgID: &orgID, Permissions: []string{"roles:delete"}}
	ctx := domain.WithCaller(context.Background(), caller)

	roleRepo.On("FindByID", ctx, roleID).Return(&domain.Role{ID: roleID, OrganizationID: &orgID, Name: "Grader"}, nil)
	roleRepo.On("Delete", ctx, roleID).Return(nil)

	svc := roles.NewService(roleRepo, &mockPermRepo{}, noopTx{}, spy, nil, slog.Default())

	err := svc.Delete(ctx, roleID)
	assert.NoError(t, err)
	assert.Len(t, spy.records, 1)
	assert.Equal(t, domain.AuditDeleted, spy.records[0].Action)
	assert.Equal(t, domain.AuditTargetRole, spy.records[0].TargetType)
	assert.Equal(t, "Grader", spy.records[0].TargetLabel)
	assert.NotNil(t, spy.records[0].TargetID)
	assert.Equal(t, roleID, *spy.records[0].TargetID)
}
