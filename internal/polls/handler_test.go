package polls_test

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
	"github.com/4H1R/zoora/internal/polls"
)

type mockPollSvc struct{ mock.Mock }

func (m *mockPollSvc) Create(ctx context.Context, dto domain.CreatePollDTO) (*domain.Poll, error) {
	a := m.Called(ctx, dto)
	poll, _ := a.Get(0).(*domain.Poll)
	return poll, a.Error(1)
}

func (m *mockPollSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.Poll, error) {
	a := m.Called(ctx, id)
	poll, _ := a.Get(0).(*domain.Poll)
	return poll, a.Error(1)
}

func (m *mockPollSvc) Update(ctx context.Context, id uuid.UUID, dto domain.UpdatePollDTO) (*domain.Poll, error) {
	a := m.Called(ctx, id, dto)
	poll, _ := a.Get(0).(*domain.Poll)
	return poll, a.Error(1)
}

func (m *mockPollSvc) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockPollSvc) List(ctx context.Context, q domain.ListPollsQuery) ([]domain.Poll, int64, error) {
	a := m.Called(ctx, q)
	items, _ := a.Get(0).([]domain.Poll)
	return items, a.Get(1).(int64), a.Error(2)
}

func (m *mockPollSvc) Answer(ctx context.Context, pollID uuid.UUID, dto domain.AnswerPollDTO) ([]domain.PollAnswer, error) {
	a := m.Called(ctx, pollID, dto)
	items, _ := a.Get(0).([]domain.PollAnswer)
	return items, a.Error(1)
}

func (m *mockPollSvc) ListAnswers(ctx context.Context, pollID uuid.UUID, q domain.ListPollAnswersQuery) ([]domain.PollAnswer, int64, error) {
	a := m.Called(ctx, pollID, q)
	items, _ := a.Get(0).([]domain.PollAnswer)
	return items, a.Get(1).(int64), a.Error(2)
}

func (m *mockPollSvc) Results(ctx context.Context, pollID uuid.UUID) (*domain.PollResults, error) {
	a := m.Called(ctx, pollID)
	r, _ := a.Get(0).(*domain.PollResults)
	return r, a.Error(1)
}

func (m *mockPollSvc) AdminList(ctx context.Context, q domain.AdminListPollsQuery) ([]domain.Poll, int64, error) {
	a := m.Called(ctx, q)
	items, _ := a.Get(0).([]domain.Poll)
	return items, a.Get(1).(int64), a.Error(2)
}

func (m *mockPollSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockPollSvc) CloseByModel(ctx context.Context, modelType string, modelID uuid.UUID) error {
	return m.Called(ctx, modelType, modelID).Error(0)
}

func newPollRouter(t *testing.T) (*gin.Engine, *mockPollSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockPollSvc{}
	h := polls.NewHandler(svc)
	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	h.RegisterRoutes(r.Group("/api/v1"), noop, perm)
	return r, svc
}

func newPollAdminRouter(t *testing.T) (*gin.Engine, *mockPollSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockPollSvc{}
	h := polls.NewAdminHandler(svc)
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

func TestHandlerListSuccessBindsFilters(t *testing.T) {
	r, svc := newPollRouter(t)
	modelID := uuid.New()
	svc.On("List", mock.Anything, mock.MatchedBy(func(q domain.ListPollsQuery) bool {
		return q.ModelType != nil &&
			*q.ModelType == "class" &&
			q.ModelID != nil &&
			*q.ModelID == modelID &&
			q.ListParams.Search == "quiz" &&
			q.ListParams.OrderBy == "name" &&
			q.ListParams.OrderDir == "asc"
	})).Return([]domain.Poll{{ID: uuid.New(), Name: "Quiz poll"}}, int64(1), nil)

	w := do(t, r, http.MethodGet, "/api/v1/polls?model_type=class&model_id="+modelID.String()+"&search=quiz&order_by=name&order_dir=asc", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListRejectsInvalidModelID(t *testing.T) {
	r, svc := newPollRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/polls?model_id=bad", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "List")
}

func TestHandlerCreateValidationDoesNotCallService(t *testing.T) {
	r, svc := newPollRouter(t)

	w := do(t, r, http.MethodPost, "/api/v1/polls", map[string]any{
		"model_type":            "class",
		"model_id":              uuid.New().String(),
		"name":                  "A",
		"allowed_answers_count": 0,
		"options":               []map[string]string{{"label": "Only", "value": "only"}},
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestHandlerGetInvalidUUIDMaps400(t *testing.T) {
	r, svc := newPollRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/polls/not-a-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "GetByID")
}

func TestHandlerAnswerSuccess(t *testing.T) {
	r, svc := newPollRouter(t)
	id := uuid.New()
	svc.On("Answer", mock.Anything, id, domain.AnswerPollDTO{Options: []string{"yes"}}).
		Return([]domain.PollAnswer{{ID: uuid.New(), PollID: id, Option: "yes"}}, nil)

	w := do(t, r, http.MethodPost, "/api/v1/polls/"+id.String()+"/answer", map[string]any{"options": []string{"yes"}})

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandlerListAnswersRejectsInvalidUserID(t *testing.T) {
	r, svc := newPollRouter(t)
	id := uuid.New()

	w := do(t, r, http.MethodGet, "/api/v1/polls/"+id.String()+"/answers?user_id=bad", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "ListAnswers")
}

func TestAdminHandlerListRejectsInvalidUUIDFilters(t *testing.T) {
	r, svc := newPollAdminRouter(t)

	w := do(t, r, http.MethodGet, "/admin/polls?user_id=bad", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "AdminList")
}

func TestAdminHandlerHardDeleteNotFound(t *testing.T) {
	r, svc := newPollAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(domain.ErrNotFound)

	w := do(t, r, http.MethodDelete, "/admin/polls/"+id.String(), nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
