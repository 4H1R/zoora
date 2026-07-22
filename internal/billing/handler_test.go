package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// callbackStubService implements domain.BillingService; only HandleCallback is
// exercised by the callback tests, the rest panic if unexpectedly called.
type callbackStubService struct {
	called bool
	inv    *domain.Invoice
	err    error
}

func (s *callbackStubService) HandleCallback(ctx context.Context, gateway domain.GatewayName, authority string, gatewayOK bool) (*domain.Invoice, error) {
	s.called = true
	return s.inv, s.err
}

func (s *callbackStubService) ListPlanPrices(ctx context.Context) ([]domain.PlanPrice, error) {
	panic("unexpected call")
}

func (s *callbackStubService) Checkout(ctx context.Context, dto domain.CheckoutDTO) (*domain.CheckoutResult, error) {
	panic("unexpected call")
}

func (s *callbackStubService) ListInvoices(ctx context.Context, q domain.ListInvoicesQuery) ([]domain.Invoice, int64, error) {
	panic("unexpected call")
}

func (s *callbackStubService) GetInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	panic("unexpected call")
}

func (s *callbackStubService) InvoicePDFURL(ctx context.Context, id uuid.UUID) (string, error) {
	panic("unexpected call")
}

func (s *callbackStubService) GeneratePDF(ctx context.Context, invoiceID uuid.UUID) error {
	panic("unexpected call")
}

func (s *callbackStubService) RunReminderSweep(ctx context.Context, now time.Time) error {
	panic("unexpected call")
}

func (s *callbackStubService) ExpireStaleInvoices(ctx context.Context, now time.Time) error {
	panic("unexpected call")
}

const testAppURLTemplate = "http://{slug}.localhost:5173"

func doCallback(t *testing.T, svc domain.BillingService, org string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc, testAppURLTemplate)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "gateway", Value: "zarinpal"}}
	req := httptest.NewRequest(http.MethodGet, "/billing/callback/zarinpal?Authority=A123&Status=OK&org="+org, nil)
	c.Request = req
	h.Callback(c)
	return w
}

func TestCallbackRejectsMaliciousSlugs(t *testing.T) {
	cases := map[string]string{
		"slash path":     "evil.com%2F", // decodes to evil.com/
		"full host":      "attacker.example",
		"at sign":        "user%40evil.com", // decodes to user@evil.com
		"colon port":     "host%3A8080",     // decodes to host:8080
		"uppercase":      "Acme",
		"empty":          "",
		"reserved admin": "admin",
	}
	for name, org := range cases {
		t.Run(name, func(t *testing.T) {
			svc := &callbackStubService{}
			w := doCallback(t, svc, org)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", w.Code)
			}
			if loc := w.Header().Get("Location"); loc != "" {
				t.Fatalf("unexpected redirect Location = %q, want none", loc)
			}
			if svc.called {
				t.Fatalf("HandleCallback should not run for an invalid slug")
			}
		})
	}
}

func TestCallbackAllowsValidSlug(t *testing.T) {
	inv := &domain.Invoice{ID: uuid.New(), Status: domain.InvoiceStatusPaid}
	svc := &callbackStubService{inv: inv}
	w := doCallback(t, svc, "acme")

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, "http://acme.localhost:5173/org/billing/result?status=success") {
		t.Fatalf("Location = %q, want templated acme host with success", loc)
	}
	if !strings.Contains(loc, "invoice="+inv.ID.String()) {
		t.Fatalf("Location = %q, want invoice id", loc)
	}
	if !svc.called {
		t.Fatalf("HandleCallback should run for a valid slug")
	}
}
