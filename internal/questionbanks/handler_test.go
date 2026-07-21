package questionbanks_test

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
	"github.com/4H1R/zoora/internal/questionbanks"
)

type mockBankSvc struct{ mock.Mock }

func (m *mockBankSvc) Create(ctx context.Context, dto domain.CreateQuestionBankDTO) (*domain.QuestionBank, error) {
	a := m.Called(ctx, dto)
	b, _ := a.Get(0).(*domain.QuestionBank)
	return b, a.Error(1)
}

func (m *mockBankSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.QuestionBank, error) {
	a := m.Called(ctx, id)
	b, _ := a.Get(0).(*domain.QuestionBank)
	return b, a.Error(1)
}

func (m *mockBankSvc) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateQuestionBankDTO) (*domain.QuestionBank, error) {
	a := m.Called(ctx, id, dto)
	b, _ := a.Get(0).(*domain.QuestionBank)
	return b, a.Error(1)
}

func (m *mockBankSvc) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockBankSvc) List(ctx context.Context, p domain.ListParams) ([]domain.QuestionBank, int64, error) {
	a := m.Called(ctx, p)
	bs, _ := a.Get(0).([]domain.QuestionBank)
	return bs, a.Get(1).(int64), a.Error(2)
}

func (m *mockBankSvc) CreateQuestion(ctx context.Context, bankID uuid.UUID, dto domain.CreateQuestionDTO) (*domain.Question, error) {
	a := m.Called(ctx, bankID, dto)
	q, _ := a.Get(0).(*domain.Question)
	return q, a.Error(1)
}

func (m *mockBankSvc) GetQuestion(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	a := m.Called(ctx, id)
	q, _ := a.Get(0).(*domain.Question)
	return q, a.Error(1)
}

func (m *mockBankSvc) UpdateQuestion(ctx context.Context, id uuid.UUID, dto domain.UpdateQuestionDTO) (*domain.Question, error) {
	a := m.Called(ctx, id, dto)
	q, _ := a.Get(0).(*domain.Question)
	return q, a.Error(1)
}

func (m *mockBankSvc) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockBankSvc) ListQuestions(ctx context.Context, bankID uuid.UUID, q domain.ListQuestionsQuery) ([]domain.Question, int64, error) {
	a := m.Called(ctx, bankID, q)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Get(1).(int64), a.Error(2)
}

func (m *mockBankSvc) GenerateShareCode(ctx context.Context, bankID uuid.UUID, dto domain.GenerateShareCodeDTO) (*domain.QuestionBankShareCode, error) {
	a := m.Called(ctx, bankID, dto)
	c, _ := a.Get(0).(*domain.QuestionBankShareCode)
	return c, a.Error(1)
}

func (m *mockBankSvc) GetShareCode(ctx context.Context, bankID uuid.UUID) (*domain.QuestionBankShareCode, error) {
	a := m.Called(ctx, bankID)
	c, _ := a.Get(0).(*domain.QuestionBankShareCode)
	return c, a.Error(1)
}

func (m *mockBankSvc) RevokeShareCode(ctx context.Context, bankID uuid.UUID) error {
	return m.Called(ctx, bankID).Error(0)
}

func (m *mockBankSvc) PreviewShareCode(ctx context.Context, code string) (*domain.ShareCodePreview, error) {
	a := m.Called(ctx, code)
	p, _ := a.Get(0).(*domain.ShareCodePreview)
	return p, a.Error(1)
}

func (m *mockBankSvc) RedeemShareCode(ctx context.Context, dto domain.RedeemShareCodeDTO) (*domain.QuestionBank, error) {
	a := m.Called(ctx, dto)
	b, _ := a.Get(0).(*domain.QuestionBank)
	return b, a.Error(1)
}

func (m *mockBankSvc) AdminList(ctx context.Context, q domain.AdminListQuestionBanksQuery) ([]domain.QuestionBank, int64, error) {
	a := m.Called(ctx, q)
	bs, _ := a.Get(0).([]domain.QuestionBank)
	return bs, a.Get(1).(int64), a.Error(2)
}

func (m *mockBankSvc) AdminCreate(ctx context.Context, dto domain.AdminCreateQuestionBankDTO) (*domain.QuestionBank, error) {
	a := m.Called(ctx, dto)
	b, _ := a.Get(0).(*domain.QuestionBank)
	return b, a.Error(1)
}

func (m *mockBankSvc) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateQuestionBankDTO) (*domain.QuestionBank, error) {
	a := m.Called(ctx, id, dto)
	b, _ := a.Get(0).(*domain.QuestionBank)
	return b, a.Error(1)
}

func (m *mockBankSvc) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockBankSvc) AdminHardDeleteQuestion(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockBankSvc) AdminListQuestions(ctx context.Context, q domain.AdminListQuestionsQuery) ([]domain.Question, int64, error) {
	a := m.Called(ctx, q)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Get(1).(int64), a.Error(2)
}

func newBankHandlerRouter(t *testing.T) (*gin.Engine, *mockBankSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockBankSvc{}
	h := questionbanks.NewHandler(svc)
	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1, noop, perm)
	return r, svc
}

func newBankAdminRouter(t *testing.T) (*gin.Engine, *mockBankSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()
	svc := &mockBankSvc{}
	h := questionbanks.NewAdminHandler(svc)
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

func TestBankHandler_List_Success(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	svc.On("List", mock.Anything, mock.AnythingOfType("domain.ListParams")).
		Return([]domain.QuestionBank{{ID: uuid.New(), Name: "bank1"}}, int64(1), nil)

	w := doReq(t, r, "GET", "/api/v1/question-banks?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestBankHandler_List_Forbidden_Maps403(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	svc.On("List", mock.Anything, mock.Anything).
		Return(nil, int64(0), domain.ErrForbidden)

	w := doReq(t, r, "GET", "/api/v1/question-banks", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestBankHandler_Create_Success(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	svc.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateQuestionBankDTO")).
		Return(&domain.QuestionBank{ID: uuid.New(), Name: "Physics"}, nil)

	body := map[string]any{"name": "Physics", "description": "Physics questions"}
	w := doReq(t, r, "POST", "/api/v1/question-banks", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBankHandler_Create_ValidationError_Maps400(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	w := doReq(t, r, "POST", "/api/v1/question-banks", map[string]any{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestBankHandler_Get_InvalidUUID_Maps400(t *testing.T) {
	r, _ := newBankHandlerRouter(t)
	w := doReq(t, r, "GET", "/api/v1/question-banks/not-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBankHandler_Update_NotFound_Maps404(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.Anything).
		Return((*domain.QuestionBank)(nil), domain.ErrNotFound)

	w := doReq(t, r, "PUT", "/api/v1/question-banks/"+id.String(), map[string]any{"name": "New"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBankHandler_Delete_Success(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(nil)

	w := doReq(t, r, "DELETE", "/api/v1/question-banks/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBankHandler_CreateQuestion_Success(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	bankID := uuid.New()
	svc.On("CreateQuestion", mock.Anything, bankID, mock.AnythingOfType("domain.CreateQuestionDTO")).
		Return(&domain.Question{ID: uuid.New(), BankID: bankID, Type: domain.QuestionTypeChoice}, nil)

	body := map[string]any{
		"text":    "What is 2+2?",
		"type":    "choice",
		"options": []map[string]any{{"id": "a", "value": "4", "score": 1}},
	}
	w := doReq(t, r, "POST", "/api/v1/question-banks/"+bankID.String()+"/questions", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBankHandler_CreateQuestion_InvalidType_Maps400(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	bankID := uuid.New()
	body := map[string]any{"text": "Q", "type": "bogus"}
	w := doReq(t, r, "POST", "/api/v1/question-banks/"+bankID.String()+"/questions", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "CreateQuestion")
}

func TestBankHandler_ListQuestions_Success(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	bankID := uuid.New()
	svc.On("ListQuestions", mock.Anything, bankID, mock.AnythingOfType("domain.ListQuestionsQuery")).
		Return([]domain.Question{{ID: uuid.New()}}, int64(1), nil)

	w := doReq(t, r, "GET", "/api/v1/question-banks/"+bankID.String()+"/questions", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBankHandler_DeleteQuestion_Success(t *testing.T) {
	r, svc := newBankHandlerRouter(t)
	qID := uuid.New()
	svc.On("DeleteQuestion", mock.Anything, qID).Return(nil)

	w := doReq(t, r, "DELETE", "/api/v1/question-banks/questions/"+qID.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBankAdminHandler_List_Success(t *testing.T) {
	r, svc := newBankAdminRouter(t)
	svc.On("AdminList", mock.Anything, mock.AnythingOfType("domain.AdminListQuestionBanksQuery")).
		Return([]domain.QuestionBank{{ID: uuid.New()}}, int64(1), nil)

	w := doReq(t, r, "GET", "/admin/question-banks?page=1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBankAdminHandler_HardDelete_Success(t *testing.T) {
	r, svc := newBankAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDelete", mock.Anything, id).Return(nil)

	w := doReq(t, r, "DELETE", "/admin/question-banks/"+id.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBankAdminHandler_HardDeleteQuestion_NotFound_Maps404(t *testing.T) {
	r, svc := newBankAdminRouter(t)
	id := uuid.New()
	svc.On("AdminHardDeleteQuestion", mock.Anything, id).Return(domain.ErrNotFound)

	w := doReq(t, r, "DELETE", "/admin/question-banks/questions/"+id.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
