package practices_test

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
	"github.com/4H1R/zoora/internal/practices"
)

type mockPracticeSvc struct{ mock.Mock }

func (m *mockPracticeSvc) CreateRoom(ctx context.Context, dto domain.CreatePracticeRoomDTO) (*domain.PracticeRoom, error) {
	a := m.Called(ctx, dto)
	r, _ := a.Get(0).(*domain.PracticeRoom)
	return r, a.Error(1)
}
func (m *mockPracticeSvc) GetRoom(ctx context.Context, id uuid.UUID) (*domain.PracticeRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.PracticeRoom)
	return r, a.Error(1)
}
func (m *mockPracticeSvc) UpdateRoom(ctx context.Context, id uuid.UUID, dto domain.UpdatePracticeRoomDTO) (*domain.PracticeRoom, error) {
	a := m.Called(ctx, id, dto)
	r, _ := a.Get(0).(*domain.PracticeRoom)
	return r, a.Error(1)
}
func (m *mockPracticeSvc) DeleteRoom(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockPracticeSvc) ListRooms(ctx context.Context, q domain.ListPracticeRoomsQuery) ([]domain.PracticeRoom, int64, error) {
	a := m.Called(ctx, q)
	rs, _ := a.Get(0).([]domain.PracticeRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mockPracticeSvc) Submit(ctx context.Context, roomID uuid.UUID, dto domain.CreatePracticeSubmissionDTO) (*domain.PracticeSubmission, error) {
	a := m.Called(ctx, roomID, dto)
	s, _ := a.Get(0).(*domain.PracticeSubmission)
	return s, a.Error(1)
}
func (m *mockPracticeSvc) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.PracticeSubmission, error) {
	a := m.Called(ctx, id)
	s, _ := a.Get(0).(*domain.PracticeSubmission)
	return s, a.Error(1)
}
func (m *mockPracticeSvc) ListSubmissions(ctx context.Context, roomID uuid.UUID, q domain.ListPracticeSubmissionsQuery) ([]domain.PracticeSubmission, int64, error) {
	a := m.Called(ctx, roomID, q)
	ss, _ := a.Get(0).([]domain.PracticeSubmission)
	return ss, a.Get(1).(int64), a.Error(2)
}
func (m *mockPracticeSvc) Grade(ctx context.Context, subID uuid.UUID, dto domain.GradePracticeSubmissionDTO) (*domain.PracticeSubmission, error) {
	a := m.Called(ctx, subID, dto)
	s, _ := a.Get(0).(*domain.PracticeSubmission)
	return s, a.Error(1)
}
func (m *mockPracticeSvc) AdminList(ctx context.Context, q domain.AdminListPracticeRoomsQuery) ([]domain.PracticeRoom, int64, error) {
	a := m.Called(ctx, q)
	rs, _ := a.Get(0).([]domain.PracticeRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mockPracticeSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func newPracticeRouter(t *testing.T) (*gin.Engine, *mockPracticeSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()

	svc := &mockPracticeSvc{}
	h := practices.NewHandler(svc)

	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1, noop, perm)
	return r, svc
}

func newPracticeAdminRouter(t *testing.T) (*gin.Engine, *mockPracticeSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockPracticeSvc{}
	h := practices.NewAdminHandler(svc)
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
	r, svc := newPracticeRouter(t)
	svc.On("ListRooms", mock.Anything, mock.AnythingOfType("domain.ListPracticeRoomsQuery")).
		Return([]domain.PracticeRoom{{ID: uuid.New(), Title: "HW1"}}, int64(1), nil)

	w := do(t, r, "GET", "/api/v1/practices?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_ListRooms_Forbidden(t *testing.T) {
	r, svc := newPracticeRouter(t)
	svc.On("ListRooms", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/api/v1/practices", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_CreateRoom_Success(t *testing.T) {
	r, svc := newPracticeRouter(t)
	svc.On("CreateRoom", mock.Anything, mock.AnythingOfType("domain.CreatePracticeRoomDTO")).
		Return(&domain.PracticeRoom{ID: uuid.New(), Title: "HW1"}, nil)

	body := map[string]any{
		"class_session_id": uuid.New().String(),
		"title":            "HW1",
		"max_score":        100,
		"start_time":       "2026-05-01T10:00:00Z",
		"end_time":         "2026-05-10T10:00:00Z",
	}
	w := do(t, r, "POST", "/api/v1/practices", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_CreateRoom_MissingTitle_Maps400(t *testing.T) {
	r, svc := newPracticeRouter(t)
	body := map[string]any{
		"class_session_id": uuid.New().String(),
		"start_time":       "2026-05-01T10:00:00Z",
		"end_time":         "2026-05-10T10:00:00Z",
	}
	w := do(t, r, "POST", "/api/v1/practices", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "CreateRoom")
}

func TestHandler_GetRoom_InvalidUUID_Maps400(t *testing.T) {
	r, _ := newPracticeRouter(t)
	w := do(t, r, "GET", "/api/v1/practices/not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetRoom_NotFound(t *testing.T) {
	r, svc := newPracticeRouter(t)
	id := uuid.New()
	svc.On("GetRoom", mock.Anything, id).
		Return((*domain.PracticeRoom)(nil), domain.ErrNotFound)

	w := do(t, r, "GET", "/api/v1/practices/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_UpdateRoom_NotFound(t *testing.T) {
	r, svc := newPracticeRouter(t)
	id := uuid.New()
	svc.On("UpdateRoom", mock.Anything, id, mock.Anything).
		Return((*domain.PracticeRoom)(nil), domain.ErrNotFound)

	w := do(t, r, "PUT", "/api/v1/practices/"+id.String(), map[string]any{"title": "Updated"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_DeleteRoom_Success(t *testing.T) {
	r, svc := newPracticeRouter(t)
	id := uuid.New()
	svc.On("DeleteRoom", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/api/v1/practices/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_DeleteRoom_Forbidden(t *testing.T) {
	r, svc := newPracticeRouter(t)
	id := uuid.New()
	svc.On("DeleteRoom", mock.Anything, id).Return(domain.ErrForbidden)

	w := do(t, r, "DELETE", "/api/v1/practices/"+id.String(), nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- Submission handler tests ---

func TestHandler_Submit_Success(t *testing.T) {
	r, svc := newPracticeRouter(t)
	roomID := uuid.New()
	svc.On("Submit", mock.Anything, roomID, mock.AnythingOfType("domain.CreatePracticeSubmissionDTO")).
		Return(&domain.PracticeSubmission{ID: uuid.New(), PracticeRoomID: roomID}, nil)

	body := map[string]any{"content": "my answer"}
	w := do(t, r, "POST", "/api/v1/practices/"+roomID.String()+"/submissions", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_Submit_Conflict(t *testing.T) {
	r, svc := newPracticeRouter(t)
	roomID := uuid.New()
	svc.On("Submit", mock.Anything, roomID, mock.Anything).
		Return((*domain.PracticeSubmission)(nil), domain.ErrConflict)

	w := do(t, r, "POST", "/api/v1/practices/"+roomID.String()+"/submissions", map[string]any{"content": "x"})
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_ListSubmissions_Success(t *testing.T) {
	r, svc := newPracticeRouter(t)
	roomID := uuid.New()
	svc.On("ListSubmissions", mock.Anything, roomID, mock.AnythingOfType("domain.ListPracticeSubmissionsQuery")).
		Return([]domain.PracticeSubmission{{ID: uuid.New()}}, int64(1), nil)

	w := do(t, r, "GET", "/api/v1/practices/"+roomID.String()+"/submissions", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_ListSubmissions_Forbidden(t *testing.T) {
	r, svc := newPracticeRouter(t)
	roomID := uuid.New()
	svc.On("ListSubmissions", mock.Anything, roomID, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/api/v1/practices/"+roomID.String()+"/submissions", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_GetSubmission_NotFound(t *testing.T) {
	r, svc := newPracticeRouter(t)
	id := uuid.New()
	svc.On("GetSubmission", mock.Anything, id).
		Return((*domain.PracticeSubmission)(nil), domain.ErrNotFound)

	w := do(t, r, "GET", "/api/v1/practices/submissions/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_Grade_Success(t *testing.T) {
	r, svc := newPracticeRouter(t)
	subID := uuid.New()
	score := 85.0
	svc.On("Grade", mock.Anything, subID, mock.AnythingOfType("domain.GradePracticeSubmissionDTO")).
		Return(&domain.PracticeSubmission{ID: subID, Score: &score}, nil)

	body := map[string]any{"score": 85, "teacher_comment": "Good work"}
	w := do(t, r, "PUT", "/api/v1/practices/submissions/"+subID.String()+"/grade", body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Grade_Forbidden(t *testing.T) {
	r, svc := newPracticeRouter(t)
	subID := uuid.New()
	svc.On("Grade", mock.Anything, subID, mock.Anything).
		Return((*domain.PracticeSubmission)(nil), domain.ErrForbidden)

	w := do(t, r, "PUT", "/api/v1/practices/submissions/"+subID.String()+"/grade", map[string]any{"score": 50})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- Admin handler tests ---

func TestAdminHandler_List_Success(t *testing.T) {
	r, svc := newPracticeAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.AnythingOfType("domain.AdminListPracticeRoomsQuery")).
		Return([]domain.PracticeRoom{{ID: uuid.New()}}, int64(1), nil)

	w := do(t, r, "GET", "/admin/practices?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_List_Forbidden(t *testing.T) {
	r, svc := newPracticeAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/admin/practices", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminHandler_HardDelete_Success(t *testing.T) {
	r, svc := newPracticeAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/admin/practices/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_HardDelete_NotFound(t *testing.T) {
	r, svc := newPracticeAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(domain.ErrNotFound)

	w := do(t, r, "DELETE", "/admin/practices/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
