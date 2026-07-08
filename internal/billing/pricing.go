package billing

import "github.com/4H1R/zoora/internal/domain"

func lineAmount(quantity int, unitAmount int64) int64 {
	return int64(quantity) * unitAmount
}

// computeTotals returns (subtotal, taxAmount, total) in minor units. taxPercent
// is an integer percent; taxAmount uses integer division (rounds toward zero) —
// acceptable because taxPercent is 0 in v1 and prices are large Rial values.
func computeTotals(items []domain.InvoiceItem, taxPercent int) (subtotal, taxAmount, total int64) {
	for _, it := range items {
		subtotal += it.Amount
	}
	taxAmount = subtotal * int64(taxPercent) / 100
	total = subtotal + taxAmount
	return
}
