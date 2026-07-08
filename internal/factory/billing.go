package factory

import (
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewPlanPrice(plan domain.Plan, interval domain.BillingInterval, opts ...func(*domain.PlanPrice)) *domain.PlanPrice {
	p := &domain.PlanPrice{
		Plan:     plan,
		Interval: interval,
		Currency: domain.CurrencyIRR,
		Amount:   1_500_000, // 150,000 Toman
		Active:   true,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func NewInvoice(orgID uuid.UUID, opts ...func(*domain.Invoice)) *domain.Invoice {
	inv := &domain.Invoice{
		OrganizationID: orgID,
		Status:         domain.InvoiceStatusDraft,
		Currency:       domain.CurrencyIRR,
		Subtotal:       1_500_000,
		Total:          1_500_000,
		Description:    T("Pro plan — monthly", "پلن حرفه‌ای — ماهانه"),
	}
	for _, o := range opts {
		o(inv)
	}
	return inv
}

func NewInvoiceItem(invoiceID uuid.UUID, opts ...func(*domain.InvoiceItem)) *domain.InvoiceItem {
	plan := domain.PlanPro
	interval := domain.BillingIntervalMonthly
	item := &domain.InvoiceItem{
		InvoiceID:   invoiceID,
		Kind:        domain.InvoiceItemPlanSubscription,
		Description: T("Pro plan — monthly", "پلن حرفه‌ای — ماهانه"),
		Plan:        &plan,
		Interval:    &interval,
		Quantity:    1,
		UnitAmount:  1_500_000,
		Amount:      1_500_000,
		Currency:    domain.CurrencyIRR,
	}
	for _, o := range opts {
		o(item)
	}
	return item
}

func NewPayment(invoiceID, orgID uuid.UUID, opts ...func(*domain.Payment)) *domain.Payment {
	p := &domain.Payment{
		InvoiceID:      invoiceID,
		OrganizationID: orgID,
		Gateway:        domain.GatewayZarinpal,
		Status:         domain.PaymentStatusPending,
		Amount:         1_500_000,
		Currency:       domain.CurrencyIRR,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}
