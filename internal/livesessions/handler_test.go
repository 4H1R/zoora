package livesessions_test

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
	"github.com/4H1R/zoora/internal/livesessions"
	"github.com/4H1R/zoora/internal/middleware"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type mockLiveSessionSvc struct{ mock.Mock }

func (m *mockLiveSessionSvc) CreateRoom(ctx context.Context, dto domain.CreateLiveRoomDTO) (*domain.LiveRoom, error) {
	a := m.Called(ctx, dto)
	r, _ := a.Get(0).(*domain.LiveRoom)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) GetRoom(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.LiveRoom)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) JoinRoom(ctx context.Context, id uuid.UUID) (*domain.JoinLiveRoomResponse, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.JoinLiveRoomResponse)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) LeaveRoom(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockLiveSessionSvc) StartRoom(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.LiveRoom)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) EndRoom(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.LiveRoom)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) UpdateRoomConfig(ctx context.Context, id uuid.UUID, dto domain.UpdateLiveRoomConfigDTO) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id, dto)
	r, _ := a.Get(0).(*domain.LiveRoom)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) Heartbeat(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockLiveSessionSvc) List(ctx context.Context, q domain.ListLiveRoomsQuery) ([]domain.LiveRoom, int64, error) {
	a := m.Called(ctx, q)
	rooms, _ := a.Get(0).([]domain.LiveRoom)
	return rooms, a.Get(1).(int64), a.Error(2)
}
func (m *mockLiveSessionSvc) StartRecording(ctx context.Context, id uuid.UUID) (*domain.LiveRecording, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.LiveRecording)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) StopRecording(ctx context.Context, id uuid.UUID) (*domain.LiveRecording, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.LiveRecording)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) ListRecordings(ctx context.Context, id uuid.UUID, q domain.ListLiveRecordingsQuery) ([]domain.LiveRecording, int64, error) {
	a := m.Called(ctx, id, q)
	recs, _ := a.Get(0).([]domain.LiveRecording)
	return recs, a.Get(1).(int64), a.Error(2)
}
func (m *mockLiveSessionSvc) ListParticipants(ctx context.Context, id uuid.UUID, q domain.ListLiveParticipantsQuery) ([]domain.LiveParticipant, int64, error) {
	a := m.Called(ctx, id, q)
	ps, _ := a.Get(0).([]domain.LiveParticipant)
	return ps, a.Get(1).(int64), a.Error(2)
}
func (m *mockLiveSessionSvc) AdminList(ctx context.Context, q domain.AdminListLiveRoomsQuery) ([]domain.LiveRoom, int64, error) {
	a := m.Called(ctx, q)
	rooms, _ := a.Get(0).([]domain.LiveRoom)
	return rooms, a.Get(1).(int64), a.Error(2)
}
func (m *mockLiveSessionSvc) AdminEndRoom(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.LiveRoom)
	return r, a.Error(1)
}
func (m *mockLiveSessionSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockLiveSessionSvc) AutoCloseStaleRooms(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *mockLiveSessionSvc) OnLiveKitEvent(ctx context.Context, eventType, livekitRoomName, participantIdentity string) error {
	return m.Called(ctx, eventType, livekitRoomName, participantIdentity).Error(0)
}
func (m *mockLiveSessionSvc) OnEgressEnded(ctx context.Context, result domain.EgressResult) error {
	return m.Called(ctx, result).Error(0)
}
func (m *mockLiveSessionSvc) CloseRoomIfNoHost(ctx context.Context, roomID uuid.UUID) error {
	return m.Called(ctx, roomID).Error(0)
}
func (m *mockLiveSessionSvc) SetParticipantRole(ctx context.Context, roomID uuid.UUID, identity string, dto domain.SetParticipantRoleDTO) (*domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID, identity, dto)
	p, _ := a.Get(0).(*domain.LiveParticipant)
	return p, a.Error(1)
}
func (m *mockLiveSessionSvc) MuteParticipant(ctx context.Context, roomID uuid.UUID, identity string, dto domain.MuteParticipantDTO) error {
	return m.Called(ctx, roomID, identity, dto).Error(0)
}
func (m *mockLiveSessionSvc) SetHand(ctx context.Context, roomID uuid.UUID, dto domain.SetHandDTO) (*domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID, dto)
	p, _ := a.Get(0).(*domain.LiveParticipant)
	return p, a.Error(1)
}
func (m *mockLiveSessionSvc) SetParticipantHand(ctx context.Context, roomID uuid.UUID, identity string, dto domain.SetHandDTO) (*domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID, identity, dto)
	p, _ := a.Get(0).(*domain.LiveParticipant)
	return p, a.Error(1)
}
func (m *mockLiveSessionSvc) GetWhiteboard(ctx context.Context, roomID uuid.UUID) (*domain.LiveWhiteboard, error) {
	a := m.Called(ctx, roomID)
	wb, _ := a.Get(0).(*domain.LiveWhiteboard)
	return wb, a.Error(1)
}
func (m *mockLiveSessionSvc) SaveWhiteboard(ctx context.Context, roomID uuid.UUID, dto domain.SaveWhiteboardDTO) (*domain.LiveWhiteboard, error) {
	a := m.Called(ctx, roomID, dto)
	wb, _ := a.Get(0).(*domain.LiveWhiteboard)
	return wb, a.Error(1)
}

func newLiveRouter(t *testing.T) (*gin.Engine, *mockLiveSessionSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockLiveSessionSvc{}
	h := livesessions.NewHandler(svc)
	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1, noop, perm)
	return r, svc
}

func newAdminLiveRouter(t *testing.T) (*gin.Engine, *mockLiveSessionSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockLiveSessionSvc{}
	h := livesessions.NewAdminHandler(svc)
	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	grp := r.Group("/admin")
	h.RegisterAdminRoutes(grp)
	return r, svc
}

func doReq(t *testing.T, r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
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

func TestHandler_CreateRoom_Success(t *testing.T) {
	r, svc := newLiveRouter(t)
	svc.On("CreateRoom", mock.Anything, mock.AnythingOfType("domain.CreateLiveRoomDTO")).
		Return(&domain.LiveRoom{ID: uuid.New(), Status: domain.LiveRoomStatusCreated}, nil)

	body := map[string]any{"class_session_id": uuid.New().String(), "name": "Morning session"}
	w := doReq(t, r, "POST", "/api/v1/live-rooms", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_CreateRoom_MissingField_400(t *testing.T) {
	r, svc := newLiveRouter(t)
	w := doReq(t, r, "POST", "/api/v1/live-rooms", map[string]any{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "CreateRoom")
}

func TestHandler_List_Success(t *testing.T) {
	r, svc := newLiveRouter(t)
	svc.On("List", mock.Anything, mock.AnythingOfType("domain.ListLiveRoomsQuery")).
		Return([]domain.LiveRoom{{ID: uuid.New()}}, int64(1), nil)

	w := doReq(t, r, "GET", "/api/v1/live-rooms", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_List_WhitelistsScheduledStartAndName(t *testing.T) {
	r, svc := newLiveRouter(t)
	var captured domain.ListLiveRoomsQuery
	svc.On("List", mock.Anything, mock.MatchedBy(func(q domain.ListLiveRoomsQuery) bool {
		captured = q
		return true
	})).Return([]domain.LiveRoom{}, int64(0), nil)

	w := doReq(t, r, "GET", "/api/v1/live-rooms?order_by=scheduled_start_time&order_dir=asc&search=algebra", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "scheduled_start_time", captured.ListParams.OrderBy,
		"scheduled_start_time must be an allowed order field")
	assert.Equal(t, "asc", captured.ListParams.OrderDir)
	assert.Contains(t, captured.ListParams.SearchFields, "name",
		"name must be an allowed search field")
}

func TestHandler_GetRoom_InvalidUUID_400(t *testing.T) {
	r, _ := newLiveRouter(t)
	w := doReq(t, r, "GET", "/api/v1/live-rooms/not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetRoom_NotFound_404(t *testing.T) {
	r, svc := newLiveRouter(t)
	id := uuid.New()
	svc.On("GetRoom", mock.Anything, id).Return((*domain.LiveRoom)(nil), domain.ErrNotFound)

	w := doReq(t, r, "GET", "/api/v1/live-rooms/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_JoinRoom_Forbidden_403(t *testing.T) {
	r, svc := newLiveRouter(t)
	id := uuid.New()
	svc.On("JoinRoom", mock.Anything, id).Return((*domain.JoinLiveRoomResponse)(nil), domain.ErrForbidden)

	w := doReq(t, r, "POST", "/api/v1/live-rooms/"+id.String()+"/join", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_Heartbeat_Success(t *testing.T) {
	r, svc := newLiveRouter(t)
	id := uuid.New()
	svc.On("Heartbeat", mock.Anything, id).Return(nil)

	w := doReq(t, r, "POST", "/api/v1/live-rooms/"+id.String()+"/heartbeat", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_StartRecording_Success(t *testing.T) {
	r, svc := newLiveRouter(t)
	id := uuid.New()
	svc.On("StartRecording", mock.Anything, id).
		Return(&domain.LiveRecording{ID: uuid.New()}, nil)

	w := doReq(t, r, "POST", "/api/v1/live-rooms/"+id.String()+"/recordings", nil)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_StopRecording_PostStopRoute(t *testing.T) {
	r, svc := newLiveRouter(t)
	roomID := uuid.New()
	recID := uuid.New()
	svc.On("StopRecording", mock.Anything, recID).
		Return(&domain.LiveRecording{ID: recID, Status: domain.LiveRecordingStatusCompleted}, nil)

	w := doReq(t, r, "POST",
		"/api/v1/live-rooms/"+roomID.String()+"/recordings/"+recID.String()+"/stop", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_SetParticipantHand_Success(t *testing.T) {
	r, svc := newLiveRouter(t)
	id := uuid.New()
	svc.On("SetParticipantHand", mock.Anything, id, "user-9", domain.SetHandDTO{Raised: false}).
		Return(&domain.LiveParticipant{Identity: "user-9"}, nil)

	w := doReq(t, r, "PUT",
		"/api/v1/live-rooms/"+id.String()+"/participants/user-9/hand",
		map[string]any{"raised": false})

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_SetParticipantHand_Forbidden_403(t *testing.T) {
	r, svc := newLiveRouter(t)
	id := uuid.New()
	svc.On("SetParticipantHand", mock.Anything, id, "user-9", domain.SetHandDTO{Raised: false}).
		Return((*domain.LiveParticipant)(nil), domain.ErrForbidden)

	w := doReq(t, r, "PUT",
		"/api/v1/live-rooms/"+id.String()+"/participants/user-9/hand",
		map[string]any{"raised": false})

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminHandler_List_Success(t *testing.T) {
	r, svc := newAdminLiveRouter(t)
	svc.On("AdminList", mock.Anything, mock.AnythingOfType("domain.AdminListLiveRoomsQuery")).
		Return([]domain.LiveRoom{{ID: uuid.New()}}, int64(1), nil)

	w := doReq(t, r, "GET", "/admin/live-rooms?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_EndRoom_Success(t *testing.T) {
	r, svc := newAdminLiveRouter(t)
	id := uuid.New()
	svc.On("AdminEndRoom", mock.Anything, id).
		Return(&domain.LiveRoom{ID: id, Status: domain.LiveRoomStatusFinished}, nil)

	w := doReq(t, r, "POST", "/admin/live-rooms/"+id.String()+"/end", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_HardDelete_Success(t *testing.T) {
	r, svc := newAdminLiveRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(nil)

	w := doReq(t, r, "DELETE", "/admin/live-rooms/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_HardDelete_NotFound_404(t *testing.T) {
	r, svc := newAdminLiveRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(domain.ErrNotFound)

	w := doReq(t, r, "DELETE", "/admin/live-rooms/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
