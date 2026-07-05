package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
)

func revokeTestSetup(t *testing.T) (*gin.Engine, *auth.JWTService, *redis.Client, *miniredis.Miniredis) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{JWTSecret: "s", JWTExpiry: time.Hour}
	svc := auth.NewJWTService(cfg)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return gin.New(), svc, rdb, mr
}

func TestMiddleware_RevokedBeforeIssuance_Rejects(t *testing.T) {
	router, svc, rdb, _ := revokeTestSetup(t)
	userID := uuid.New()

	// Set revocation marker to now+10min → any token issued now has
	// IssuedAt < revokedAt, so it must be rejected.
	future := time.Now().Add(10 * time.Minute).Unix()
	assert.NoError(t, rdb.Set(context.Background(), auth.RevokedKey(userID.String()),
		strconv.FormatInt(future, 10), time.Hour).Err())

	token, _ := svc.GenerateToken(userID)
	router.GET("/t", auth.Middleware(svc, rdb, nil, nil, nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/t", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_IssuedAfterRevocation_Accepts(t *testing.T) {
	router, svc, rdb, _ := revokeTestSetup(t)
	userID := uuid.New()

	past := time.Now().Add(-10 * time.Minute).Unix()
	assert.NoError(t, rdb.Set(context.Background(), auth.RevokedKey(userID.String()),
		strconv.FormatInt(past, 10), time.Hour).Err())

	token, _ := svc.GenerateToken(userID)
	router.GET("/t", auth.Middleware(svc, rdb, nil, nil, nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/t", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_NoRevocationMarker_Accepts(t *testing.T) {
	router, svc, rdb, _ := revokeTestSetup(t)
	userID := uuid.New()

	token, _ := svc.GenerateToken(userID)
	router.GET("/t", auth.Middleware(svc, rdb, nil, nil, nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/t", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_NilRedis_SkipsRevocationCheck(t *testing.T) {
	router, svc, _, _ := revokeTestSetup(t)
	userID := uuid.New()

	token, _ := svc.GenerateToken(userID)
	// nil rdb → no check path.
	router.GET("/t", auth.Middleware(svc, nil, nil, nil, nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/t", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RedisError_FailsClosed(t *testing.T) {
	router, svc, rdb, mr := revokeTestSetup(t)
	userID := uuid.New()
	token, _ := svc.GenerateToken(userID)

	// Corrupt the revocation marker to a non-integer so parsing fails →
	// middleware must treat as revoked and reject.
	assert.NoError(t, rdb.Set(context.Background(), auth.RevokedKey(userID.String()),
		"not-a-number", time.Hour).Err())

	router.GET("/t", auth.Middleware(svc, rdb, nil, nil, nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/t", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Now close miniredis entirely; connection errors should also fail closed.
	mr.Close()
	req2, _ := http.NewRequest("GET", "/t", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}
