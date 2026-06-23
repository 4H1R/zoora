package organizations_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/middleware"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type mockOrganizationSvc struct{ mock.Mock }

func (m *mockOrganizationSvc) Create(ctx context.Context, dto domain.CreateOrganizationDTO) (*domain.Organization, error) {
	a := m.Called(ctx, dto)
	org, _ := a.Get(0).(*domain.Organization)
	return org, a.Error(1)
}

func (m *mockOrganizationSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	a := m.Called(ctx, id)
	org, _ := a.Get(0).(*domain.Organization)
	return org, a.Error(1)
}

func (m *mockOrganizationSvc) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateOrganizationDTO) (*domain.Organization, error) {
	a := m.Called(ctx, id, dto)
	org, _ := a.Get(0).(*domain.Organization)
	return org, a.Error(1)
}

func (m *mockOrganizationSvc) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockOrganizationSvc) List(ctx context.Context, f domain.OrganizationFilter) ([]domain.Organization, int64, error) {
	a := m.Called(ctx, f)
	orgs, _ := a.Get(0).([]domain.Organization)
	return orgs, a.Get(1).(int64), a.Error(2)
}

func (m *mockOrganizationSvc) GetStats(ctx context.Context) (*domain.OrganizationStats, error) {
	a := m.Called(ctx)
	stats, _ := a.Get(0).(*domain.OrganizationStats)
	return stats, a.Error(1)
}

func (m *mockOrganizationSvc) AdminList(ctx context.Context, q domain.AdminListOrganizationsQuery) ([]domain.Organization, int64, error) {
	a := m.Called(ctx, q)
	orgs, _ := a.Get(0).([]domain.Organization)
	return orgs, a.Get(1).(int64), a.Error(2)
}

func (m *mockOrganizationSvc) AdminCreate(ctx context.Context, dto domain.AdminCreateOrganizationDTO) (*domain.Organization, error) {
	a := m.Called(ctx, dto)
	org, _ := a.Get(0).(*domain.Organization)
	return org, a.Error(1)
}

func (m *mockOrganizationSvc) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateOrganizationDTO) (*domain.Organization, error) {
	a := m.Called(ctx, id, dto)
	org, _ := a.Get(0).(*domain.Organization)
	return org, a.Error(1)
}

func (m *mockOrganizationSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockOrganizationSvc) AdminRestore(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func newOrganizationRouter(t *testing.T) (*gin.Engine, *mockOrganizationSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()

	svc := &mockOrganizationSvc{}
	h := organizations.NewHandler(svc)

	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	h.RegisterRoutes(r.Group("/api/v1"), noop, perm)
	return r, svc
}

func newOrganizationAdminRouter(t *testing.T) (*gin.Engine, *mockOrganizationSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()

	svc := &mockOrganizationSvc{}
	h := organizations.NewAdminHandler(svc)

	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	h.RegisterAdminRoutes(r.Group("/admin"))
	return r, svc
}

func do(t *testing.T, r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHandlerGetSuccess(t *testing.T) {
	r, svc := newOrganizationRouter(t)
	id := uuid.New()
	svc.On("GetByID", mock.Anything, id).
		Return(&domain.Organization{ID: id, Name: "Zoora", Status: domain.OrganizationStatusActive}, nil)

	w := do(t, r, http.MethodGet, "/api/v1/organizations/"+id.String(), nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Zoora")
}

func TestHandlerGetInvalidUUIDMaps400(t *testing.T) {
	r, svc := newOrganizationRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/organizations/bad", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "GetByID")
}

func TestHandlerUpdateValidationDoesNotCallService(t *testing.T) {
	r, svc := newOrganizationRouter(t)
	id := uuid.New()

	w := do(t, r, http.MethodPut, "/api/v1/organizations/"+id.String(), map[string]any{"name": "A"})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Update")
}

func TestHandlerUpdateNotFoundMaps404(t *testing.T) {
	r, svc := newOrganizationRouter(t)
	id := uuid.New()
	name := "Updated"
	svc.On("Update", mock.Anything, id, domain.UpdateOrganizationDTO{Name: &name}).
		Return((*domain.Organization)(nil), domain.ErrNotFound)

	w := do(t, r, http.MethodPut, "/api/v1/organizations/"+id.String(), map[string]any{"name": name})

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminHandlerListBindsFiltersAndListParams(t *testing.T) {
	r, svc := newOrganizationAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.MatchedBy(func(q domain.AdminListOrganizationsQuery) bool {
		return q.Status != nil &&
			*q.Status == domain.OrganizationStatusTrial &&
			q.IncludeDeleted &&
			q.ListParams.Search == "school" &&
			q.ListParams.OrderBy == "name" &&
			q.ListParams.OrderDir == "asc"
	})).Return([]domain.Organization{{ID: uuid.New(), Name: "School"}}, int64(1), nil)

	w := do(t, r, http.MethodGet, "/admin/organizations?status=trial&include_deleted=true&search=school&order_by=name&order_dir=asc", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestAdminHandlerListRejectsInvalidStatus(t *testing.T) {
	r, svc := newOrganizationAdminRouter(t)

	w := do(t, r, http.MethodGet, "/admin/organizations?status=paused", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "AdminList")
}

func TestAdminHandlerStatsSuccess(t *testing.T) {
	r, svc := newOrganizationAdminRouter(t)
	svc.On("GetStats", mock.Anything).
		Return(&domain.OrganizationStats{TotalOrganizations: 3, ActiveCount: 2, TotalUsers: 15}, nil)

	w := do(t, r, http.MethodGet, "/admin/organizations/stats", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total_organizations":3`)
}

func TestAdminHandlerCreateValidationDoesNotCallService(t *testing.T) {
	r, svc := newOrganizationAdminRouter(t)

	w := do(t, r, http.MethodPost, "/admin/organizations", map[string]any{
		"name":   "A",
		"status": "invalid",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "AdminCreate")
}

func TestAdminHandlerCreateConflictMaps409(t *testing.T) {
	r, svc := newOrganizationAdminRouter(t)
	dto := domain.AdminCreateOrganizationDTO{Name: "Zoora", Slug: "zoora", Status: domain.OrganizationStatusActive}
	svc.On("AdminCreate", mock.Anything, dto).Return((*domain.Organization)(nil), domain.ErrConflict)

	w := do(t, r, http.MethodPost, "/admin/organizations", map[string]any{
		"name":   "Zoora",
		"slug":   "zoora",
		"status": "active",
	})

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestAdminHandlerHardDeleteAndRestoreMapErrors(t *testing.T) {
	r, svc := newOrganizationAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(domain.ErrForbidden)
	svc.On("AdminRestore", mock.Anything, id).Return(domain.ErrNotFound)

	deleteResp := do(t, r, http.MethodDelete, "/admin/organizations/"+id.String(), nil)
	restoreResp := do(t, r, http.MethodPost, "/admin/organizations/"+id.String()+"/restore", nil)

	assert.Equal(t, http.StatusForbidden, deleteResp.Code)
	assert.Equal(t, http.StatusNotFound, restoreResp.Code)
}
