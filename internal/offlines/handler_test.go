package offlines_test

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
	"github.com/4H1R/zoora/internal/offlines"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type mockOfflineSvc struct{ mock.Mock }

func (m *mockOfflineSvc) CreateRoom(ctx context.Context, dto domain.CreateOfflineRoomDTO) (*domain.OfflineRoom, error) {
	a := m.Called(ctx, dto)
	r, _ := a.Get(0).(*domain.OfflineRoom)
	return r, a.Error(1)
}
func (m *mockOfflineSvc) GetRoom(ctx context.Context, id uuid.UUID) (*domain.OfflineRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.OfflineRoom)
	return r, a.Error(1)
}
func (m *mockOfflineSvc) UpdateRoom(ctx context.Context, id uuid.UUID, dto domain.UpdateOfflineRoomDTO) (*domain.OfflineRoom, error) {
	a := m.Called(ctx, id, dto)
	r, _ := a.Get(0).(*domain.OfflineRoom)
	return r, a.Error(1)
}
func (m *mockOfflineSvc) DeleteRoom(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockOfflineSvc) ListRooms(ctx context.Context, q domain.ListOfflineRoomsQuery) ([]domain.OfflineRoom, int64, error) {
	a := m.Called(ctx, q)
	rs, _ := a.Get(0).([]domain.OfflineRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mockOfflineSvc) AdminList(ctx context.Context, q domain.AdminListOfflineRoomsQuery) ([]domain.OfflineRoom, int64, error) {
	a := m.Called(ctx, q)
	rs, _ := a.Get(0).([]domain.OfflineRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mockOfflineSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func newOfflineRouter(t *testing.T) (*gin.Engine, *mockOfflineSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()

	svc := &mockOfflineSvc{}
	h := offlines.NewHandler(svc)

	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1, noop, perm)
	return r, svc
}

func newOfflineAdminRouter(t *testing.T) (*gin.Engine, *mockOfflineSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockOfflineSvc{}
	h := offlines.NewAdminHandler(svc)
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

// --- Room handler tests ---

func TestHandler_ListRooms_Success(t *testing.T) {
	r, svc := newOfflineRouter(t)
	svc.On("ListRooms", mock.Anything, mock.AnythingOfType("domain.ListOfflineRoomsQuery")).
		Return([]domain.OfflineRoom{{ID: uuid.New(), Title: "Lecture 1"}}, int64(1), nil)

	w := do(t, r, "GET", "/api/v1/offlines?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_ListRooms_Forbidden(t *testing.T) {
	r, svc := newOfflineRouter(t)
	svc.On("ListRooms", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/api/v1/offlines", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_CreateRoom_Success(t *testing.T) {
	r, svc := newOfflineRouter(t)
	svc.On("CreateRoom", mock.Anything, mock.AnythingOfType("domain.CreateOfflineRoomDTO")).
		Return(&domain.OfflineRoom{ID: uuid.New(), Title: "Lecture 1"}, nil)

	body := map[string]any{
		"class_session_id": uuid.New().String(),
		"title":            "Lecture 1",
	}
	w := do(t, r, "POST", "/api/v1/offlines", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_CreateRoom_MissingTitle_Maps400(t *testing.T) {
	r, svc := newOfflineRouter(t)
	body := map[string]any{
		"class_session_id": uuid.New().String(),
	}
	w := do(t, r, "POST", "/api/v1/offlines", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "CreateRoom")
}

func TestHandler_CreateRoom_MissingSessionID_Maps400(t *testing.T) {
	r, svc := newOfflineRouter(t)
	body := map[string]any{
		"title": "Lecture 1",
	}
	w := do(t, r, "POST", "/api/v1/offlines", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "CreateRoom")
}

func TestHandler_CreateRoom_Forbidden(t *testing.T) {
	r, svc := newOfflineRouter(t)
	svc.On("CreateRoom", mock.Anything, mock.Anything).
		Return((*domain.OfflineRoom)(nil), domain.ErrForbidden)

	body := map[string]any{
		"class_session_id": uuid.New().String(),
		"title":            "Lecture 1",
	}
	w := do(t, r, "POST", "/api/v1/offlines", body)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_GetRoom_Success(t *testing.T) {
	r, svc := newOfflineRouter(t)
	id := uuid.New()
	svc.On("GetRoom", mock.Anything, id).
		Return(&domain.OfflineRoom{ID: id, Title: "Lecture 1"}, nil)

	w := do(t, r, "GET", "/api/v1/offlines/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_GetRoom_InvalidUUID_Maps400(t *testing.T) {
	r, _ := newOfflineRouter(t)
	w := do(t, r, "GET", "/api/v1/offlines/not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetRoom_NotFound(t *testing.T) {
	r, svc := newOfflineRouter(t)
	id := uuid.New()
	svc.On("GetRoom", mock.Anything, id).
		Return((*domain.OfflineRoom)(nil), domain.ErrNotFound)

	w := do(t, r, "GET", "/api/v1/offlines/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_UpdateRoom_Success(t *testing.T) {
	r, svc := newOfflineRouter(t)
	id := uuid.New()
	svc.On("UpdateRoom", mock.Anything, id, mock.Anything).
		Return(&domain.OfflineRoom{ID: id, Title: "Updated"}, nil)

	w := do(t, r, "PUT", "/api/v1/offlines/"+id.String(), map[string]any{"title": "Updated"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_UpdateRoom_NotFound(t *testing.T) {
	r, svc := newOfflineRouter(t)
	id := uuid.New()
	svc.On("UpdateRoom", mock.Anything, id, mock.Anything).
		Return((*domain.OfflineRoom)(nil), domain.ErrNotFound)

	w := do(t, r, "PUT", "/api/v1/offlines/"+id.String(), map[string]any{"title": "Updated"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_DeleteRoom_Success(t *testing.T) {
	r, svc := newOfflineRouter(t)
	id := uuid.New()
	svc.On("DeleteRoom", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/api/v1/offlines/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_DeleteRoom_Forbidden(t *testing.T) {
	r, svc := newOfflineRouter(t)
	id := uuid.New()
	svc.On("DeleteRoom", mock.Anything, id).Return(domain.ErrForbidden)

	w := do(t, r, "DELETE", "/api/v1/offlines/"+id.String(), nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_DeleteRoom_NotFound(t *testing.T) {
	r, svc := newOfflineRouter(t)
	id := uuid.New()
	svc.On("DeleteRoom", mock.Anything, id).Return(domain.ErrNotFound)

	w := do(t, r, "DELETE", "/api/v1/offlines/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Admin handler tests ---

func TestAdminHandler_List_Success(t *testing.T) {
	r, svc := newOfflineAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.AnythingOfType("domain.AdminListOfflineRoomsQuery")).
		Return([]domain.OfflineRoom{{ID: uuid.New()}}, int64(1), nil)

	w := do(t, r, "GET", "/admin/offlines?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_List_Forbidden(t *testing.T) {
	r, svc := newOfflineAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/admin/offlines", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminHandler_HardDelete_Success(t *testing.T) {
	r, svc := newOfflineAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/admin/offlines/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_HardDelete_NotFound(t *testing.T) {
	r, svc := newOfflineAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(domain.ErrNotFound)

	w := do(t, r, "DELETE", "/admin/offlines/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
