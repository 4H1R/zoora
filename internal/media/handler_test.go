package media_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/media"
	"github.com/4H1R/zoora/internal/middleware"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type mockMediaSvc struct{ mock.Mock }

func (m *mockMediaSvc) PresignUpload(ctx context.Context, dto domain.PresignUploadDTO) (*domain.PresignUploadResponse, error) {
	a := m.Called(ctx, dto)
	resp, _ := a.Get(0).(*domain.PresignUploadResponse)
	return resp, a.Error(1)
}

func (m *mockMediaSvc) PresignDownload(ctx context.Context, id uuid.UUID, expiry time.Duration) (*domain.PresignDownloadResponse, error) {
	a := m.Called(ctx, id, expiry)
	resp, _ := a.Get(0).(*domain.PresignDownloadResponse)
	return resp, a.Error(1)
}

func (m *mockMediaSvc) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	a := m.Called(ctx, id)
	item, _ := a.Get(0).(*domain.Media)
	return item, a.Error(1)
}

func (m *mockMediaSvc) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockMediaSvc) ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]domain.Media, error) {
	a := m.Called(ctx, modelType, modelID, collection)
	items, _ := a.Get(0).([]domain.Media)
	return items, a.Error(1)
}

func (m *mockMediaSvc) CleanupByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) error {
	return m.Called(ctx, modelType, modelID, collection).Error(0)
}

func newMediaRouter(t *testing.T) (*gin.Engine, *mockMediaSvc) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_ = httpx.RegisterValidators()

	svc := &mockMediaSvc{}
	h := media.NewHandler(svc)

	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.Default()))
	noop := func(c *gin.Context) { c.Next() }
	perm := func(domain.PermissionName) gin.HandlerFunc { return noop }
	h.RegisterRoutes(r.Group("/api/v1"), noop, perm)
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

func TestHandlerPresignUploadSuccess(t *testing.T) {
	r, svc := newMediaRouter(t)
	modelID := uuid.New()
	mediaID := uuid.New()
	svc.On("PresignUpload", mock.Anything, mock.AnythingOfType("domain.PresignUploadDTO")).
		Return(&domain.PresignUploadResponse{
			UploadURL: "https://upload.example.test",
			Key:       "practices/" + modelID.String() + "/attachments/file.pdf",
			Media:     &domain.Media{ID: mediaID, FileName: "file.pdf"},
		}, nil)

	w := do(t, r, http.MethodPost, "/api/v1/media/presign", map[string]any{
		"model_type":      "practices",
		"model_id":        modelID.String(),
		"collection_name": "attachments",
		"file_name":       "file.pdf",
		"mime_type":       "application/pdf",
		"size":            10,
	})

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "https://upload.example.test")
	svc.AssertExpectations(t)
}

func TestHandlerPresignUploadValidationErrorDoesNotCallService(t *testing.T) {
	r, svc := newMediaRouter(t)

	w := do(t, r, http.MethodPost, "/api/v1/media/presign", map[string]any{
		"model_type": "practices",
		"model_id":   "not-a-uuid",
		"mime_type":  "application/pdf",
		"size":       0,
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "PresignUpload")
}

func TestHandlerGetInvalidUUIDMaps400(t *testing.T) {
	r, svc := newMediaRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/media/not-a-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "GetByID")
}

func TestHandlerGetNotFoundMaps404(t *testing.T) {
	r, svc := newMediaRouter(t)
	id := uuid.New()
	svc.On("GetByID", mock.Anything, id).Return((*domain.Media)(nil), domain.ErrNotFound)

	w := do(t, r, http.MethodGet, "/api/v1/media/"+id.String(), nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandlerPresignDownloadSuccess(t *testing.T) {
	r, svc := newMediaRouter(t)
	id := uuid.New()
	svc.On("PresignDownload", mock.Anything, id, time.Hour).
		Return(&domain.PresignDownloadResponse{URL: "https://download.example.test", Key: "key"}, nil)

	w := do(t, r, http.MethodGet, "/api/v1/media/"+id.String()+"/download-url", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://download.example.test")
}

func TestHandlerDeleteForbiddenMaps403(t *testing.T) {
	r, svc := newMediaRouter(t)
	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(domain.ErrForbidden)

	w := do(t, r, http.MethodDelete, "/api/v1/media/"+id.String(), nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandlerListByModelValidatesRequiredQuery(t *testing.T) {
	r, svc := newMediaRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/media?model_type=practice", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "ListByModel")
}

func TestHandlerListByModelValidatesModelID(t *testing.T) {
	r, svc := newMediaRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/media?model_type=practice&model_id=bad", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "ListByModel")
}

func TestHandlerPresignDownloadRejectsUnknownExpiry(t *testing.T) {
	r, svc := newMediaRouter(t)
	w := do(t, r, http.MethodGet, "/api/v1/media/"+uuid.NewString()+"/download-url?expiry=2d", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "PresignDownload")
}

func TestHandlerPresignDownloadPassesExpiry(t *testing.T) {
	r, svc := newMediaRouter(t)
	id := uuid.New()
	svc.On("PresignDownload", mock.Anything, id, 7*24*time.Hour).
		Return(&domain.PresignDownloadResponse{URL: "https://signed", Key: "k"}, nil)
	w := do(t, r, http.MethodGet, "/api/v1/media/"+id.String()+"/download-url?expiry=7d", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListByModelSuccess(t *testing.T) {
	r, svc := newMediaRouter(t)
	modelID := uuid.New()
	svc.On("ListByModel", mock.Anything, "practice", modelID, "attachments").
		Return([]domain.Media{{ID: uuid.New(), ModelType: "practice", ModelID: modelID}}, nil)

	w := do(t, r, http.MethodGet, "/api/v1/media?model_type=practice&model_id="+modelID.String()+"&collection=attachments", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "practice")
}
