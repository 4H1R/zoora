package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLivenessHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checker := NewChecker(nil, nil, nil)
	router := gin.New()
	router.GET("/live", checker.LivenessHandler)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/live", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status body = %#v, want ok", body)
	}
}

func TestStatusString(t *testing.T) {
	if got := statusString(http.StatusOK); got != "ok" {
		t.Fatalf("statusString(200) = %q, want ok", got)
	}
	if got := statusString(http.StatusServiceUnavailable); got != "unavailable" {
		t.Fatalf("statusString(503) = %q, want unavailable", got)
	}
}
