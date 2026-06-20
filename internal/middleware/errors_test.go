package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
)

func TestErrorHandlerMapsDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"not found", domain.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, "FORBIDDEN"},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED"},
		{"conflict", domain.ErrConflict, http.StatusConflict, "CONFLICT"},
		{"validation", domain.NewValidationError(map[string]string{"name": "required"}), http.StatusBadRequest, "VALIDATION_ERROR"},
		{"internal", errors.New("database exploded"), http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := testRouterWithErrorHandler()
			router.GET("/", func(c *gin.Context) {
				_ = c.Error(tt.err)
			})

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, tt.status, w.Body.String())
			}

			var body domain.Response
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if body.Success {
				t.Fatal("Success = true, want false")
			}
			if body.Error == nil || body.Error.Code != tt.code {
				t.Fatalf("error code = %#v, want %s", body.Error, tt.code)
			}
			if tt.status >= http.StatusInternalServerError && body.Error.Message != "internal server error" {
				t.Fatalf("internal message leaked: %q", body.Error.Message)
			}
			if tt.code == "VALIDATION_ERROR" && body.Error.Fields["name"] != "required" {
				t.Fatalf("validation fields = %#v, want name required", body.Error.Fields)
			}
		})
	}
}

func TestErrorHandlerLeavesWrittenResponsesUntouched(t *testing.T) {
	router := testRouterWithErrorHandler()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusTeapot, gin.H{"already": "written"})
		_ = c.Error(domain.ErrForbidden)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTeapot)
	}
	if got := w.Body.String(); got != `{"already":"written"}` {
		t.Fatalf("body = %s, want original response", got)
	}
}

func testRouterWithErrorHandler() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil))))
	return router
}
