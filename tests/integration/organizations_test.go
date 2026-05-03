//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/tests/testutil"
)

func TestOrganizationCRUD(t *testing.T) {
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(&domain.Organization{}))

	logger := slog.Default()
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: 3600_000_000_000}

	jwtService := auth.NewJWTService(cfg)
	orgRepo := organizations.NewRepository(db)
	orgSvc := organizations.NewService(orgRepo, nil, logger)
	handler := organizations.NewHandler(orgSvc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	authMiddleware := auth.Middleware(jwtService, nil, nil, nil)
	perm := func(string) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }
	handler.RegisterRoutes(v1, authMiddleware, perm)

	adminToken, _ := jwtService.GenerateToken(uuid.New(), nil, true)

	body, _ := json.Marshal(domain.CreateOrganizationDTO{
		Name:        "Test University",
		Description: "A test organization",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	orgData := createResp["data"].(map[string]interface{})
	orgID := orgData["id"].(string)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/organizations/"+orgID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/organizations", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	nonAdminToken, _ := jwtService.GenerateToken(uuid.New(), nil, false)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/organizations", nil)
	req.Header.Set("Authorization", "Bearer "+nonAdminToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
