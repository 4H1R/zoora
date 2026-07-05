package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
)

var _ = redis.Nil

func setupTestRouter() (*gin.Engine, *auth.JWTService) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: 3600_000_000_000}
	svc := auth.NewJWTService(cfg)
	return gin.New(), svc
}

func newUserRepo(userID uuid.UUID, isAdmin bool) *mockUserRepo {
	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, userID).Return(&domain.User{
		ID:      userID,
		IsAdmin: isAdmin,
	}, nil)
	return repo
}

func TestMiddleware_ValidToken(t *testing.T) {
	router, svc := setupTestRouter()
	userID := uuid.New()

	token, err := svc.GenerateToken(userID)
	assert.NoError(t, err)

	repo := newUserRepo(userID, false)
	router.GET("/test", auth.Middleware(svc, nil, nil, repo, nil), func(c *gin.Context) {
		assert.Equal(t, userID, auth.GetUserID(c))
		assert.False(t, auth.GetIsAdmin(c))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_NoToken(t *testing.T) {
	router, svc := setupTestRouter()
	router.GET("/test", auth.Middleware(svc, nil, nil, nil, nil), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAdmin_AdminUser(t *testing.T) {
	router, svc := setupTestRouter()
	userID := uuid.New()

	token, _ := svc.GenerateToken(userID)

	repo := newUserRepo(userID, true)
	router.GET("/test", auth.Middleware(svc, nil, nil, repo, nil), auth.RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdmin_NonAdmin(t *testing.T) {
	router, svc := setupTestRouter()
	userID := uuid.New()

	token, _ := svc.GenerateToken(userID)

	repo := newUserRepo(userID, false)
	router.GET("/test", auth.Middleware(svc, nil, nil, repo, nil), auth.RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermission_AdminBypass(t *testing.T) {
	router, svc := setupTestRouter()
	userID := uuid.New()

	token, _ := svc.GenerateToken(userID)

	repo := newUserRepo(userID, true)
	router.GET("/test", auth.Middleware(svc, nil, nil, repo, nil), auth.RequirePermission(domain.PermUsersCreate), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
