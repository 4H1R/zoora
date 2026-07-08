package billing

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type adminService struct {
	repo domain.BillingRepository
	svc  *service // reuse markPaidAndActivate + assignInvoiceNumber
}

// NewAdminService shares the concrete *service so admin mark-paid reuses the
// exact same activation path as the gateway callback.
func NewAdminService(core domain.BillingService) domain.BillingAdminService {
	s := core.(*service)
	return &adminService{repo: s.repo, svc: s}
}

func (a *adminService) ListPrices(ctx context.Context) ([]domain.PlanPrice, error) {
	return a.repo.ListActivePrices(ctx)
}

func (a *adminService) UpsertPrice(ctx context.Context, dto domain.UpsertPlanPriceDTO) (*domain.PlanPrice, error) {
	if !dto.Plan.Valid() || !dto.Interval.Valid() {
		return nil, domain.ErrInvalidPlan
	}
	currency := dto.Currency
	if currency == "" {
		currency = domain.CurrencyIRR
	}
	p := &domain.PlanPrice{Plan: dto.Plan, Interval: dto.Interval, Currency: currency, Amount: dto.Amount, Active: true}
	if err := a.repo.UpsertPrice(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (a *adminService) DeactivatePrice(ctx context.Context, id uuid.UUID) error {
	return a.repo.DeactivatePrice(ctx, id)
}

// CreateInvoice builds a DRAFT custom invoice with arbitrary line items.
func (a *adminService) CreateInvoice(ctx context.Context, dto domain.AdminCreateInvoiceDTO) (*domain.Invoice, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return nil, domain.ErrForbidden
	}
	items := make([]domain.InvoiceItem, 0, len(dto.Items))
	for _, it := range dto.Items {
		items = append(items, domain.InvoiceItem{
			Kind:        it.Kind,
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitAmount:  it.UnitAmount,
			Amount:      lineAmount(it.Quantity, it.UnitAmount),
			Currency:    domain.CurrencyIRR,
		})
	}
	sub, tax, total := computeTotals(items, dto.TaxPercent)
	inv := &domain.Invoice{
		OrganizationID: dto.OrganizationID,
		Status:         domain.InvoiceStatusDraft,
		Currency:       domain.CurrencyIRR,
		Subtotal:       sub,
		TaxPercent:     dto.TaxPercent,
		TaxAmount:      tax,
		Total:          total,
		Description:    dto.Description,
		CreatedBy:      &caller.UserID,
		Items:          items,
	}
	if err := a.repo.CreateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// IssueInvoice moves draft -> pending and assigns the sequential number.
func (a *adminService) IssueInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	inv, err := a.repo.FindInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.Status != domain.InvoiceStatusDraft {
		return nil, domain.ErrInvoiceNotDraft
	}
	now := time.Now()
	number, err := a.svc.assignInvoiceNumber(ctx, now)
	if err != nil {
		return nil, err
	}
	expiresAt := now.Add(a.svc.cfg.PendingTTL)
	inv.Number = &number
	inv.Status = domain.InvoiceStatusPending
	inv.IssuedAt = &now
	inv.ExpiresAt = &expiresAt
	if err := a.repo.UpdateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// MarkPaid records a manual (offline) payment and runs the SAME activation path.
func (a *adminService) MarkPaid(ctx context.Context, id uuid.UUID, dto domain.AdminMarkPaidDTO) (*domain.Invoice, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return nil, domain.ErrForbidden
	}
	inv, err := a.repo.FindInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.Status != domain.InvoiceStatusPending && inv.Status != domain.InvoiceStatusDraft {
		return nil, domain.ErrInvoiceNotPayable
	}
	now := time.Now()
	refID := dto.RefID
	pay := &domain.Payment{
		InvoiceID:      inv.ID,
		OrganizationID: inv.OrganizationID,
		Gateway:        domain.GatewayManual,
		Status:         domain.PaymentStatusSucceeded,
		Amount:         inv.Total,
		Currency:       inv.Currency,
		Note:           dto.Note,
		CreatedBy:      &caller.UserID,
		VerifiedAt:     &now,
	}
	if refID != "" {
		pay.RefID = &refID
	}
	if err := a.repo.CreatePayment(ctx, pay); err != nil {
		return nil, err
	}
	return a.svc.markPaidAndActivate(ctx, inv.ID, now)
}

func (a *adminService) CancelInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	inv, err := a.repo.FindInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.Status == domain.InvoiceStatusPaid || inv.Status == domain.InvoiceStatusRefunded {
		return nil, domain.ErrInvoiceNotPayable
	}
	inv.Status = domain.InvoiceStatusCanceled
	if err := a.repo.UpdateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// RefundInvoice is record-only: paid -> refunded, no gateway call, no plan revoke.
func (a *adminService) RefundInvoice(ctx context.Context, id uuid.UUID, dto domain.AdminRefundDTO) (*domain.Invoice, error) {
	inv, err := a.repo.FindInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.Status != domain.InvoiceStatusPaid {
		return nil, domain.ErrInvoiceNotPayable
	}
	inv.Status = domain.InvoiceStatusRefunded
	inv.Description = inv.Description + " | refund: " + dto.Reason
	if err := a.repo.UpdateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (a *adminService) ListInvoices(ctx context.Context, q domain.AdminListInvoicesQuery) ([]domain.Invoice, int64, error) {
	return a.repo.AdminListInvoices(ctx, q)
}

func (a *adminService) GetInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	return a.repo.FindInvoiceByID(ctx, id)
}

var _ domain.BillingAdminService = (*adminService)(nil)
