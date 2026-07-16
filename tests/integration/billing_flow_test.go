//go:build integration

package integration

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/billing"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/platform/payment"
	"github.com/4H1R/zoora/tests/testutil"
)

// ---- fakes / stubs -------------------------------------------------------
//
// The billing service constructor takes UNEXPORTED local interfaces
// (planActivator, entitlementsCacheBuster, systemNotifier, objectStorage,
// enqueuer, pdfRenderer). Go interface satisfaction is structural, so these
// exported fakes with matching method sets are assignable as arguments even
// though we cannot name the interface types from this external test package.

// stubGateway satisfies payment.Gateway with canned success responses.
type stubGateway struct{}

func (stubGateway) Name() string { return string(domain.GatewayZarinpal) }

func (stubGateway) Request(_ context.Context, _ payment.RequestInput) (payment.RequestOutput, error) {
	return payment.RequestOutput{Authority: "AUTH-TEST-1", RedirectURL: "https://sandbox/redir"}, nil
}

func (stubGateway) Verify(_ context.Context, _ payment.VerifyInput) (payment.VerifyOutput, error) {
	return payment.VerifyOutput{Status: payment.VerifyStatusSucceeded, RefID: "REF123"}, nil
}

// fakePDFRenderer records invocations and returns a canned object key.
type fakePDFRenderer struct {
	key   string
	calls int
}

func (f *fakePDFRenderer) RenderAndStore(_ context.Context, _ *domain.Invoice) (string, error) {
	f.calls++
	return f.key, nil
}

// fakeEnqueuer records enqueue-PDF requests.
type fakeEnqueuer struct {
	invoiceIDs []uuid.UUID
}

func (f *fakeEnqueuer) EnqueuePDF(_ context.Context, invoiceID uuid.UUID) error {
	f.invoiceIDs = append(f.invoiceIDs, invoiceID)
	return nil
}

// fakeCacheBuster records entitlement-cache invalidations.
type fakeCacheBuster struct {
	orgIDs []uuid.UUID
}

func (f *fakeCacheBuster) Invalidate(_ context.Context, orgID uuid.UUID) error {
	f.orgIDs = append(f.orgIDs, orgID)
	return nil
}

// fakeSystemNotifier captures every SendSystem call.
type fakeSystemNotifier struct {
	sent []domain.SystemNotificationInput
}

func (f *fakeSystemNotifier) SendSystem(_ context.Context, in domain.SystemNotificationInput) error {
	f.sent = append(f.sent, in)
	return nil
}

// fakeStorage satisfies objectStorage (presign).
type fakeStorage struct{}

func (fakeStorage) GeneratePresignedDownloadURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://sandbox/download/" + key, nil
}

// billingHarness bundles the wired service and its collaborators for a test.
type billingHarness struct {
	svc      domain.BillingService
	repo     domain.BillingRepository
	orgs     domain.OrganizationRepository
	pdf      *fakePDFRenderer
	queue    *fakeEnqueuer
	cache    *fakeCacheBuster
	notifier *fakeSystemNotifier
}

func setupBillingFlow(t *testing.T) billingHarness {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.PlanPrice{},
		&domain.Invoice{},
		&domain.InvoiceItem{},
		&domain.Payment{},
		&domain.BillingReminderSent{},
	))

	repo := billing.NewRepository(db)
	orgRepo := organizations.NewRepository(db)
	pdf := &fakePDFRenderer{key: "orgs/x/invoices/y.pdf"}
	queue := &fakeEnqueuer{}
	cache := &fakeCacheBuster{}
	notifier := &fakeSystemNotifier{}

	svc := billing.NewService(
		repo,
		orgRepo,
		orgRepo, // planActivator == real org repo (UpdatePlan)
		cache,
		payment.NewRegistry(stubGateway{}),
		fakeStorage{},
		queue,
		notifier,
		pdf,
		billing.BillingConfig{
			CallbackBaseURL: "https://api.test",
			AppURLTemplate:  "https://{slug}.app.test",
		},
		slog.Default(),
	)

	return billingHarness{
		svc:      svc,
		repo:     repo,
		orgs:     orgRepo,
		pdf:      pdf,
		queue:    queue,
		cache:    cache,
		notifier: notifier,
	}
}

func TestBilling_Checkout_Activate_Idempotent_PDF(t *testing.T) {
	h := setupBillingFlow(t)
	base := context.Background()

	// Seed a free-plan org and a Pro monthly price.
	org := factory.NewOrganization(func(o *domain.Organization) { o.Plan = domain.PlanFree })
	require.NoError(t, h.orgs.Create(base, org))
	require.NoError(t, h.repo.UpsertPrice(base, factory.NewPlanPrice(domain.PlanKey(domain.TierPro, 50), domain.BillingIntervalMonthly)))

	// Caller: a manager of the org (Checkout only requires caller.OrgID != nil).
	ctx := domain.WithCaller(base, domain.Caller{UserID: uuid.New(), OrgID: &org.ID})

	// 1. Checkout -> pending invoice + gateway attempt.
	res, err := h.svc.Checkout(ctx, domain.CheckoutDTO{
		Plan:     domain.PlanKey(domain.TierPro, 50),
		Interval: domain.BillingIntervalMonthly,
		Gateway:  domain.GatewayZarinpal,
	})
	require.NoError(t, err)
	require.NotNil(t, res.Invoice)
	assert.Equal(t, domain.InvoiceStatusPending, res.Invoice.Status)
	assert.Equal(t, "https://sandbox/redir", res.RedirectURL)

	pay, err := h.repo.FindPaymentByAuthority(base, domain.GatewayZarinpal, "AUTH-TEST-1")
	require.NoError(t, err)
	require.NotNil(t, pay.Authority)
	assert.Equal(t, "AUTH-TEST-1", *pay.Authority)
	invoiceID := res.Invoice.ID

	// 2. Gateway callback (verified server-side) -> invoice paid.
	inv, err := h.svc.HandleCallback(base, domain.GatewayZarinpal, "AUTH-TEST-1", true)
	require.NoError(t, err)
	assert.Equal(t, domain.InvoiceStatusPaid, inv.Status)

	// 3. Org activated to Pro, expiry ~ now + 1 month.
	gotOrg, err := h.orgs.FindByID(base, org.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PlanKey(domain.TierPro, 50), gotOrg.Plan)
	require.NotNil(t, gotOrg.PlanExpiresAt)
	wantExpiry := time.Now().AddDate(0, 1, 0)
	assert.WithinDuration(t, wantExpiry, *gotOrg.PlanExpiresAt, 24*time.Hour)

	// 4. CRITICAL: a duplicate callback must NOT double-extend the plan.
	firstExpiry := *gotOrg.PlanExpiresAt
	dup, err := h.svc.HandleCallback(base, domain.GatewayZarinpal, "AUTH-TEST-1", true)
	require.NoError(t, err)
	assert.Equal(t, domain.InvoiceStatusPaid, dup.Status)

	gotOrg, err = h.orgs.FindByID(base, org.ID)
	require.NoError(t, err)
	require.NotNil(t, gotOrg.PlanExpiresAt)
	assert.True(t, gotOrg.PlanExpiresAt.Equal(firstExpiry),
		"idempotent: duplicate callback must not extend expiry (got %s, want %s)",
		gotOrg.PlanExpiresAt, firstExpiry)

	// The paid transition ran exactly once: PDF enqueued once, cache busted once.
	assert.Equal(t, []uuid.UUID{invoiceID}, h.queue.invoiceIDs)
	assert.Equal(t, []uuid.UUID{org.ID}, h.cache.orgIDs)

	// 5. GeneratePDF records the object key via the fake renderer.
	require.NoError(t, h.svc.GeneratePDF(base, invoiceID))
	assert.Equal(t, 1, h.pdf.calls)

	gotInv, err := h.repo.FindInvoiceByID(base, invoiceID)
	require.NoError(t, err)
	require.NotNil(t, gotInv.PDFObjectKey)
	assert.Equal(t, "orgs/x/invoices/y.pdf", *gotInv.PDFObjectKey)
}

func TestBilling_ReminderDedup(t *testing.T) {
	h := setupBillingFlow(t)
	base := context.Background()

	// Fixed UTC clock so the whole-day date math is deterministic.
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	expiry := now.AddDate(0, 0, 7) // exactly 7 days ahead -> renewal_7d stage

	org := factory.NewOrganization(func(o *domain.Organization) {
		o.Plan = domain.PlanKey(domain.TierPro, 50)
		o.PlanExpiresAt = &expiry
	})
	require.NoError(t, h.orgs.Create(base, org))

	// Two sweeps with the same `now`: the dedup guard must suppress the second.
	require.NoError(t, h.svc.RunReminderSweep(base, now))
	require.NoError(t, h.svc.RunReminderSweep(base, now))

	var forOrg int
	for _, n := range h.notifier.sent {
		if n.OrganizationID != nil && *n.OrganizationID == org.ID {
			forOrg++
		}
	}
	assert.Equal(t, 1, forOrg, "renewal_7d reminder must fire exactly once across sweeps (dedup)")
}
