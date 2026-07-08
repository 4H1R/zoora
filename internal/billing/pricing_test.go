package billing

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

func TestComputeTotals(t *testing.T) {
	// two items: 1× 1,500,000 + 2× 500,000 = 2,500,000 subtotal; 10% tax.
	items := []domain.InvoiceItem{
		{Quantity: 1, UnitAmount: 1_500_000, Amount: 1_500_000},
		{Quantity: 2, UnitAmount: 500_000, Amount: 1_000_000},
	}
	sub, tax, total := computeTotals(items, 10)
	if sub != 2_500_000 {
		t.Errorf("subtotal = %d, want 2500000", sub)
	}
	if tax != 250_000 {
		t.Errorf("tax = %d, want 250000", tax)
	}
	if total != 2_750_000 {
		t.Errorf("total = %d, want 2750000", total)
	}
}

func TestLineAmount(t *testing.T) {
	if got := lineAmount(3, 400_000); got != 1_200_000 {
		t.Errorf("lineAmount = %d, want 1200000", got)
	}
}
