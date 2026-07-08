//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/billing"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
	"github.com/4H1R/zoora/tests/testutil"
)

func setupBillingRepo(t *testing.T) domain.BillingRepository {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.PlanPrice{},
		&domain.Invoice{},
		&domain.InvoiceItem{},
		&domain.Payment{},
		&domain.BillingReminderSent{},
	))
	return billing.NewRepository(db)
}

func TestBillingRepository_PriceRoundTrip(t *testing.T) {
	repo := setupBillingRepo(t)
	ctx := context.Background()

	price := factory.NewPlanPrice(domain.PlanKey(domain.TierPro, 50), domain.BillingIntervalMonthly)
	require.NoError(t, repo.UpsertPrice(ctx, price))

	got, err := repo.FindActivePrice(ctx, domain.PlanKey(domain.TierPro, 50), domain.BillingIntervalMonthly, domain.CurrencyIRR)
	require.NoError(t, err)
	assert.Equal(t, price.Amount, got.Amount)

	// Upserting again deactivates the old row and keeps a single active one.
	newer := factory.NewPlanPrice(domain.PlanKey(domain.TierPro, 50), domain.BillingIntervalMonthly, func(p *domain.PlanPrice) {
		p.Amount = 2_000_000
	})
	require.NoError(t, repo.UpsertPrice(ctx, newer))

	got, err = repo.FindActivePrice(ctx, domain.PlanKey(domain.TierPro, 50), domain.BillingIntervalMonthly, domain.CurrencyIRR)
	require.NoError(t, err)
	assert.Equal(t, int64(2_000_000), got.Amount)

	// Exactly one active row survives for the (plan,interval,currency) tuple.
	active, err := repo.ListActivePrices(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 1)
}
