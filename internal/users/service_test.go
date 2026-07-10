package users_test

import (
	"context"
	"testing"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/users"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) FindByIDWithPermissions(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) FindByUsernameAndOrg(ctx context.Context, username string, orgID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, username, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) SearchActiveInOrg(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]domain.User, error) {
	args := m.Called(ctx, orgID, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}
func (m *mockUserRepo) FindAdminByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) List(ctx context.Context, scope domain.UserListScope, p domain.ListParams) ([]domain.User, int64, error) {
	args := m.Called(ctx, scope, p)
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}
func (m *mockUserRepo) StatusCounts(ctx context.Context, scope domain.UserListScope) (domain.UserStatusCounts, error) {
	args := m.Called(ctx, scope)
	return args.Get(0).(domain.UserStatusCounts), args.Error(1)
}
func (m *mockUserRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) AdminList(ctx context.Context, q domain.AdminListUsersQuery) ([]domain.User, int64, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}
func (m *mockUserRepo) CountAll(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *mockRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *mockRoleRepo) FindPresetByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
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
	return args.Get(0).(*domain.RoleStats), args.Error(1)
}
func (m *mockRoleRepo) GetPermissionNames(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]string), args.Error(1)
}

func TestCreateUser_AsAdmin(t *testing.T) {
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
	logger := slog.Default()
	repo := &mockUserRepo{}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	user, err := svc.Create(ctx, domain.CreateUserDTO{
		Username: "newuser",
		Name:     "New User",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "newuser", user.Username)
	repo.AssertExpectations(t)
}

// fakeEntService lets the seat-limit test inject a canned CheckUserLimit result.
type fakeEntService struct{ userLimitErr error }

func (f fakeEntService) CheckUserLimit(context.Context, uuid.UUID, domain.Entitlements) error {
	return f.userLimitErr
}
func (f fakeEntService) CheckStorageLimit(context.Context, uuid.UUID, domain.Entitlements, int64) error {
	return nil
}
func (f fakeEntService) CheckConcurrentRoomsLimit(context.Context, uuid.UUID, domain.Entitlements) error {
	return nil
}

func TestCreateUser_SeatLimitReached(t *testing.T) {
	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(),
		OrgID:  &orgID,
		Ent:    domain.PlanCatalog[domain.PlanFree],
	})
	repo := &mockUserRepo{}
	ent := fakeEntService{userLimitErr: domain.NewLimitError(domain.PlanFree, domain.LimitMaxUsers, 10, 10)}
	svc := users.NewService(repo, &mockRoleRepo{}, ent, nil, slog.Default())

	_, err := svc.Create(ctx, domain.CreateUserDTO{Username: "u", Name: "U", Password: "password123"})
	assert.ErrorIs(t, err, domain.ErrPlanLimitReached)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateUser_UnderSeatLimitAllows(t *testing.T) {
	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(),
		OrgID:  &orgID,
		Ent:    domain.PlanCatalog[domain.PlanFree],
	})
	repo := &mockUserRepo{}
	repo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	svc := users.NewService(repo, &mockRoleRepo{}, fakeEntService{}, nil, slog.Default())

	_, err := svc.Create(ctx, domain.CreateUserDTO{Username: "u", Name: "U", Password: "password123"})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreateUser_DuplicateReturnedByRepo(t *testing.T) {
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
	logger := slog.Default()
	repo := &mockUserRepo{}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(domain.ErrConflict)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	_, err := svc.Create(ctx, domain.CreateUserDTO{
		Username: "newuser",
		Name:     "New User",
		Password: "password123",
	})

	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestGetProfile(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	repo := &mockUserRepo{}

	userID := uuid.New()
	expected := &domain.User{ID: userID, Name: "Test"}
	repo.On("FindByIDWithPermissions", ctx, userID).Return(expected, nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	user, err := svc.GetProfile(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, "Test", user.Name)
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	repo := &mockUserRepo{}

	userID := uuid.New()
	existing := &domain.User{ID: userID, Name: "Old Name"}
	repo.On("FindByID", ctx, userID).Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	user, err := svc.UpdateProfile(ctx, userID, domain.UpdateProfileDTO{Name: "New Name"})
	assert.NoError(t, err)
	assert.Equal(t, "New Name", user.Name)
}

func TestList_NoCaller_Forbidden(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	repo := &mockUserRepo{}

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	_, _, err := svc.List(ctx, domain.ListParams{}, nil)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestList_AdminGetsAll(t *testing.T) {
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
	logger := slog.Default()
	repo := &mockUserRepo{}

	userList := []domain.User{{Name: "User 1"}}
	repo.On("List", ctx, mock.MatchedBy(func(scope domain.UserListScope) bool {
		return scope.All && scope.OrganizationID == nil
	}), mock.AnythingOfType("domain.ListParams")).Return(userList, int64(1), nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	result, total, err := svc.List(ctx, domain.ListParams{Page: 1, PageSize: domain.DefaultPageSize}, nil)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
}

func TestList_ViewAnyScopedToOrg(t *testing.T) {
	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgID,
		Permissions: []string{string(domain.PermUsersViewAny)},
	})
	logger := slog.Default()
	repo := &mockUserRepo{}

	userList := []domain.User{{Name: "User 1"}}
	repo.On("List", ctx, mock.MatchedBy(func(scope domain.UserListScope) bool {
		return !scope.All && scope.OrganizationID != nil && *scope.OrganizationID == orgID && scope.UserID == nil
	}), mock.AnythingOfType("domain.ListParams")).Return(userList, int64(1), nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	result, total, err := svc.List(ctx, domain.ListParams{Page: 1, PageSize: domain.DefaultPageSize}, nil)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
}

func TestList_ViewOnlyScopedToSelf(t *testing.T) {
	orgID := uuid.New()
	callerID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      callerID,
		OrgID:       &orgID,
		Permissions: []string{string(domain.PermUsersView)},
	})
	logger := slog.Default()
	repo := &mockUserRepo{}

	userList := []domain.User{{Name: "Me"}}
	repo.On("List", ctx, mock.MatchedBy(func(scope domain.UserListScope) bool {
		return !scope.All && scope.OrganizationID == nil && scope.UserID != nil && *scope.UserID == callerID
	}), mock.AnythingOfType("domain.ListParams")).Return(userList, int64(1), nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	result, total, err := svc.List(ctx, domain.ListParams{Page: 1, PageSize: domain.DefaultPageSize}, nil)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
}

func TestGetByID_NoCaller_Forbidden(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	repo := &mockUserRepo{}

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	_, err := svc.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDelete_NoCaller_Forbidden(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	repo := &mockUserRepo{}

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	err := svc.Delete(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDelete_SelfDelete_Forbidden(t *testing.T) {
	callerID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: callerID, IsAdmin: true})
	repo := &mockUserRepo{}

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, slog.Default())

	err := svc.Delete(ctx, callerID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Delete")
}

func TestDelete_Success(t *testing.T) {
	callerID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: callerID, IsAdmin: true})
	logger := slog.Default()
	repo := &mockUserRepo{}

	userID := uuid.New()
	repo.On("Delete", ctx, userID).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	err := svc.Delete(ctx, userID)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	repo := &mockUserRepo{}

	userID := uuid.New()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	repo.On("FindByID", ctx, userID).Return(&domain.User{ID: userID, Password: string(hashed)}, nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	err := svc.ChangePassword(ctx, userID, domain.ChangePasswordDTO{
		CurrentPassword: "wrongpass",
		NewPassword:     "newStrongPass1!",
	})
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestChangePassword_Success(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	repo := &mockUserRepo{}

	userID := uuid.New()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	repo.On("FindByID", ctx, userID).Return(&domain.User{ID: userID, Password: string(hashed)}, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, logger)

	err := svc.ChangePassword(ctx, userID, domain.ChangePasswordDTO{
		CurrentPassword: "correctpass",
		NewPassword:     "newStrongPass1!",
	})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_Disable_SetsFields(t *testing.T) {
	repo := &mockUserRepo{}
	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, slog.Default())

	targetID := uuid.New()
	callerID := uuid.New()
	target := &domain.User{ID: targetID}
	repo.On("FindByID", mock.Anything, targetID).Return(target, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.DisabledAt != nil && u.DisabledBy != nil && *u.DisabledBy == callerID &&
			u.DisabledReason != nil && *u.DisabledReason == "left the org"
	})).Return(nil)

	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: callerID, IsAdmin: true})
	got, err := svc.Disable(ctx, targetID, domain.DisableUserDTO{Reason: "left the org"})

	assert.NoError(t, err)
	assert.NotNil(t, got.DisabledAt)
	repo.AssertExpectations(t)
}

func TestService_Disable_CannotDisableSelf(t *testing.T) {
	repo := &mockUserRepo{}
	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, slog.Default())

	id := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: id, IsAdmin: true})
	_, err := svc.Disable(ctx, id, domain.DisableUserDTO{})

	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestService_Disable_NonAdminCannotDisableAdmin(t *testing.T) {
	repo := &mockUserRepo{}
	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, slog.Default())

	orgID := uuid.New()
	targetID := uuid.New()
	target := &domain.User{ID: targetID, IsAdmin: true, OrganizationID: &orgID}
	repo.On("FindByID", mock.Anything, targetID).Return(target, nil)

	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: false, OrgID: &orgID})
	_, err := svc.Disable(ctx, targetID, domain.DisableUserDTO{})

	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestService_Disable_Idempotent(t *testing.T) {
	repo := &mockUserRepo{}
	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, slog.Default())

	now := time.Now()
	targetID := uuid.New()
	target := &domain.User{ID: targetID, DisabledAt: &now}
	repo.On("FindByID", mock.Anything, targetID).Return(target, nil)

	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
	got, err := svc.Disable(ctx, targetID, domain.DisableUserDTO{})

	assert.NoError(t, err)
	assert.Equal(t, &now, got.DisabledAt)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestService_Enable_ClearsFields(t *testing.T) {
	repo := &mockUserRepo{}
	svc := users.NewService(repo, &mockRoleRepo{}, nil, nil, slog.Default())

	now := time.Now()
	by := uuid.New()
	reason := "x"
	targetID := uuid.New()
	target := &domain.User{ID: targetID, DisabledAt: &now, DisabledBy: &by, DisabledReason: &reason}
	repo.On("FindByID", mock.Anything, targetID).Return(target, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.DisabledAt == nil && u.DisabledBy == nil && u.DisabledReason == nil
	})).Return(nil)

	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
	got, err := svc.Enable(ctx, targetID)

	assert.NoError(t, err)
	assert.Nil(t, got.DisabledAt)
	repo.AssertExpectations(t)
}
