//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/audit"
	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/roles"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

func TestRoleCreationAndAssignment(t *testing.T) {
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Permission{},
		&domain.Role{},
		&domain.RolePermission{},
	))

	perms := []domain.PermissionName{
		domain.PermUsersView, domain.PermUsersCreate, domain.PermUsersViewAny,
	}
	for _, p := range perms {
		db.Create(&domain.Permission{Name: p})
	}

	logger := slog.Default()
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiry: 3600_000_000_000}

	jwtService := auth.NewJWTService(cfg)

	orgRepo := organizations.NewRepository(db)
	org := &domain.Organization{Name: "Test Org"}
	require.NoError(t, orgRepo.Create(context.Background(), org))
	userRepo := users.NewRepository(db)
	admin := seedUser(t, userRepo, &org.ID, "role-admin", true)

	roleRepo := roles.NewRoleRepository(db)
	permRepo := roles.NewPermissionRepository(db)

	transactor := database.NewTransactor(db)
	auditSvc := audit.NewService(audit.NewRepository(db), logger)
	roleSvc := roles.NewService(roleRepo, permRepo, transactor, auditSvc, nil, logger)
	roleHandler := roles.NewHandler(roleSvc, permRepo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	authMiddleware := auth.Middleware(jwtService, nil, roleRepo, userRepo, nil)
	perm := func(domain.PermissionName) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }
	roleHandler.RegisterRoutes(v1, authMiddleware, perm)

	adminToken, _ := jwtService.GenerateToken(admin.ID)

	var dbPerms []domain.Permission
	require.NoError(t, db.Find(&dbPerms).Error)
	var permIDs []uuid.UUID
	for _, p := range dbPerms {
		permIDs = append(permIDs, p.ID)
	}

	body, _ := json.Marshal(domain.CreateRoleDTO{
		OrganizationID: &org.ID,
		Name:           "Teacher",
		PermissionIDs:  permIDs,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	roleData := resp["data"].(map[string]interface{})
	assert.Equal(t, "Teacher", roleData["name"])

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
