package billing

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
)

// fakeActivationRepo embeds the full BillingRepository interface so only the
// methods markPaidAndActivate touches need real bodies; any other call panics.
type fakeActivationRepo struct {
	domain.BillingRepository
	invoice *domain.Invoice
}

func (f *fakeActivationRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (f *fakeActivationRepo) FindInvoiceByIDForUpdate(_ context.Context, _ uuid.UUID) (*domain.Invoice, error) {
	return f.invoice, nil
}

func (f *fakeActivationRepo) UpdateInvoice(_ context.Context, _ *domain.Invoice) error { return nil }

type fakeOrgRepo struct {
	domain.OrganizationRepository
	org *domain.Organization
}

func (f *fakeOrgRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Organization, error) {
	return f.org, nil
}

type fakeActivator struct{ called bool }

func (f *fakeActivator) UpdatePlan(_ context.Context, _ uuid.UUID, _ domain.Plan, _ *time.Time) error {
	f.called = true
	return nil
}

type fakeCache struct{}

func (fakeCache) Invalidate(_ context.Context, _ uuid.UUID) error { return nil }

type fakeEnqueuer struct{}

func (fakeEnqueuer) EnqueuePDF(_ context.Context, _ uuid.UUID) error { return nil }

// auditSpy captures the records the service emits so tests can assert on them.
type auditSpy struct{ records []domain.AuditRecord }

func (a *auditSpy) Record(_ context.Context, r domain.AuditRecord) error {
	a.records = append(a.records, r)
	return nil
}

func (a *auditSpy) RecordDenied(_ context.Context, _ domain.AuditRecord) error { return nil }

func TestMarkPaidAndActivate_RecordsPlanChangeAudit(t *testing.T) {
	orgID := uuid.New()
	plan := domain.PlanKey(domain.TierPro, 200)
	interval := domain.BillingIntervalMonthly

	inv := &domain.Invoice{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Status:         domain.InvoiceStatusPending,
		Items: []domain.InvoiceItem{{
			Kind:     domain.InvoiceItemPlanSubscription,
			Plan:     &plan,
			Interval: &interval,
		}},
	}
	org := &domain.Organization{ID: orgID, Plan: domain.PlanFree}
	activator := &fakeActivator{}
	spy := &auditSpy{}

	s := &service{
		repo:      &fakeActivationRepo{invoice: inv},
		orgRepo:   &fakeOrgRepo{org: org},
		activator: activator,
		cache:     fakeCache{},
		queue:     fakeEnqueuer{},
		audit:     spy,
		logger:    slog.Default(),
	}

	// No Caller in ctx: this is the public gateway-callback path. The recorder
	// must still file the entry under the target org (System actor).
	_, err := s.markPaidAndActivate(context.Background(), inv.ID, time.Now())
	require.NoError(t, err)
	require.True(t, activator.called)

	require.Len(t, spy.records, 1)
	rec := spy.records[0]
	assert.Equal(t, domain.AuditUpdated, rec.Action)
	assert.Equal(t, domain.AuditTargetBilling, rec.TargetType)
	require.NotNil(t, rec.TargetID)
	assert.Equal(t, inv.ID, *rec.TargetID)
	assert.Equal(t, string(plan), rec.TargetLabel)
	require.NotNil(t, rec.OrgID)
	assert.Equal(t, orgID, *rec.OrgID)
	assert.Equal(t, string(domain.PlanFree), rec.Metadata["from_plan"])
	assert.Equal(t, string(plan), rec.Metadata["to_plan"])
}
