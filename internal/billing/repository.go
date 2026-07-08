package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.BillingRepository {
	return &repository{db: db}
}

func (r *repository) ListActivePrices(ctx context.Context) ([]domain.PlanPrice, error) {
	var prices []domain.PlanPrice
	if err := database.DB(ctx, r.db).
		Where("active = ?", true).
		Order("plan asc, interval asc").
		Find(&prices).Error; err != nil {
		return nil, fmt.Errorf("billing.repository.ListActivePrices: %w", err)
	}
	return prices, nil
}

func (r *repository) FindActivePrice(ctx context.Context, plan domain.Plan, interval domain.BillingInterval, currency string) (*domain.PlanPrice, error) {
	var p domain.PlanPrice
	err := database.DB(ctx, r.db).
		Where("plan = ? AND interval = ? AND currency = ? AND active = ?", plan, interval, currency, true).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("billing.repository.FindActivePrice: %w", err)
	}
	return &p, nil
}

// UpsertPrice deactivates any current active row for (plan,interval,currency)
// then inserts the new active one — preserving history and honoring the partial
// unique index.
func (r *repository) UpsertPrice(ctx context.Context, p *domain.PlanPrice) error {
	if p.Currency == "" {
		p.Currency = domain.CurrencyIRR
	}
	return r.WithTx(ctx, func(ctx context.Context) error {
		if err := database.DB(ctx, r.db).
			Model(&domain.PlanPrice{}).
			Where("plan = ? AND interval = ? AND currency = ? AND active = ?", p.Plan, p.Interval, p.Currency, true).
			Update("active", false).Error; err != nil {
			return fmt.Errorf("billing.repository.UpsertPrice.deactivate: %w", err)
		}
		p.Active = true
		if err := database.DB(ctx, r.db).Create(p).Error; err != nil {
			return fmt.Errorf("billing.repository.UpsertPrice.create: %w", err)
		}
		return nil
	})
}

func (r *repository) DeactivatePrice(ctx context.Context, id uuid.UUID) error {
	res := database.DB(ctx, r.db).Model(&domain.PlanPrice{}).Where("id = ?", id).Update("active", false)
	if res.Error != nil {
		return fmt.Errorf("billing.repository.DeactivatePrice: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) CreateInvoice(ctx context.Context, inv *domain.Invoice) error {
	if err := database.DB(ctx, r.db).Create(inv).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("billing.repository.CreateInvoice: %w", err)
	}
	return nil
}

func (r *repository) FindInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := database.DB(ctx, r.db).
		Preload("Items").
		Preload("Payments").
		First(&inv, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("billing.repository.FindInvoiceByID: %w", err)
	}
	return &inv, nil
}

func (r *repository) FindInvoiceByIDForUpdate(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := database.DB(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Items").
		Preload("Payments").
		First(&inv, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("billing.repository.FindInvoiceByIDForUpdate: %w", err)
	}
	return &inv, nil
}

func (r *repository) UpdateInvoice(ctx context.Context, inv *domain.Invoice) error {
	// Save the header only; items/payments are managed explicitly elsewhere.
	res := database.DB(ctx, r.db).Model(&domain.Invoice{}).
		Where("id = ?", inv.ID).
		Select("Number", "Status", "Currency", "Subtotal", "TaxPercent", "TaxAmount",
			"Total", "Description", "ExpiresAt", "IssuedAt", "PaidAt", "PDFObjectKey", "UpdatedAt").
		Updates(inv)
	if res.Error != nil {
		if database.IsUniqueViolation(res.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("billing.repository.UpdateInvoice: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) ListInvoices(ctx context.Context, orgID uuid.UUID, q domain.ListInvoicesQuery) ([]domain.Invoice, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Invoice{}).
		Preload("Items").
		Where("organization_id = ?", orgID)
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	var items []domain.Invoice
	total, err := listparams.Paginate(base, q.ListParams, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("billing.repository.ListInvoices: %w", err)
	}
	return items, total, nil
}

func (r *repository) AdminListInvoices(ctx context.Context, q domain.AdminListInvoicesQuery) ([]domain.Invoice, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Invoice{}).Preload("Items")
	if q.OrganizationID != nil {
		base = base.Where("organization_id = ?", *q.OrganizationID)
	}
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	var items []domain.Invoice
	total, err := listparams.Paginate(base, q.ListParams, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("billing.repository.AdminListInvoices: %w", err)
	}
	return items, total, nil
}

// NextInvoiceSequence returns a gapless per-year counter derived from the
// highest existing number for the year prefix, +1. Runs inside the issue
// transaction so concurrent issues serialize on the invoices row lock.
func (r *repository) NextInvoiceSequence(ctx context.Context, yearPrefix string) (int64, error) {
	var maxSeq int64
	err := database.DB(ctx, r.db).
		Model(&domain.Invoice{}).
		Where("number LIKE ?", yearPrefix+"-%").
		Select("COALESCE(MAX(CAST(split_part(number, '-', 2) AS BIGINT)), 0)").
		Scan(&maxSeq).Error
	if err != nil {
		return 0, fmt.Errorf("billing.repository.NextInvoiceSequence: %w", err)
	}
	return maxSeq + 1, nil
}

func (r *repository) CreatePayment(ctx context.Context, p *domain.Payment) error {
	if err := database.DB(ctx, r.db).Create(p).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("billing.repository.CreatePayment: %w", err)
	}
	return nil
}

func (r *repository) FindPaymentByAuthority(ctx context.Context, gateway domain.GatewayName, authority string) (*domain.Payment, error) {
	var p domain.Payment
	err := database.DB(ctx, r.db).
		Where("gateway = ? AND authority = ?", gateway, authority).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("billing.repository.FindPaymentByAuthority: %w", err)
	}
	return &p, nil
}

func (r *repository) UpdatePayment(ctx context.Context, p *domain.Payment) error {
	res := database.DB(ctx, r.db).Model(&domain.Payment{}).
		Where("id = ?", p.ID).
		Select("Status", "RefID", "RawResponse", "Note", "VerifiedAt").
		Updates(p)
	if res.Error != nil {
		return fmt.Errorf("billing.repository.UpdatePayment: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) ReminderAlreadySent(ctx context.Context, kind domain.BillingReminderKind, subjectID uuid.UUID, periodKey string) (bool, error) {
	var count int64
	if err := database.DB(ctx, r.db).Model(&domain.BillingReminderSent{}).
		Where("kind = ? AND subject_id = ? AND period_key = ?", kind, subjectID, periodKey).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("billing.repository.ReminderAlreadySent: %w", err)
	}
	return count > 0, nil
}

// MarkReminderSent inserts the dedup row; a unique-violation means a concurrent
// sweep already recorded it — treat as success (idempotent).
func (r *repository) MarkReminderSent(ctx context.Context, rec *domain.BillingReminderSent) error {
	if err := database.DB(ctx, r.db).Create(rec).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return nil
		}
		return fmt.Errorf("billing.repository.MarkReminderSent: %w", err)
	}
	return nil
}

// OrgsWithExpiryBetween returns non-free orgs whose plan_expires_at falls in
// [from,to). Used by the renewal sweep to find reminder candidates.
func (r *repository) OrgsWithExpiryBetween(ctx context.Context, from, to time.Time) ([]domain.Organization, error) {
	var orgs []domain.Organization
	if err := database.DB(ctx, r.db).
		Where("plan <> ? AND plan_expires_at >= ? AND plan_expires_at < ?", domain.PlanFree, from, to).
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("billing.repository.OrgsWithExpiryBetween: %w", err)
	}
	return orgs, nil
}

func (r *repository) PendingInvoicesIssuedBetween(ctx context.Context, from, to time.Time) ([]domain.Invoice, error) {
	var invs []domain.Invoice
	if err := database.DB(ctx, r.db).
		Where("status = ? AND issued_at >= ? AND issued_at < ?", domain.InvoiceStatusPending, from, to).
		Find(&invs).Error; err != nil {
		return nil, fmt.Errorf("billing.repository.PendingInvoicesIssuedBetween: %w", err)
	}
	return invs, nil
}

// ExpirePendingInvoices flips pending invoices past their deadline to expired,
// returning the affected rows.
func (r *repository) ExpirePendingInvoices(ctx context.Context, before time.Time) ([]domain.Invoice, error) {
	var invs []domain.Invoice
	if err := database.DB(ctx, r.db).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", domain.InvoiceStatusPending, before).
		Find(&invs).Error; err != nil {
		return nil, fmt.Errorf("billing.repository.ExpirePendingInvoices.find: %w", err)
	}
	if len(invs) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, len(invs))
	for i := range invs {
		ids[i] = invs[i].ID
	}
	if err := database.DB(ctx, r.db).Model(&domain.Invoice{}).
		Where("id IN ?", ids).
		Update("status", domain.InvoiceStatusExpired).Error; err != nil {
		return nil, fmt.Errorf("billing.repository.ExpirePendingInvoices.update: %w", err)
	}
	return invs, nil
}

// WithTx runs fn inside a DB transaction. It delegates to the shared
// database.Transactor, which stashes the tx in ctx so that database.DB(ctx, r.db)
// inside fn picks it up. Nested calls reuse the existing tx.
func (r *repository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return database.NewTransactor(r.db).RunInTx(ctx, fn)
}
