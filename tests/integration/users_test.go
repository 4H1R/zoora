//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/roles"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

func TestAdminCreatesUser(t *testing.T) {
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Permission{},
		&domain.Role{},
		&domain.RolePermission{},
	))

	logger := slog.Default()
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: 3600_000_000_000}

	jwtService := auth.NewJWTService(cfg)

	orgRepo := organizations.NewRepository(db)
	org := &domain.Organization{Name: "Test Uni"}
	require.NoError(t, orgRepo.Create(context.Background(), org))

	userRepo := users.NewRepository(db)
	roleRepo := roles.NewRoleRepository(db)
	admin := seedUser(t, userRepo, &org.ID, "users-admin", true)

	userSvc := users.NewService(userRepo, roleRepo, nil, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	authMiddleware := auth.Middleware(jwtService, nil, roleRepo, userRepo)
	perm := auth.RequirePermission
	handler := users.NewHandler(userSvc)
	handler.RegisterRoutes(v1, authMiddleware, perm)

	adminToken, _ := jwtService.GenerateToken(admin.ID)

	body, _ := json.Marshal(domain.CreateUserDTO{
		OrganizationID: &org.ID,
		Username:       "student1",
		Name:           "Student One",
		Password:       "password123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}
