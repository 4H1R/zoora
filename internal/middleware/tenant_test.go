package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func TestParseHostLabel(t *testing.T) {
	cases := []struct {
		host, base, want string
	}{
		{"acme.localhost:3000", "localhost", "acme"},
		{"acme.zoora.ir", "zoora.ir", "acme"},
		{"admin.zoora.ir", "zoora.ir", "admin"},
		{"zoora.ir", "zoora.ir", ""},        // apex
		{"localhost:8080", "localhost", ""}, // apex dev
		{"acme.zoora.ir:443", "zoora.ir", "acme"},
	}
	for _, c := range cases {
		if got := parseHostLabel(c.host, c.base); got != c.want {
			t.Fatalf("parseHostLabel(%q,%q)=%q want %q", c.host, c.base, got, c.want)
		}
	}
}

// fakeOrgRepo implements domain.OrganizationRepository but only answers
// FindBySlug; the embedded nil interface panics if any other method is hit,
// which flags accidental coupling.
type fakeOrgRepo struct {
	domain.OrganizationRepository
	slugs map[string]bool
}

func (f fakeOrgRepo) FindBySlug(_ context.Context, slug string) (*domain.Organization, error) {
	if f.slugs[slug] {
		return &domain.Organization{ID: uuid.New(), Slug: slug, Status: domain.OrganizationStatusActive}, nil
	}
	return nil, domain.ErrNotFound
}

func TestOnDemandTLSCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := fakeOrgRepo{slugs: map[string]bool{"acme": true}}
	handler := OnDemandTLSCheck(nil, repo, "zoora.ir", "admin")

	r := gin.New()
	r.GET("/internal/tls-check", handler)

	cases := []struct {
		name, domain string
		want         int
	}{
		{"known tenant", "acme.zoora.ir", http.StatusOK},
		{"admin label", "admin.zoora.ir", http.StatusOK},
		{"www canonical", "www.zoora.ir", http.StatusOK},
		{"unknown tenant", "random.zoora.ir", http.StatusForbidden},
		{"apex", "zoora.ir", http.StatusForbidden},
		{"foreign base", "acme.evil.com", http.StatusForbidden},
		{"missing domain", "", http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/internal/tls-check?domain="+c.domain, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != c.want {
				t.Fatalf("domain=%q got %d want %d", c.domain, w.Code, c.want)
			}
		})
	}
}
