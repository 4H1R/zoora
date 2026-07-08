package billing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/payment"
)

// planActivator applies the resolved plan/expiry to the org. Satisfied by the
// organizations repository (domain.OrganizationRepository.UpdatePlan); injected
// so billing never imports another feature package.
type planActivator interface {
	UpdatePlan(ctx context.Context, id uuid.UUID, plan domain.Plan, expiresAt *time.Time) error
}

// entitlementsCacheBuster drops the cached (plan, expiry) snapshot for an org
// after a plan change. The real invalidation is the package-level
// cache.InvalidateOrgPlan(ctx, rdb, orgID); a thin adapter injected at wiring
// time satisfies this interface (kept as a method so billing stays free of the
// redis client type).
type entitlementsCacheBuster interface {
	Invalidate(ctx context.Context, orgID uuid.UUID) error
}

// systemNotifier sends a caller-less reminder (notifications SendSystem).
type systemNotifier interface {
	SendSystem(ctx context.Context, in domain.SystemNotificationInput) error
}

// objectStorage is the S3 subset billing needs (presign for PDF download).
type objectStorage interface {
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// enqueuer abstracts the Asynq client for enqueuing the PDF task.
type enqueuer interface {
	EnqueuePDF(ctx context.Context, invoiceID uuid.UUID) error
}

// pdfRenderer renders an invoice to PDF bytes + uploads to S3, returning the key.
type pdfRenderer interface {
	RenderAndStore(ctx context.Context, inv *domain.Invoice) (objectKey string, err error)
}

type service struct {
	repo      domain.BillingRepository
	orgRepo   domain.OrganizationRepository
	activator planActivator
	cache     entitlementsCacheBuster
	gateways  *payment.Registry
	storage   objectStorage
	queue     enqueuer
	notifier  systemNotifier
	pdf       pdfRenderer
	cfg       BillingConfig
	logger    *slog.Logger
}

// BillingConfig carries the runtime knobs the service needs (from config.Config).
type BillingConfig struct {
	CallbackBaseURL string
	AppBaseURL      string
	PendingTTL      time.Duration // how long a pending invoice stays payable (default 7d)
	Issuer          IssuerConfig
}

// IssuerConfig identifies the merchant on the PDF receipt (added early so the
// Phase 5 renderer wiring is clean).
type IssuerConfig struct {
	Name       string
	EconomicID string
	Address    string
	Phone      string
}

func NewService(
	repo domain.BillingRepository,
	orgRepo domain.OrganizationRepository,
	activator planActivator,
	cache entitlementsCacheBuster,
	gateways *payment.Registry,
	storage objectStorage,
	queue enqueuer,
	notifier systemNotifier,
	pdf pdfRenderer,
	cfg BillingConfig,
	logger *slog.Logger,
) domain.BillingService {
	if cfg.PendingTTL == 0 {
		cfg.PendingTTL = 7 * 24 * time.Hour
	}
	return &service{
		repo: repo, orgRepo: orgRepo, activator: activator, cache: cache,
		gateways: gateways, storage: storage, queue: queue, notifier: notifier,
		pdf: pdf, cfg: cfg, logger: logger,
	}
}

func (s *service) ListPlanPrices(ctx context.Context) ([]domain.PlanPrice, error) {
	return s.repo.ListActivePrices(ctx)
}

// Checkout creates a pending invoice for a single plan+interval line, opens a
// gateway payment attempt, and returns the redirect URL. Downgrade-while-active
// is rejected up-front by NextPlanState so we never create an unpayable invoice.
func (s *service) Checkout(ctx context.Context, dto domain.CheckoutDTO) (*domain.CheckoutResult, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	if !dto.Plan.Valid() || dto.Plan == domain.PlanFree {
		return nil, domain.ErrInvalidPlan
	}
	if !dto.Interval.Valid() {
		return nil, domain.ErrInvalidInterval
	}
	gw, err := s.gateways.Get(string(dto.Gateway))
	if err != nil {
		return nil, domain.ErrGatewayNotFound
	}

	org, err := s.orgRepo.FindByID(ctx, *caller.OrgID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	// Validate the transition is allowed (blocks downgrade) before charging.
	if _, _, err := domain.NextPlanState(org.Plan, org.PlanExpiresAt, dto.Plan, dto.Interval, now); err != nil {
		return nil, err
	}

	price, err := s.repo.FindActivePrice(ctx, dto.Plan, dto.Interval, domain.CurrencyIRR)
	if err != nil {
		return nil, err
	}

	planCopy, intervalCopy := dto.Plan, dto.Interval
	item := domain.InvoiceItem{
		Kind:        domain.InvoiceItemPlanSubscription,
		Description: planItemDescription(dto.Plan, dto.Interval),
		Plan:        &planCopy,
		Interval:    &intervalCopy,
		Quantity:    1,
		UnitAmount:  price.Amount,
		Amount:      lineAmount(1, price.Amount),
		Currency:    price.Currency,
	}
	sub, tax, total := computeTotals([]domain.InvoiceItem{item}, 0)
	expiresAt := now.Add(s.cfg.PendingTTL)
	inv := &domain.Invoice{
		OrganizationID: org.ID,
		Status:         domain.InvoiceStatusPending,
		Currency:       price.Currency,
		Subtotal:       sub,
		TaxPercent:     0,
		TaxAmount:      tax,
		Total:          total,
		Description:    item.Description,
		IssuedAt:       &now,
		ExpiresAt:      &expiresAt,
		Items:          []domain.InvoiceItem{item},
	}
	// Assign a human number at issue (pending == issued for self-serve).
	number, err := s.assignInvoiceNumber(ctx, now)
	if err != nil {
		return nil, err
	}
	inv.Number = &number
	if err := s.repo.CreateInvoice(ctx, inv); err != nil {
		return nil, err
	}

	// Open the gateway attempt.
	callbackURL := fmt.Sprintf("%s/api/billing/callback/%s", s.cfg.CallbackBaseURL, dto.Gateway)
	out, err := gw.Request(ctx, payment.RequestInput{
		Amount:      inv.Total,
		Currency:    inv.Currency,
		CallbackURL: callbackURL,
		Description: inv.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("billing.Checkout.gatewayRequest: %w", err)
	}
	authority := out.Authority
	pay := &domain.Payment{
		InvoiceID:      inv.ID,
		OrganizationID: org.ID,
		Gateway:        dto.Gateway,
		Status:         domain.PaymentStatusPending,
		Amount:         inv.Total,
		Currency:       inv.Currency,
		Authority:      &authority,
	}
	if err := s.repo.CreatePayment(ctx, pay); err != nil {
		return nil, err
	}
	return &domain.CheckoutResult{Invoice: inv, RedirectURL: out.RedirectURL}, nil
}

func (s *service) ListInvoices(ctx context.Context, q domain.ListInvoicesQuery) ([]domain.Invoice, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || caller.OrgID == nil {
		return nil, 0, domain.ErrForbidden
	}
	return s.repo.ListInvoices(ctx, *caller.OrgID, q)
}

func (s *service) GetInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	inv, err := s.repo.FindInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.OrganizationID != *caller.OrgID {
		return nil, domain.ErrNotFound // don't leak cross-tenant existence
	}
	return inv, nil
}

func (s *service) InvoicePDFURL(ctx context.Context, id uuid.UUID) (string, error) {
	inv, err := s.GetInvoice(ctx, id) // enforces tenant ownership
	if err != nil {
		return "", err
	}
	if inv.PDFObjectKey == nil {
		return "", domain.ErrNotFound
	}
	return s.storage.GeneratePresignedDownloadURL(ctx, *inv.PDFObjectKey, 10*time.Minute)
}

// assignInvoiceNumber builds a gapless "1405-000123" number using the Jalali
// year prefix.
func (s *service) assignInvoiceNumber(ctx context.Context, now time.Time) (string, error) {
	prefix := jalaliYearPrefix(now)
	seq, err := s.repo.NextInvoiceSequence(ctx, prefix)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%06d", prefix, seq), nil
}

func planItemDescription(plan domain.Plan, interval domain.BillingInterval) string {
	return fmt.Sprintf("%s plan — %s", plan, interval)
}

// HandleCallback is invoked from the public gateway-redirect route. It NEVER
// trusts gatewayOK alone — it always re-verifies server-side, checks the
// amount, then settles idempotently. Returns the (possibly already-paid)
// invoice so the handler can redirect the user to a result page.
func (s *service) HandleCallback(ctx context.Context, gateway domain.GatewayName, authority string, gatewayOK bool) (*domain.Invoice, error) {
	gw, err := s.gateways.Get(string(gateway))
	if err != nil {
		return nil, domain.ErrGatewayNotFound
	}
	pay, err := s.repo.FindPaymentByAuthority(ctx, gateway, authority)
	if err != nil {
		return nil, err
	}
	inv, err := s.repo.FindInvoiceByID(ctx, pay.InvoiceID)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	// User bailed / gateway said NOK: mark the attempt canceled, leave invoice pending.
	if !gatewayOK {
		return s.failPayment(ctx, pay, domain.PaymentStatusCanceled, now, nil)
	}

	// SERVER-SIDE VERIFY (mandatory) with the INVOICE amount, not the callback's.
	// The gateway only returns success when the merchant+amount+authority triple
	// matches its record, so a tampered amount naturally yields a non-success
	// code -> VerifyStatusFailed. Never pass a client-supplied amount here.
	v, err := gw.Verify(ctx, payment.VerifyInput{Authority: authority, Amount: inv.Total})
	if err != nil {
		return nil, fmt.Errorf("billing.HandleCallback.verify: %w", err)
	}
	if v.Status == payment.VerifyStatusFailed {
		return s.failPayment(ctx, pay, domain.PaymentStatusFailed, now, v.Raw)
	}

	// Success or already-verified (Zarinpal 101): both mean the money is captured.
	// markPaidAndActivate is idempotent, so a duplicate callback cannot double-extend.
	pay.Status = domain.PaymentStatusSucceeded
	pay.VerifiedAt = &now
	if v.RefID != "" {
		ref := v.RefID
		pay.RefID = &ref
	}
	pay.RawResponse = v.Raw
	if err := s.repo.UpdatePayment(ctx, pay); err != nil {
		return nil, err
	}
	return s.markPaidAndActivate(ctx, inv.ID, now)
}

func (s *service) failPayment(ctx context.Context, pay *domain.Payment, status domain.PaymentStatus, now time.Time, raw []byte) (*domain.Invoice, error) {
	pay.Status = status
	pay.VerifiedAt = &now
	if raw != nil {
		pay.RawResponse = raw
	}
	if err := s.repo.UpdatePayment(ctx, pay); err != nil {
		return nil, err
	}
	return s.repo.FindInvoiceByID(ctx, pay.InvoiceID)
}

// GeneratePDF renders + stores the receipt PDF for a paid invoice and records
// the object key. Only paid invoices get receipts; a retry after a status
// change is a no-op.
func (s *service) GeneratePDF(ctx context.Context, invoiceID uuid.UUID) error {
	inv, err := s.repo.FindInvoiceByID(ctx, invoiceID)
	if err != nil {
		return err
	}
	if inv.Status != domain.InvoiceStatusPaid {
		return nil
	}
	key, err := s.pdf.RenderAndStore(ctx, inv)
	if err != nil {
		return err
	}
	inv.PDFObjectKey = &key
	return s.repo.UpdateInvoice(ctx, inv)
}

var _ domain.BillingService = (*service)(nil)
