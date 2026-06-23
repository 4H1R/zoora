package quizzes_test

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
	"github.com/4H1R/zoora/internal/quizzes"
)

type mockQuizSvc struct{ mock.Mock }

func (m *mockQuizSvc) Create(ctx context.Context, dto domain.CreateQuizDTO) (*domain.Quiz, error) {
	a := m.Called(ctx, dto)
	q, _ := a.Get(0).(*domain.Quiz)
	return q, a.Error(1)
}
func (m *mockQuizSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	a := m.Called(ctx, id)
	q, _ := a.Get(0).(*domain.Quiz)
	return q, a.Error(1)
}
func (m *mockQuizSvc) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateQuizDTO) (*domain.Quiz, error) {
	a := m.Called(ctx, id, dto)
	q, _ := a.Get(0).(*domain.Quiz)
	return q, a.Error(1)
}
func (m *mockQuizSvc) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockQuizSvc) List(ctx context.Context, q domain.ListQuizzesQuery) ([]domain.Quiz, int64, error) {
	a := m.Called(ctx, q)
	qs, _ := a.Get(0).([]domain.Quiz)
	return qs, a.Get(1).(int64), a.Error(2)
}
func (m *mockQuizSvc) ListMine(ctx context.Context, p domain.ListParams) ([]domain.MyExam, int64, error) {
	a := m.Called(ctx, p)
	es, _ := a.Get(0).([]domain.MyExam)
	return es, a.Get(1).(int64), a.Error(2)
}
func (m *mockQuizSvc) CreateRule(ctx context.Context, quizID uuid.UUID, dto domain.CreateQuizRuleDTO) (*domain.QuizRule, error) {
	a := m.Called(ctx, quizID, dto)
	r, _ := a.Get(0).(*domain.QuizRule)
	return r, a.Error(1)
}
func (m *mockQuizSvc) GetRule(ctx context.Context, id uuid.UUID) (*domain.QuizRule, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.QuizRule)
	return r, a.Error(1)
}
func (m *mockQuizSvc) UpdateRule(ctx context.Context, id uuid.UUID, dto domain.UpdateQuizRuleDTO) (*domain.QuizRule, error) {
	a := m.Called(ctx, id, dto)
	r, _ := a.Get(0).(*domain.QuizRule)
	return r, a.Error(1)
}
func (m *mockQuizSvc) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockQuizSvc) ListRules(ctx context.Context, quizID uuid.UUID, q domain.ListQuizRulesQuery) ([]domain.QuizRule, int64, error) {
	a := m.Called(ctx, quizID, q)
	rs, _ := a.Get(0).([]domain.QuizRule)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mockQuizSvc) CreateRoom(ctx context.Context, quizID uuid.UUID, dto domain.CreateQuizRoomDTO) (*domain.QuizRoom, error) {
	a := m.Called(ctx, quizID, dto)
	r, _ := a.Get(0).(*domain.QuizRoom)
	return r, a.Error(1)
}
func (m *mockQuizSvc) GetRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.QuizRoom)
	return r, a.Error(1)
}
func (m *mockQuizSvc) StartRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.QuizRoom)
	return r, a.Error(1)
}
func (m *mockQuizSvc) EndRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	a := m.Called(ctx, id)
	r, _ := a.Get(0).(*domain.QuizRoom)
	return r, a.Error(1)
}
func (m *mockQuizSvc) ListRooms(ctx context.Context, quizID uuid.UUID, q domain.ListQuizRoomsQuery) ([]domain.QuizRoom, int64, error) {
	a := m.Called(ctx, quizID, q)
	rs, _ := a.Get(0).([]domain.QuizRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mockQuizSvc) ListQuestionsForTaking(ctx context.Context, quizID uuid.UUID) ([]domain.Question, error) {
	a := m.Called(ctx, quizID)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Error(1)
}
func (m *mockQuizSvc) StartSubmission(ctx context.Context, quizID uuid.UUID, dto domain.StartQuizSubmissionDTO) (*domain.QuizSubmission, error) {
	a := m.Called(ctx, quizID, dto)
	s, _ := a.Get(0).(*domain.QuizSubmission)
	return s, a.Error(1)
}
func (m *mockQuizSvc) SubmitQuiz(ctx context.Context, submissionID uuid.UUID, dto domain.SubmitQuizDTO) (*domain.QuizSubmission, error) {
	a := m.Called(ctx, submissionID, dto)
	s, _ := a.Get(0).(*domain.QuizSubmission)
	return s, a.Error(1)
}
func (m *mockQuizSvc) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.QuizSubmission, error) {
	a := m.Called(ctx, id)
	s, _ := a.Get(0).(*domain.QuizSubmission)
	return s, a.Error(1)
}
func (m *mockQuizSvc) ListSubmissions(ctx context.Context, quizID uuid.UUID, q domain.ListSubmissionsQuery) ([]domain.QuizSubmission, int64, error) {
	a := m.Called(ctx, quizID, q)
	ss, _ := a.Get(0).([]domain.QuizSubmission)
	return ss, a.Get(1).(int64), a.Error(2)
}
func (m *mockQuizSvc) GradeSubmission(ctx context.Context, id uuid.UUID, dto domain.GradeSubmissionDTO) (*domain.QuizSubmission, error) {
	a := m.Called(ctx, id, dto)
	s, _ := a.Get(0).(*domain.QuizSubmission)
	return s, a.Error(1)
}
func (m *mockQuizSvc) AdminList(ctx context.Context, q domain.AdminListQuizzesQuery) ([]domain.Quiz, int64, error) {
	a := m.Called(ctx, q)
	qs, _ := a.Get(0).([]domain.Quiz)
	return qs, a.Get(1).(int64), a.Error(2)
}
func (m *mockQuizSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func newQuizHandlerRouter(t *testing.T) (*gin.Engine, *mockQuizSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockQuizSvc{}
	h := quizzes.NewHandler(svc)
	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	permAny := func(...domain.PermissionName) gin.HandlerFunc { return noop }
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1, noop, perm, permAny)
	return r, svc
}

func newQuizAdminRouter(t *testing.T) (*gin.Engine, *mockQuizSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockQuizSvc{}
	h := quizzes.NewAdminHandler(svc)
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

func TestQuizHandler_List_Success(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	svc.On("List", mock.Anything, mock.AnythingOfType("domain.ListQuizzesQuery")).
		Return([]domain.Quiz{{ID: uuid.New(), Title: "q1"}}, int64(1), nil)

	w := do(t, r, "GET", "/api/v1/quizzes?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestQuizHandler_List_Forbidden_Maps403(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	svc.On("List", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := do(t, r, "GET", "/api/v1/quizzes", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestQuizHandler_Create_Success(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	classID := uuid.New()
	svc.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateQuizDTO")).
		Return(&domain.Quiz{ID: uuid.New(), Title: "Final"}, nil)

	body := map[string]any{
		"class_id":         classID,
		"title":            "Final",
		"description":      "Final exam",
		"duration_minutes": 60,
	}
	w := do(t, r, "POST", "/api/v1/quizzes", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestQuizHandler_Create_ValidationError_Maps400(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	w := do(t, r, "POST", "/api/v1/quizzes", map[string]any{"title": "x"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestQuizHandler_Get_InvalidUUID_Maps400(t *testing.T) {
	r, _ := newQuizHandlerRouter(t)
	w := do(t, r, "GET", "/api/v1/quizzes/not-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestQuizHandler_Update_NotFound_Maps404(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.Anything).
		Return((*domain.Quiz)(nil), domain.ErrNotFound)

	w := do(t, r, "PUT", "/api/v1/quizzes/"+id.String(), map[string]any{"title": "New"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestQuizHandler_Delete_Success(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/api/v1/quizzes/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQuizHandler_StartSubmission_Success(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	quizID := uuid.New()
	roomID := uuid.New()
	svc.On("StartSubmission", mock.Anything, quizID, mock.AnythingOfType("domain.StartQuizSubmissionDTO")).
		Return(&domain.QuizSubmission{ID: uuid.New(), QuizID: quizID}, nil)

	body := map[string]any{"quiz_room_id": roomID}
	w := do(t, r, "POST", "/api/v1/quizzes/"+quizID.String()+"/submissions", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestQuizHandler_StartSubmission_Conflict_Maps409(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	quizID := uuid.New()
	svc.On("StartSubmission", mock.Anything, quizID, mock.Anything).
		Return((*domain.QuizSubmission)(nil), domain.ErrConflict)

	body := map[string]any{"quiz_room_id": uuid.New()}
	w := do(t, r, "POST", "/api/v1/quizzes/"+quizID.String()+"/submissions", body)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestQuizHandler_SubmitQuiz_Success(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	subID := uuid.New()
	svc.On("SubmitQuiz", mock.Anything, subID, mock.AnythingOfType("domain.SubmitQuizDTO")).
		Return(&domain.QuizSubmission{ID: subID, Status: domain.SubmissionStatusSubmitted}, nil)

	body := map[string]any{
		"answers": []map[string]any{
			{"question_id": uuid.New(), "selected_option_ids": []string{"a"}, "spent_seconds": 30},
		},
	}
	w := do(t, r, "POST", "/api/v1/quizzes/submissions/"+subID.String()+"/submit", body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQuizHandler_GetSubmission_Success(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	subID := uuid.New()
	svc.On("GetSubmission", mock.Anything, subID).
		Return(&domain.QuizSubmission{ID: subID}, nil)

	w := do(t, r, "GET", "/api/v1/quizzes/submissions/"+subID.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQuizHandler_ListSubmissions_Success(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	quizID := uuid.New()
	svc.On("ListSubmissions", mock.Anything, quizID, mock.AnythingOfType("domain.ListSubmissionsQuery")).
		Return([]domain.QuizSubmission{{ID: uuid.New()}}, int64(1), nil)

	w := do(t, r, "GET", "/api/v1/quizzes/"+quizID.String()+"/submissions", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQuizHandler_GradeSubmission_Success(t *testing.T) {
	r, svc := newQuizHandlerRouter(t)
	subID := uuid.New()
	svc.On("GradeSubmission", mock.Anything, subID, mock.AnythingOfType("domain.GradeSubmissionDTO")).
		Return(&domain.QuizSubmission{ID: subID, Status: domain.SubmissionStatusGraded}, nil)

	body := map[string]any{
		"grades": []map[string]any{
			{"question_id": uuid.New(), "earned_score": 5.0},
		},
	}
	w := do(t, r, "POST", "/api/v1/quizzes/submissions/"+subID.String()+"/grade", body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQuizAdminHandler_List_Success(t *testing.T) {
	r, svc := newQuizAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.AnythingOfType("domain.AdminListQuizzesQuery")).
		Return([]domain.Quiz{{ID: uuid.New()}}, int64(1), nil)

	w := do(t, r, "GET", "/admin/quizzes?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQuizAdminHandler_HardDelete_Success(t *testing.T) {
	r, svc := newQuizAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(nil)

	w := do(t, r, "DELETE", "/admin/quizzes/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQuizAdminHandler_HardDelete_NotFound_Maps404(t *testing.T) {
	r, svc := newQuizAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(domain.ErrNotFound)

	w := do(t, r, "DELETE", "/admin/quizzes/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
