package users_test

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
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/users"
)

// mockUserSvc is a full domain.UserService mock. Only AdminXxx methods are
// exercised by handler tests; others return zero values so the interface is
// satisfied without panics.
type mockUserSvc struct{ mock.Mock }

func (m *mockUserSvc) Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error) {
	a := m.Called(ctx, dto)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	a := m.Called(ctx, id)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateUserDTO) (*domain.User, error) {
	a := m.Called(ctx, id, dto)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserSvc) List(ctx context.Context, p domain.ListParams, disabled *bool) ([]domain.User, int64, error) {
	a := m.Called(ctx, p, disabled)
	us, _ := a.Get(0).([]domain.User)
	return us, a.Get(1).(int64), a.Error(2)
}
func (m *mockUserSvc) StatusCounts(ctx context.Context) (domain.UserStatusCounts, error) {
	a := m.Called(ctx)
	return a.Get(0).(domain.UserStatusCounts), a.Error(1)
}
func (m *mockUserSvc) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	a := m.Called(ctx, id)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) ChangePassword(ctx context.Context, id uuid.UUID, dto domain.ChangePasswordDTO) (string, error) {
	a := m.Called(ctx, id, dto)
	return a.String(0), a.Error(1)
}
func (m *mockUserSvc) AdminGetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	a := m.Called(ctx, id)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) AdminList(ctx context.Context, q domain.AdminListUsersQuery) ([]domain.User, int64, error) {
	a := m.Called(ctx, q)
	us, _ := a.Get(0).([]domain.User)
	return us, a.Get(1).(int64), a.Error(2)
}
func (m *mockUserSvc) AdminCreate(ctx context.Context, dto domain.AdminCreateUserDTO) (*domain.User, error) {
	a := m.Called(ctx, dto)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateUserDTO) (*domain.User, error) {
	a := m.Called(ctx, id, dto)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) AdminForceResetPassword(ctx context.Context, id uuid.UUID, dto domain.AdminForceResetPasswordDTO) error {
	return m.Called(ctx, id, dto).Error(0)
}
func (m *mockUserSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserSvc) AssignRole(ctx context.Context, userID uuid.UUID, dto domain.AssignRoleDTO) (*domain.User, error) {
	a := m.Called(ctx, userID, dto)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) RemoveRole(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	a := m.Called(ctx, userID)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) Disable(ctx context.Context, id uuid.UUID, dto domain.DisableUserDTO) (*domain.User, error) {
	a := m.Called(ctx, id, dto)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}
func (m *mockUserSvc) Enable(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	a := m.Called(ctx, id)
	u, _ := a.Get(0).(*domain.User)
	return u, a.Error(1)
}

// mockAuthSvc satisfies domain.AuthService for the revoke-sessions handler.
type mockAuthSvc struct{ mock.Mock }

func (m *mockAuthSvc) Login(ctx context.Context, dto domain.LoginDTO, orgID *uuid.UUID) (*domain.User, string, error) {
	a := m.Called(ctx, dto, orgID)
	u, _ := a.Get(0).(*domain.User)
	return u, a.String(1), a.Error(2)
}
func (m *mockAuthSvc) AdminRevokeSessions(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func adminHandlerRouter(t *testing.T) (*gin.Engine, *mockUserSvc, *mockAuthSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	// Ensure custom validators (strongpassword) are registered for bind tests.
	_ = httpx.RegisterValidators()

	usrSvc := &mockUserSvc{}
	auSvc := &mockAuthSvc{}
	h := users.NewAdminHandler(usrSvc, auSvc)

	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	grp := r.Group("/admin")
	h.RegisterAdminRoutes(grp)
	return r, usrSvc, auSvc
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

func decodeResp(t *testing.T, w *httptest.ResponseRecorder) domain.Response {
	t.Helper()
	var got domain.Response
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	return got
}

func TestAdminHandler_List_Success(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	svc.On("AdminList", mock.Anything, mock.AnythingOfType("domain.AdminListUsersQuery")).
		Return([]domain.User{{ID: uuid.New()}}, int64(1), nil)

	w := do(t, r, "GET", "/admin/users?limit=20&offset=0", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	got := decodeResp(t, w)
	assert.True(t, got.Success)
}

func TestAdminHandler_List_ServiceForbidden_Maps403(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	svc.On("AdminList", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/admin/users", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
	got := decodeResp(t, w)
	assert.False(t, got.Success)
	assert.Equal(t, "FORBIDDEN", got.Error.Code)
}

func TestAdminHandler_Create_Success(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	svc.On("AdminCreate", mock.Anything, mock.AnythingOfType("domain.AdminCreateUserDTO")).
		Return(&domain.User{ID: uuid.New(), Username: "u"}, nil)

	body := map[string]any{
		"username": "user1", "name": "User",
		"password": "Secret1A", "is_admin": true,
	}
	w := do(t, r, "POST", "/admin/users", body)
	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestAdminHandler_Create_BindValidationError_Maps400(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	// Missing required fields and weak password.
	w := do(t, r, "POST", "/admin/users", map[string]any{"username": "u"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	got := decodeResp(t, w)
	assert.Equal(t, "VALIDATION_ERROR", got.Error.Code)
	svc.AssertNotCalled(t, "AdminCreate")
}

func TestAdminHandler_Create_ServiceConflict_Maps409(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	svc.On("AdminCreate", mock.Anything, mock.Anything).
		Return((*domain.User)(nil), domain.ErrConflict)

	body := map[string]any{
		"username": "user1", "name": "User",
		"password": "Secret1A",
	}
	w := do(t, r, "POST", "/admin/users", body)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestAdminHandler_Get_Success(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	id := uuid.New()
	svc.On("AdminGetByID", mock.Anything, id).Return(&domain.User{ID: id}, nil)

	w := do(t, r, "GET", "/admin/users/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_Get_InvalidUUID_Maps400(t *testing.T) {
	r, _, _ := adminHandlerRouter(t)
	w := do(t, r, "GET", "/admin/users/not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandler_Update_Success(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	id := uuid.New()
	svc.On("AdminUpdate", mock.Anything, id, mock.AnythingOfType("domain.AdminUpdateUserDTO")).
		Return(&domain.User{ID: id, Name: "New"}, nil)

	w := do(t, r, "PUT", "/admin/users/"+id.String(), map[string]any{"name": "New"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_Update_NotFound_Maps404(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	id := uuid.New()
	svc.On("AdminUpdate", mock.Anything, id, mock.Anything).
		Return((*domain.User)(nil), domain.ErrNotFound)

	w := do(t, r, "PUT", "/admin/users/"+id.String(), map[string]any{})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminHandler_HardDelete_Success(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/admin/users/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_HardDelete_Forbidden_Maps403(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(domain.ErrForbidden)

	w := do(t, r, "DELETE", "/admin/users/"+id.String(), nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminHandler_ForceResetPassword_Success(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	id := uuid.New()
	svc.On("AdminForceResetPassword", mock.Anything, id, mock.AnythingOfType("domain.AdminForceResetPasswordDTO")).
		Return(nil)

	w := do(t, r, "POST", "/admin/users/"+id.String()+"/password",
		map[string]any{"new_password": "NewPass1A"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_ForceResetPassword_WeakPassword_Maps400(t *testing.T) {
	r, svc, _ := adminHandlerRouter(t)
	id := uuid.New()
	w := do(t, r, "POST", "/admin/users/"+id.String()+"/password",
		map[string]any{"new_password": "weak"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "AdminForceResetPassword")
}

func TestAdminHandler_RevokeSessions_Success(t *testing.T) {
	r, _, auSvc := adminHandlerRouter(t)
	id := uuid.New()
	auSvc.On("AdminRevokeSessions", mock.Anything, id).Return(nil)

	w := do(t, r, "POST", "/admin/users/"+id.String()+"/revoke-sessions", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	auSvc.AssertExpectations(t)
}

func TestAdminHandler_RevokeSessions_ServiceError_Maps500(t *testing.T) {
	r, _, auSvc := adminHandlerRouter(t)
	id := uuid.New()
	auSvc.On("AdminRevokeSessions", mock.Anything, id).
		Return(assertAnyErr{msg: "boom"})

	w := do(t, r, "POST", "/admin/users/"+id.String()+"/revoke-sessions", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	got := decodeResp(t, w)
	// 5xx messages are redacted by the error middleware.
	assert.Equal(t, "internal server error", got.Error.Message)
}

type assertAnyErr struct{ msg string }

func (e assertAnyErr) Error() string { return e.msg }
