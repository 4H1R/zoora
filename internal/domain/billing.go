package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Money is an integer amount in the MINOR unit of Currency (ISO 4217).
// IRR: Amount is in Rial. The UI speaks Toman (= 10 Rial); convert only at the
// edge. No floats. Every money field carries its own currency — there is no
// global "the currency".
type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

const CurrencyIRR = "IRR"

// BillingInterval is the subscription period a price or purchase covers.
type BillingInterval string

const (
	BillingIntervalMonthly BillingInterval = "monthly"
	BillingIntervalYearly  BillingInterval = "yearly"
)

func (b BillingInterval) Valid() bool {
	return b == BillingIntervalMonthly || b == BillingIntervalYearly
}

// Extend advances from by one interval (calendar-aware).
func (b BillingInterval) Extend(from time.Time) time.Time {
	if b == BillingIntervalYearly {
		return from.AddDate(1, 0, 0)
	}
	return from.AddDate(0, 1, 0)
}

// ---- plan_prices ----

type PlanPrice struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Plan      Plan            `gorm:"type:varchar(20);not null" json:"plan"`
	Interval  BillingInterval `gorm:"type:varchar(20);not null" json:"interval"`
	Currency  string          `gorm:"type:char(3);not null;default:'IRR'" json:"currency"`
	Amount    int64           `gorm:"not null" json:"amount"` // minor units (Rial)
	Active    bool            `gorm:"not null;default:true" json:"active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (PlanPrice) TableName() string { return "plan_prices" }
func (p PlanPrice) Money() Money    { return Money{Amount: p.Amount, Currency: p.Currency} }

// ---- invoices ----

type InvoiceStatus string

const (
	InvoiceStatusDraft    InvoiceStatus = "draft"
	InvoiceStatusPending  InvoiceStatus = "pending"
	InvoiceStatusPaid     InvoiceStatus = "paid"
	InvoiceStatusCanceled InvoiceStatus = "canceled"
	InvoiceStatusExpired  InvoiceStatus = "expired"
	InvoiceStatusRefunded InvoiceStatus = "refunded"
)

type Invoice struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Number         *string       `gorm:"uniqueIndex" json:"number,omitempty"` // assigned at issue
	OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index" json:"organization_id"`
	Status         InvoiceStatus `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	Currency       string        `gorm:"type:char(3);not null;default:'IRR'" json:"currency"`
	Subtotal       int64         `gorm:"not null" json:"subtotal"`
	TaxPercent     int           `gorm:"not null;default:0" json:"tax_percent"`
	TaxAmount      int64         `gorm:"not null;default:0" json:"tax_amount"`
	Total          int64         `gorm:"not null" json:"total"`
	Description    string        `json:"description"`
	ExpiresAt      *time.Time    `json:"expires_at,omitempty"` // pending payment deadline
	IssuedAt       *time.Time    `json:"issued_at,omitempty"`
	PaidAt         *time.Time    `json:"paid_at,omitempty"`
	PDFObjectKey   *string       `json:"-"`
	CreatedBy      *uuid.UUID    `gorm:"type:uuid" json:"created_by,omitempty"` // admin; nil = self-serve

	Items    []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`
	Payments []Payment     `gorm:"foreignKey:InvoiceID" json:"payments,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Invoice) TableName() string { return "invoices" }
func (i Invoice) Money() Money    { return Money{Amount: i.Total, Currency: i.Currency} }

// ---- invoice_items ----

type InvoiceItemKind string

const (
	InvoiceItemPlanSubscription InvoiceItemKind = "plan_subscription"
	InvoiceItemCustom           InvoiceItemKind = "custom"
	InvoiceItemAddon            InvoiceItemKind = "addon"
)

type InvoiceItem struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	InvoiceID   uuid.UUID        `gorm:"type:uuid;not null;index" json:"invoice_id"`
	Kind        InvoiceItemKind  `gorm:"type:varchar(30);not null" json:"kind"`
	Description string           `gorm:"not null" json:"description"`
	Plan        *Plan            `gorm:"type:varchar(20)" json:"plan,omitempty"`
	Interval    *BillingInterval `gorm:"type:varchar(20)" json:"interval,omitempty"`
	PeriodStart *time.Time       `json:"period_start,omitempty"`
	PeriodEnd   *time.Time       `json:"period_end,omitempty"`
	Quantity    int              `gorm:"not null;default:1" json:"quantity"`
	UnitAmount  int64            `gorm:"not null" json:"unit_amount"`
	Amount      int64            `gorm:"not null" json:"amount"`
	Currency    string           `gorm:"type:char(3);not null;default:'IRR'" json:"currency"`
	CreatedAt   time.Time        `json:"created_at"`
}

func (InvoiceItem) TableName() string { return "invoice_items" }

// ---- payments (gateway attempts) ----

type GatewayName string

const (
	GatewayZarinpal GatewayName = "zarinpal"
	GatewayManual   GatewayName = "manual"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCanceled  PaymentStatus = "canceled"
)

type Payment struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	InvoiceID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"invoice_id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	Gateway        GatewayName    `gorm:"type:varchar(20);not null" json:"gateway"`
	Status         PaymentStatus  `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Amount         int64          `gorm:"not null" json:"amount"`
	Currency       string         `gorm:"type:char(3);not null;default:'IRR'" json:"currency"`
	Authority      *string        `gorm:"index" json:"authority,omitempty"` // gateway token (zarinpal)
	RefID          *string        `json:"ref_id,omitempty"`
	RawResponse    datatypes.JSON `gorm:"type:jsonb" json:"-"`
	Note           string         `json:"note,omitempty"`                        // manual admin note
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"` // admin for manual
	CreatedAt      time.Time      `json:"created_at"`
	VerifiedAt     *time.Time     `json:"verified_at,omitempty"`
}

func (Payment) TableName() string { return "payments" }

// ---- billing_reminders_sent (dedup guard) ----

type BillingReminderKind string

const (
	ReminderRenewal7d        BillingReminderKind = "renewal_7d"
	ReminderRenewal3d        BillingReminderKind = "renewal_3d"
	ReminderRenewalDue       BillingReminderKind = "renewal_due"
	ReminderRenewalGrace     BillingReminderKind = "renewal_grace"
	ReminderInvoiceUnpaid24h BillingReminderKind = "invoice_unpaid_24h"
	ReminderInvoiceUnpaid72h BillingReminderKind = "invoice_unpaid_72h"
)

type BillingReminderSent struct {
	ID        uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuidv7()"`
	Kind      BillingReminderKind `gorm:"type:varchar(30);not null"`
	SubjectID uuid.UUID           `gorm:"type:uuid;not null"`
	PeriodKey string              `gorm:"type:varchar(40);not null"`
	CreatedAt time.Time
}

func (BillingReminderSent) TableName() string { return "billing_reminders_sent" }

// ---- plan activation logic ----

// tierRank orders tiers for upgrade/downgrade comparison.
func tierRank(t PlanTier) int {
	switch t {
	case TierMax:
		return 4
	case TierPro:
		return 3
	case TierPlus:
		return 2
	default: // free / unknown
		return 1
	}
}

// sizeRank orders member capacities (0 for unknown sizes).
func sizeRank(size int64) int {
	for i, s := range PlanSizes {
		if s == size {
			return i + 1
		}
	}
	return 0
}

// planRank orders plans tier-first, size-second: any higher tier outranks any
// size of a lower tier; within a tier, bigger capacity ranks higher.
func planRank(p Plan) int {
	return tierRank(p.Tier())*100 + sizeRank(p.Size())
}

// NextPlanState computes the (plan, expiry) an org should have after a
// successful purchase of `buy` for one `interval`, given its current plan and
// expiry. Pure — the caller supplies now.
//
//   - same tier & still active  -> extend from current expiry
//   - higher tier               -> upgrade, expiry from now
//   - free/expired              -> set tier, expiry from now
//   - lower tier while active   -> ErrDowngradeNotAllowed
func NextPlanState(curPlan Plan, curExpiry *time.Time, buy Plan, interval BillingInterval, now time.Time) (Plan, time.Time, error) {
	if !buy.Valid() {
		return "", time.Time{}, ErrInvalidPlan
	}
	if !interval.Valid() {
		return "", time.Time{}, ErrInvalidInterval
	}
	// Effective current plan: expired plans behave as free.
	effective := EffectiveEntitlements(curPlan, curExpiry, now).Plan
	active := effective.Tier() != TierFree

	if planRank(buy) < planRank(effective) && active {
		return "", time.Time{}, ErrDowngradeNotAllowed
	}

	base := now
	if buy == effective && active && curExpiry != nil && curExpiry.After(now) {
		base = *curExpiry // same tier, still active: don't burn remaining days
	}
	return buy, interval.Extend(base), nil
}

// ---- DTOs / queries ----

type CheckoutDTO struct {
	Plan     Plan            `json:"plan" binding:"required"`
	Interval BillingInterval `json:"interval" binding:"required"`
	Gateway  GatewayName     `json:"gateway" binding:"required"`
}

type CheckoutResult struct {
	Invoice     *Invoice `json:"invoice"`
	RedirectURL string   `json:"redirect_url"`
}

type ListInvoicesQuery struct {
	Status     *InvoiceStatus `form:"status"`
	ListParams ListParams     `form:"-"`
}

type AdminListInvoicesQuery struct {
	OrganizationID *uuid.UUID     `form:"-"`
	Status         *InvoiceStatus `form:"status"`
	ListParams     ListParams     `form:"-"`
}

type UpsertPlanPriceDTO struct {
	Plan     Plan            `json:"plan" binding:"required"`
	Interval BillingInterval `json:"interval" binding:"required"`
	Currency string          `json:"currency" binding:"omitempty,len=3"`
	Amount   int64           `json:"amount" binding:"required,gt=0"` // Rial
}

type AdminCreateInvoiceItemDTO struct {
	Kind        InvoiceItemKind `json:"kind" binding:"required,oneof=plan_subscription custom addon"`
	Description string          `json:"description" binding:"required,min=2"`
	Quantity    int             `json:"quantity" binding:"required,gt=0"`
	UnitAmount  int64           `json:"unit_amount" binding:"required,gt=0"` // Rial
}

type AdminCreateInvoiceDTO struct {
	OrganizationID uuid.UUID                   `json:"organization_id" binding:"required"`
	Description    string                      `json:"description"`
	TaxPercent     int                         `json:"tax_percent" binding:"omitempty,gte=0,lte=100"`
	Items          []AdminCreateInvoiceItemDTO `json:"items" binding:"required,min=1,dive"`
}

type AdminMarkPaidDTO struct {
	Note  string `json:"note"`
	RefID string `json:"ref_id"`
}

type AdminRefundDTO struct {
	Reason string `json:"reason" binding:"required,min=2"`
}

// ---- repository ----

type BillingRepository interface {
	// prices
	ListActivePrices(ctx context.Context) ([]PlanPrice, error)
	FindActivePrice(ctx context.Context, plan Plan, interval BillingInterval, currency string) (*PlanPrice, error)
	UpsertPrice(ctx context.Context, p *PlanPrice) error
	DeactivatePrice(ctx context.Context, id uuid.UUID) error

	// invoices
	CreateInvoice(ctx context.Context, inv *Invoice) error
	FindInvoiceByID(ctx context.Context, id uuid.UUID) (*Invoice, error)
	FindInvoiceByIDForUpdate(ctx context.Context, id uuid.UUID) (*Invoice, error)
	UpdateInvoice(ctx context.Context, inv *Invoice) error
	ListInvoices(ctx context.Context, orgID uuid.UUID, q ListInvoicesQuery) ([]Invoice, int64, error)
	AdminListInvoices(ctx context.Context, q AdminListInvoicesQuery) ([]Invoice, int64, error)
	NextInvoiceSequence(ctx context.Context, yearPrefix string) (int64, error)

	// payments
	CreatePayment(ctx context.Context, p *Payment) error
	FindPaymentByAuthority(ctx context.Context, gateway GatewayName, authority string) (*Payment, error)
	UpdatePayment(ctx context.Context, p *Payment) error

	// reminders (dedup + sweep source data)
	ReminderAlreadySent(ctx context.Context, kind BillingReminderKind, subjectID uuid.UUID, periodKey string) (bool, error)
	MarkReminderSent(ctx context.Context, r *BillingReminderSent) error

	// sweeps
	OrgsWithExpiryBetween(ctx context.Context, from, to time.Time) ([]Organization, error)
	PendingInvoicesIssuedBetween(ctx context.Context, from, to time.Time) ([]Invoice, error)
	ExpirePendingInvoices(ctx context.Context, before time.Time) ([]Invoice, error)

	// tx boundary (activation happens inside one transaction)
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// ---- services ----

type BillingService interface {
	// self-serve (org, billing:manage)
	ListPlanPrices(ctx context.Context) ([]PlanPrice, error)
	Checkout(ctx context.Context, dto CheckoutDTO) (*CheckoutResult, error)
	ListInvoices(ctx context.Context, q ListInvoicesQuery) ([]Invoice, int64, error)
	GetInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error)
	InvoicePDFURL(ctx context.Context, id uuid.UUID) (string, error)

	// gateway callback (no caller in ctx — invoked from a public redirect route)
	HandleCallback(ctx context.Context, gateway GatewayName, authority string, gatewayOK bool) (*Invoice, error)

	// worker
	GeneratePDF(ctx context.Context, invoiceID uuid.UUID) error
	RunReminderSweep(ctx context.Context, now time.Time) error
	ExpireStaleInvoices(ctx context.Context, now time.Time) error
}

type BillingAdminService interface {
	ListPrices(ctx context.Context) ([]PlanPrice, error)
	UpsertPrice(ctx context.Context, dto UpsertPlanPriceDTO) (*PlanPrice, error)
	DeactivatePrice(ctx context.Context, id uuid.UUID) error

	CreateInvoice(ctx context.Context, dto AdminCreateInvoiceDTO) (*Invoice, error)
	IssueInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error)
	MarkPaid(ctx context.Context, id uuid.UUID, dto AdminMarkPaidDTO) (*Invoice, error)
	CancelInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error)
	RefundInvoice(ctx context.Context, id uuid.UUID, dto AdminRefundDTO) (*Invoice, error)
	ListInvoices(ctx context.Context, q AdminListInvoicesQuery) ([]Invoice, int64, error)
	GetInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error)
}
