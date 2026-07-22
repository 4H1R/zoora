package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// mount builds a gin engine with CORS(origins) in front of a trivial handler.
func mountCORS(origins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS(origins))
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

// TestCORSBareWildcardDropsCredentials verifies that a bare "*" allow-list
// never advertises Access-Control-Allow-Credentials (a credentialed wildcard
// is insecure and browsers reject it).
func TestCORSBareWildcardDropsCredentials(t *testing.T) {
	r := mountCORS([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://anything.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want empty for bare wildcard", got)
	}
}

// TestCORSSubdomainWildcardKeepsCredentials verifies that an explicit
// subdomain wildcard allow-list still reflects the matched origin and keeps
// credentials enabled — multi-tenant cookies/sessions depend on this.
func TestCORSSubdomainWildcardKeepsCredentials(t *testing.T) {
	r := mountCORS([]string{"https://*.zoora.ir"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://acme.zoora.ir")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://acme.zoora.ir" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want the reflected tenant origin", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
	}
}
