package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

// stubUserRepo satisfies domain.UserRepository via the embedded (nil) interface;
// only FindByID is implemented, and it counts calls so tests can assert the
// auth cache spared the DB. Any other method call panics — none should fire.
type stubUserRepo struct {
	domain.UserRepository
	user      *domain.User
	findCalls int
}

func (s *stubUserRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.User, error) {
	s.findCalls++
	return s.user, nil
}

func TestMiddleware_UserCacheHit_SkipsRepo(t *testing.T) {
	router, svc, rdb, _ := revokeTestSetup(t)
	userID := uuid.New()
	orgID := uuid.New()

	// Seed the auth cache; a hit must avoid the repo entirely.
	assert.NoError(t, cache.SetUser(context.Background(), rdb, userID, cache.CachedUser{
		OrganizationID: &orgID,
		Username:       "cached",
		Name:           "Cached User",
	}))

	repo := &stubUserRepo{user: &domain.User{ID: userID}}
	token, _ := svc.GenerateToken(userID)
	router.GET("/t", auth.Middleware(svc, rdb, nil, repo, nil), func(c *gin.Context) {
		caller, _ := domain.CallerFromCtx(c.Request.Context())
		assert.Equal(t, "cached", caller.Username)
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/t", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, repo.findCalls, "cache hit must not touch the repo")
}

func TestMiddleware_UserCacheMiss_LoadsAndPopulates(t *testing.T) {
	router, svc, rdb, _ := revokeTestSetup(t)
	userID := uuid.New()

	repo := &stubUserRepo{user: &domain.User{ID: userID, Username: "fromdb", Name: "From DB"}}
	token, _ := svc.GenerateToken(userID)
	router.GET("/t", auth.Middleware(svc, rdb, nil, repo, nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/t", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, repo.findCalls, "cache miss must hit the repo once")

	// The miss should have populated the cache for the next request.
	cu, err := cache.GetUser(context.Background(), rdb, userID)
	assert.NoError(t, err)
	assert.Equal(t, "fromdb", cu.Username)
}

func TestMiddleware_CachedDisabledUser_Rejects(t *testing.T) {
	router, svc, rdb, _ := revokeTestSetup(t)
	userID := uuid.New()

	disabledAt := time.Unix(1_700_000_000, 0).UTC()
	assert.NoError(t, cache.SetUser(context.Background(), rdb, userID, cache.CachedUser{
		Username:   "locked",
		DisabledAt: &disabledAt,
	}))

	repo := &stubUserRepo{user: &domain.User{ID: userID}}
	token, _ := svc.GenerateToken(userID)
	router.GET("/t", auth.Middleware(svc, rdb, nil, repo, nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/t", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, repo.findCalls, "disabled cache hit must not touch the repo")
}
