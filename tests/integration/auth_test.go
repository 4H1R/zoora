//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

func TestLoginFlow(t *testing.T) {
	db := testutil.SetupPostgres(t)
	redisClient := testutil.SetupRedis(t)

	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
	))

	logger := slog.Default()
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: 3600_000_000_000}

	jwtService := auth.NewJWTService(cfg)
	userRepo := users.NewRepository(db)

	authSvc := auth.NewAuthService(userRepo, jwtService, redisClient, logger)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	org := &domain.Organization{Name: "Test Org"}
	require.NoError(t, db.Create(org).Error)
	user := &domain.User{
		OrganizationID: &org.ID,
		Username:       "testuser",
		Name:           "Test User",
		Password:       string(hashed),
	}
	require.NoError(t, db.Create(user).Error)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	handler := auth.NewHandler(authSvc)
	handler.RegisterRoutes(v1)

	body, _ := json.Marshal(domain.LoginDTO{
		Username: "testuser",
		Password: "password123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
}
