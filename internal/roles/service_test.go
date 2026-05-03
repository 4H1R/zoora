package roles_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"log/slog"

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

func TestCreateRole(t *testing.T) {
	logger := slog.Default()
	roleRepo := &mockRoleRepo{}
	permRepo := &mockPermRepo{}

	permID := uuid.New()

	caller := domain.Caller{UserID: uuid.New()}
	ctx := domain.WithCaller(context.Background(), caller)

	perms := []domain.Permission{{ID: permID, Name: "users:view"}}
	permRepo.On("FindByIDs", ctx, []uuid.UUID{permID}).Return(perms, nil)
	roleRepo.On("Create", ctx, mock.AnythingOfType("*domain.Role")).Return(nil)
	roleRepo.On("SetPermissions", ctx, mock.AnythingOfType("uuid.UUID"), []uuid.UUID{permID}).Return(nil)

	svc := roles.NewService(roleRepo, permRepo, noopTx{}, nil, logger)

	orgID := uuid.New()
	role, err := svc.Create(ctx, domain.CreateRoleDTO{
		OrganizationID: &orgID,
		Name:           "Viewer",
		PermissionIDs:  []uuid.UUID{permID},
	})

	assert.NoError(t, err)
	assert.Equal(t, "Viewer", role.Name)
}
