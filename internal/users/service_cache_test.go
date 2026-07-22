package users_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
	"github.com/4H1R/zoora/internal/users"
)

// newCacheHarness wires a users service against a real (miniredis) Redis and
// seeds the target user's auth-cache entry so tests can assert a mutation busts
// it. Returns the service, redis client and the seeded user ID.
func newCacheHarness(t *testing.T, repo domain.UserRepository) (domain.UserService, *redis.Client, uuid.UUID) {
	t.Helper()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	userID := uuid.New()
	if err := cache.SetUser(context.Background(), rdb, userID, cache.CachedUser{Username: "cached"}); err != nil {
		t.Fatalf("seeding user cache: %v", err)
	}

	svc := users.NewService(repo, &mockRoleRepo{}, nil, rdb, nil, fakeTransactor{}, &auditSpy{}, slog.Default())
	return svc, rdb, userID
}

func assertCacheBusted(t *testing.T, rdb *redis.Client, userID uuid.UUID) {
	t.Helper()
	if _, err := cache.GetUser(context.Background(), rdb, userID); err == nil {
		t.Fatal("user cache still present after mutation, want busted")
	}
}

func TestChangePassword_BustsUserCache(t *testing.T) {
	repo := &mockUserRepo{}
	svc, rdb, userID := newCacheHarness(t, repo)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass12"), bcrypt.DefaultCost)
	repo.On("FindByID", mock.Anything, userID).Return(&domain.User{ID: userID, Password: string(hashed)}, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	_, err := svc.ChangePassword(context.Background(), userID, domain.ChangePasswordDTO{
		CurrentPassword: "oldpass12",
		NewPassword:     "newpass123",
	})
	assert.NoError(t, err)
	assertCacheBusted(t, rdb, userID)
}

func TestDisable_BustsUserCache(t *testing.T) {
	repo := &mockUserRepo{}
	svc, rdb, userID := newCacheHarness(t, repo)

	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true, OrgID: &orgID})
	repo.On("FindByID", mock.Anything, userID).Return(&domain.User{ID: userID, OrganizationID: &orgID}, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	_, err := svc.Disable(ctx, userID, domain.DisableUserDTO{Reason: "spam"})
	assert.NoError(t, err)
	assertCacheBusted(t, rdb, userID)
}

func TestAssignRole_BustsUserCache(t *testing.T) {
	repo := &mockUserRepo{}
	svc, rdb, userID := newCacheHarness(t, repo)

	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
	repo.On("FindByID", mock.Anything, userID).Return(&domain.User{ID: userID}, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	_, err := svc.AssignRole(ctx, userID, domain.AssignRoleDTO{RoleID: uuid.New()})
	assert.NoError(t, err)
	assertCacheBusted(t, rdb, userID)
}

func TestAdminHardDelete_BustsUserCache(t *testing.T) {
	repo := &mockUserRepo{}
	svc, rdb, userID := newCacheHarness(t, repo)

	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
	repo.On("HardDelete", mock.Anything, userID).Return(nil)

	err := svc.AdminHardDelete(ctx, userID)
	assert.NoError(t, err)
	assertCacheBusted(t, rdb, userID)
}

// A failed repo write must NOT bust the cache — the stale entry is still valid.
func TestChangePassword_RepoErrorKeepsCache(t *testing.T) {
	repo := &mockUserRepo{}
	svc, rdb, userID := newCacheHarness(t, repo)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass12"), bcrypt.DefaultCost)
	repo.On("FindByID", mock.Anything, userID).Return(&domain.User{ID: userID, Password: string(hashed)}, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(assert.AnError)

	_, err := svc.ChangePassword(context.Background(), userID, domain.ChangePasswordDTO{
		CurrentPassword: "oldpass12",
		NewPassword:     "newpass123",
	})
	assert.Error(t, err)
	if _, err := cache.GetUser(context.Background(), rdb, userID); err != nil {
		t.Fatal("user cache was busted despite repo write failure")
	}
}
