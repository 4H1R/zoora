package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// --- Mocks ---

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
func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
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

// --- Tests ---

func TestLogin_ValidCredentials(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: time.Hour}
	jwtService := auth.NewJWTService(cfg)
	logger := slog.Default()

	userRepo := &mockUserRepo{}

	orgID := uuid.New()
	user := &domain.User{
		ID:             uuid.New(),
		OrganizationID: &orgID,
		Username:       "testuser",
		Name:           "Test User",
		Password:       "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "password123"
		IsAdmin:        false,
	}

	userRepo.On("FindByUsername", ctx, "testuser").Return(user, nil)

	svc := auth.NewAuthService(userRepo, jwtService, newTestRedis(t), logger)

	resultUser, token, err := svc.Login(ctx, domain.LoginDTO{
		Username: "testuser",
		Password: "password123",
	})

	if err != nil {
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	} else {
		assert.NotNil(t, resultUser)
		assert.NotEmpty(t, token)
	}
}

func TestLogin_UserNotFound_ReturnsUnauthorized(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: time.Hour}
	jwtService := auth.NewJWTService(cfg)
	logger := slog.Default()

	userRepo := &mockUserRepo{}
	userRepo.On("FindByUsername", ctx, "nonexistent").Return(nil, domain.ErrNotFound)

	svc := auth.NewAuthService(userRepo, jwtService, newTestRedis(t), logger)

	resultUser, token, err := svc.Login(ctx, domain.LoginDTO{
		Username: "nonexistent",
		Password: "password123",
	})

	assert.ErrorIs(t, err, domain.ErrUnauthorized)
	assert.Nil(t, resultUser)
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}

func TestLogin_DisabledUserRejected(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: time.Hour}
	jwtService := auth.NewJWTService(cfg)

	userRepo := &mockUserRepo{}
	now := time.Now()
	hashed, hashErr := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, hashErr)
	user := &domain.User{
		ID:         uuid.New(),
		Username:   "disabled",
		Name:       "Disabled User",
		Password:   string(hashed),
		DisabledAt: &now,
	}
	userRepo.On("FindByUsername", ctx, "disabled").Return(user, nil)

	svc := auth.NewAuthService(userRepo, jwtService, newTestRedis(t), slog.Default())

	resultUser, token, err := svc.Login(ctx, domain.LoginDTO{
		Username: "disabled",
		Password: "password123",
	})

	assert.ErrorIs(t, err, domain.ErrUserDisabled)
	assert.Nil(t, resultUser)
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}
