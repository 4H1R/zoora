package classes_test

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

	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/middleware"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

// mockClassSvc is a full domain.ClassService mock. Only methods exercised by
// handler tests are stubbed per-case; others just satisfy the interface.
type mockClassSvc struct{ mock.Mock }

func (m *mockClassSvc) Create(ctx context.Context, dto domain.CreateClassDTO) (*domain.Class, error) {
	a := m.Called(ctx, dto)
	c, _ := a.Get(0).(*domain.Class)
	return c, a.Error(1)
}
func (m *mockClassSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	a := m.Called(ctx, id)
	c, _ := a.Get(0).(*domain.Class)
	return c, a.Error(1)
}
func (m *mockClassSvc) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateClassDTO) (*domain.Class, error) {
	a := m.Called(ctx, id, dto)
	c, _ := a.Get(0).(*domain.Class)
	return c, a.Error(1)
}
func (m *mockClassSvc) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockClassSvc) List(ctx context.Context, p domain.ListParams) ([]domain.Class, int64, error) {
	a := m.Called(ctx, p)
	cs, _ := a.Get(0).([]domain.Class)
	return cs, a.Get(1).(int64), a.Error(2)
}

func (m *mockClassSvc) CreateSession(ctx context.Context, classID uuid.UUID, dto domain.CreateClassSessionDTO) (*domain.ClassSession, error) {
	a := m.Called(ctx, classID, dto)
	s, _ := a.Get(0).(*domain.ClassSession)
	return s, a.Error(1)
}
func (m *mockClassSvc) GetSession(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	a := m.Called(ctx, id)
	s, _ := a.Get(0).(*domain.ClassSession)
	return s, a.Error(1)
}
func (m *mockClassSvc) UpdateSession(ctx context.Context, id uuid.UUID, dto domain.UpdateClassSessionDTO) (*domain.ClassSession, error) {
	a := m.Called(ctx, id, dto)
	s, _ := a.Get(0).(*domain.ClassSession)
	return s, a.Error(1)
}
func (m *mockClassSvc) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockClassSvc) ListSessions(ctx context.Context, classID uuid.UUID, q domain.ListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, classID, q)
	ss, _ := a.Get(0).([]domain.ClassSession)
	return ss, a.Get(1).(int64), a.Error(2)
}

func (m *mockClassSvc) Enroll(ctx context.Context, classID uuid.UUID, dto domain.EnrollClassMemberDTO) (*domain.ClassMember, error) {
	a := m.Called(ctx, classID, dto)
	cm, _ := a.Get(0).(*domain.ClassMember)
	return cm, a.Error(1)
}
func (m *mockClassSvc) Leave(ctx context.Context, classID, userID uuid.UUID) error {
	return m.Called(ctx, classID, userID).Error(0)
}
func (m *mockClassSvc) ListMembers(ctx context.Context, classID uuid.UUID, q domain.ListClassMembersQuery) ([]domain.ClassMember, int64, error) {
	a := m.Called(ctx, classID, q)
	ms, _ := a.Get(0).([]domain.ClassMember)
	return ms, a.Get(1).(int64), a.Error(2)
}
func (m *mockClassSvc) ProvisionConversation(ctx context.Context, classID uuid.UUID, dto domain.ProvisionClassConversationDTO) (*domain.Conversation, error) {
	a := m.Called(ctx, classID, dto)
	conv, _ := a.Get(0).(*domain.Conversation)
	return conv, a.Error(1)
}

func (m *mockClassSvc) AdminList(ctx context.Context, q domain.AdminListClassesQuery) ([]domain.Class, int64, error) {
	a := m.Called(ctx, q)
	cs, _ := a.Get(0).([]domain.Class)
	return cs, a.Get(1).(int64), a.Error(2)
}
func (m *mockClassSvc) AdminCreate(ctx context.Context, dto domain.AdminCreateClassDTO) (*domain.Class, error) {
	a := m.Called(ctx, dto)
	c, _ := a.Get(0).(*domain.Class)
	return c, a.Error(1)
}
func (m *mockClassSvc) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateClassDTO) (*domain.Class, error) {
	a := m.Called(ctx, id, dto)
	c, _ := a.Get(0).(*domain.Class)
	return c, a.Error(1)
}
func (m *mockClassSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockClassSvc) AdminHardDeleteSession(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockClassSvc) AdminListSessions(ctx context.Context, q domain.AdminListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, q)
	ss, _ := a.Get(0).([]domain.ClassSession)
	return ss, a.Get(1).(int64), a.Error(2)
}

// newHandlerRouter wires the regular (non-admin) handler without auth / perm
// middleware. Service mocks accept any context — this lets us exercise
// handler binding and routing without a full auth fixture.
func newHandlerRouter(t *testing.T) (*gin.Engine, *mockClassSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()

	svc := &mockClassSvc{}
	h := classes.NewHandler(svc)

	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1, noop, perm)
	return r, svc
}

func newAdminRouter(t *testing.T) (*gin.Engine, *mockClassSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockClassSvc{}
	h := classes.NewAdminHandler(svc)
	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	grp := r.Group("/admin")
	h.RegisterAdminRoutes(grp)
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

func TestHandler_List_Success(t *testing.T) {
	r, svc := newHandlerRouter(t)
	svc.On("List", mock.Anything, mock.AnythingOfType("domain.ListParams")).
		Return([]domain.Class{{ID: uuid.New(), Name: "c1"}}, int64(1), nil)

	w := do(t, r, "GET", "/api/v1/classes?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_List_Forbidden_Maps403(t *testing.T) {
	r, svc := newHandlerRouter(t)
	svc.On("List", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/api/v1/classes", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_Create_Success(t *testing.T) {
	r, svc := newHandlerRouter(t)
	svc.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateClassDTO")).
		Return(&domain.Class{ID: uuid.New(), Name: "Algebra"}, nil)

	body := map[string]any{"name": "Algebra", "description": "101", "total_users": 30}
	w := do(t, r, "POST", "/api/v1/classes", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_Create_ValidationError_Maps400(t *testing.T) {
	r, svc := newHandlerRouter(t)
	// name missing; total_users negative.
	w := do(t, r, "POST", "/api/v1/classes", map[string]any{"total_users": -5})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestHandler_Get_InvalidUUID_Maps400(t *testing.T) {
	r, _ := newHandlerRouter(t)
	w := do(t, r, "GET", "/api/v1/classes/not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Update_NotFound_Maps404(t *testing.T) {
	r, svc := newHandlerRouter(t)
	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.Anything).
		Return((*domain.Class)(nil), domain.ErrNotFound)

	w := do(t, r, "PUT", "/api/v1/classes/"+id.String(), map[string]any{"name": "New"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_Delete_Success(t *testing.T) {
	r, svc := newHandlerRouter(t)
	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/api/v1/classes/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_CreateSession_Success(t *testing.T) {
	r, svc := newHandlerRouter(t)
	classID := uuid.New()
	svc.On("CreateSession", mock.Anything, classID, mock.AnythingOfType("domain.CreateClassSessionDTO")).
		Return(&domain.ClassSession{ID: uuid.New(), ClassID: classID}, nil)

	body := map[string]any{
		"name":       "Kickoff",
		"start_time": "2026-05-01T10:00:00Z",
	}
	w := do(t, r, "POST", "/api/v1/classes/"+classID.String()+"/sessions", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_Enroll_Conflict_Maps409(t *testing.T) {
	r, svc := newHandlerRouter(t)
	classID := uuid.New()
	userID := uuid.New()
	svc.On("Enroll", mock.Anything, classID, domain.EnrollClassMemberDTO{UserID: userID}).
		Return((*domain.ClassMember)(nil), domain.ErrConflict)

	w := do(t, r, "POST", "/api/v1/classes/"+classID.String()+"/members", map[string]any{"user_id": userID})
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_Leave_Success(t *testing.T) {
	r, svc := newHandlerRouter(t)
	classID := uuid.New()
	userID := uuid.New()
	svc.On("Leave", mock.Anything, classID, userID).Return(nil)

	w := do(t, r, "DELETE", "/api/v1/classes/"+classID.String()+"/members/"+userID.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_ListMembers_Forbidden_Maps403(t *testing.T) {
	r, svc := newHandlerRouter(t)
	classID := uuid.New()
	svc.On("ListMembers", mock.Anything, classID, mock.AnythingOfType("domain.ListClassMembersQuery")).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/api/v1/classes/"+classID.String()+"/members", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminHandler_List_Success(t *testing.T) {
	r, svc := newAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.AnythingOfType("domain.AdminListClassesQuery")).
		Return([]domain.Class{{ID: uuid.New()}}, int64(1), nil)

	w := do(t, r, "GET", "/admin/classes?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_List_Forbidden_Maps403(t *testing.T) {
	r, svc := newAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/admin/classes", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminHandler_HardDelete_Success(t *testing.T) {
	r, svc := newAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/admin/classes/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_HardDeleteSession_NotFound_Maps404(t *testing.T) {
	r, svc := newAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDeleteSession", mock.Anything, id).Return(domain.ErrNotFound)

	w := do(t, r, "DELETE", "/admin/classes/sessions/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
