package users_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/users"
)

// TestGetByID_CrossTenant_Forbidden verifies a non-admin caller in org A cannot
// read a user that belongs to org B, even though they hold users:view_any.
func TestGetByID_CrossTenant_Forbidden(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	targetID := uuid.New()

	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, targetID).Return(&domain.User{ID: targetID, OrganizationID: &orgB}, nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgA,
		Permissions: []string{string(domain.PermUsersViewAny)},
	})

	_, err := svc.GetByID(ctx, targetID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// TestGetByID_SameTenant_Success verifies a non-admin caller can read a user in
// their own org.
func TestGetByID_SameTenant_Success(t *testing.T) {
	orgA := uuid.New()
	targetID := uuid.New()

	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, targetID).Return(&domain.User{ID: targetID, Name: "Same Org", OrganizationID: &orgA}, nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgA,
		Permissions: []string{string(domain.PermUsersViewAny)},
	})

	user, err := svc.GetByID(ctx, targetID)
	assert.NoError(t, err)
	assert.Equal(t, "Same Org", user.Name)
}

// TestGetByID_Admin_AnyOrg verifies an admin bypasses tenant scoping.
func TestGetByID_Admin_AnyOrg(t *testing.T) {
	orgB := uuid.New()
	targetID := uuid.New()

	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, targetID).Return(&domain.User{ID: targetID, Name: "Other Org", OrganizationID: &orgB}, nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})

	user, err := svc.GetByID(ctx, targetID)
	assert.NoError(t, err)
	assert.Equal(t, "Other Org", user.Name)
}

// TestCreate_NonAdminForcesCallerOrg verifies a non-admin cannot plant a new
// user in another org by supplying a foreign OrganizationID in the DTO.
func TestCreate_NonAdminForcesCallerOrg(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()

	repo := &mockUserRepo{}
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.OrganizationID != nil && *u.OrganizationID == orgA
	})).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgA})

	user, err := svc.Create(ctx, domain.CreateUserDTO{
		OrganizationID: &orgB,
		Username:       "u",
		Name:           "U",
		Password:       "password123",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user.OrganizationID)
	assert.Equal(t, orgA, *user.OrganizationID)
	repo.AssertExpectations(t)
}

// TestAssignRole_ManagerPreset_Forbidden verifies a non-admin with roles:update
// cannot escalate a user (including self) to the Manager preset.
func TestAssignRole_ManagerPreset_Forbidden(t *testing.T) {
	orgA := uuid.New()
	targetID := uuid.New()
	managerRoleID := uuid.New()

	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, targetID).Return(&domain.User{ID: targetID, OrganizationID: &orgA}, nil)

	roleRepo := &mockRoleRepo{}
	roleRepo.On("FindByID", mock.Anything, managerRoleID).Return(&domain.Role{ID: managerRoleID, IsPreset: true, Name: domain.PresetRoleManager}, nil)

	svc := users.NewService(repo, roleRepo, nil, nil, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgA,
		Permissions: []string{string(domain.PermRolesUpdate)},
	})

	_, err := svc.AssignRole(ctx, targetID, domain.AssignRoleDTO{RoleID: managerRoleID})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestAssignRole_CrossTenant_Forbidden verifies a non-admin cannot assign a role
// to a user in another org.
func TestAssignRole_CrossTenant_Forbidden(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	targetID := uuid.New()
	roleID := uuid.New()

	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, targetID).Return(&domain.User{ID: targetID, OrganizationID: &orgB}, nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgA,
		Permissions: []string{string(domain.PermRolesUpdate)},
	})

	_, err := svc.AssignRole(ctx, targetID, domain.AssignRoleDTO{RoleID: roleID})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestAssignRole_SameTenantNormalRole_Success verifies the legitimate path still
// works: non-admin assigning a non-manager role to a same-org user.
func TestAssignRole_SameTenantNormalRole_Success(t *testing.T) {
	orgA := uuid.New()
	targetID := uuid.New()
	roleID := uuid.New()

	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, targetID).Return(&domain.User{ID: targetID, OrganizationID: &orgA}, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.RoleID != nil && *u.RoleID == roleID
	})).Return(nil)

	roleRepo := &mockRoleRepo{}
	roleRepo.On("FindByID", mock.Anything, roleID).Return(&domain.Role{ID: roleID, IsPreset: false, Name: "Custom"}, nil)

	svc := users.NewService(repo, roleRepo, nil, nil, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgA,
		Permissions: []string{string(domain.PermRolesUpdate)},
	})

	user, err := svc.AssignRole(ctx, targetID, domain.AssignRoleDTO{RoleID: roleID})
	assert.NoError(t, err)
	assert.Equal(t, roleID, *user.RoleID)
	repo.AssertExpectations(t)
}

// TestRemoveRole_CrossTenant_Forbidden verifies a non-admin cannot strip a role
// from a user in another org.
func TestRemoveRole_CrossTenant_Forbidden(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	targetID := uuid.New()

	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, targetID).Return(&domain.User{ID: targetID, OrganizationID: &orgB}, nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgA,
		Permissions: []string{string(domain.PermRolesUpdate)},
	})

	_, err := svc.RemoveRole(ctx, targetID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
