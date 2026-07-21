//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/middleware"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/orgsettings"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

func TestOrganizationCRUD(t *testing.T) {
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(&domain.Organization{}, &domain.User{}, &domain.OrganizationSettings{}))

	logger := slog.Default()
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: 3600_000_000_000}

	jwtService := auth.NewJWTService(cfg)
	orgRepo := organizations.NewRepository(db)
	userRepo := users.NewRepository(db)
	orgSvc := organizations.NewService(orgRepo, userRepo, orgsettings.NewRepository(db), nil, nil, logger)
	handler := organizations.NewHandler(orgSvc)

	org := &domain.Organization{Name: "Test University", Description: "A test organization"}
	require.NoError(t, orgRepo.Create(t.Context(), org))
	admin := seedUser(t, userRepo, &org.ID, "org-admin", true)
	outsider := seedUser(t, userRepo, nil, "org-outsider", false)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ErrorHandler(logger))
	v1 := router.Group("/api/v1")
	authMiddleware := auth.Middleware(jwtService, nil, nil, userRepo, nil)
	perm := func(domain.PermissionName) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }
	handler.RegisterRoutes(v1, authMiddleware, perm)

	adminToken, _ := jwtService.GenerateToken(admin.ID)
	outsiderToken, _ := jwtService.GenerateToken(outsider.ID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/organizations/"+org.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	newName := "Updated University"
	body, _ := json.Marshal(domain.UpdateOrganizationDTO{Name: &newName})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/organizations/"+org.ID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/organizations/"+org.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+outsiderToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
