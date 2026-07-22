package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// markPaidAndActivate transitions inv to paid and activates the org plan, all
// in one transaction. Idempotent: if inv is already paid it is a no-op (returns
// the invoice unchanged).
//
// The function re-checks the invoice status under a row lock to close the
// double-callback race: two concurrent gateway callbacks cannot both read
// status 'pending' and both call UpdatePlan.
func (s *service) markPaidAndActivate(ctx context.Context, invoiceID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	var result *domain.Invoice
	err := s.repo.WithTx(ctx, func(ctx context.Context) error {
		inv, err := s.repo.FindInvoiceByIDForUpdate(ctx, invoiceID)
		if err != nil {
			return err
		}
		// Idempotency guard: already settled -> no-op.
		if inv.Status == domain.InvoiceStatusPaid {
			result = inv
			return nil
		}
		if inv.Status != domain.InvoiceStatusPending && inv.Status != domain.InvoiceStatusDraft {
			return domain.ErrInvoiceNotPayable
		}

		// Resolve org + compute next plan state from the plan-subscription line.
		org, err := s.orgRepo.FindByID(ctx, inv.OrganizationID)
		if err != nil {
			return err
		}
		if planItem := findPlanItem(inv.Items); planItem != nil {
			plan, expiry, err := domain.NextPlanState(org.Plan, org.PlanExpiresAt, *planItem.Plan, *planItem.Interval, now)
			if err != nil {
				return err
			}
			if err := s.activator.UpdatePlan(ctx, org.ID, plan, &expiry); err != nil {
				return fmt.Errorf("billing.markPaidAndActivate.UpdatePlan: %w", err)
			}
			// Plan activation commits worker/callback-side where there may be no
			// Caller (public gateway callback) — set OrgID explicitly to the
			// target org so the recorder files a System-actor entry (or the admin
			// actor on the manual mark-paid path). Runs in the same tx as UpdatePlan.
			orgID := org.ID
			if err := s.audit.Record(ctx, domain.AuditRecord{
				Action:      domain.AuditUpdated,
				TargetType:  domain.AuditTargetBilling,
				TargetID:    &inv.ID,
				TargetLabel: string(plan),
				OrgID:       &orgID,
				Metadata: map[string]any{
					"from_plan":  string(org.Plan),
					"to_plan":    string(plan),
					"invoice_id": inv.ID.String(),
					"expires_at": expiry.Format(time.RFC3339),
				},
			}); err != nil {
				return err
			}
		}

		inv.Status = domain.InvoiceStatusPaid
		inv.PaidAt = &now
		inv.UpdatedAt = now
		if err := s.repo.UpdateInvoice(ctx, inv); err != nil {
			return err
		}
		result = inv
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Post-commit side effects (safe to repeat, outside the tx).
	if result != nil && result.PaidAt != nil && result.PaidAt.Equal(now) {
		if err := s.cache.Invalidate(ctx, result.OrganizationID); err != nil {
			s.logger.Error("billing: entitlements cache invalidate failed", "org_id", result.OrganizationID, "error", err)
		}
		if err := s.queue.EnqueuePDF(ctx, result.ID); err != nil {
			s.logger.Error("billing: enqueue pdf failed", "invoice_id", result.ID, "error", err)
		}
	}
	return result, nil
}

// findPlanItem returns the first plan-subscription line with a plan+interval set.
func findPlanItem(items []domain.InvoiceItem) *domain.InvoiceItem {
	for i := range items {
		it := &items[i]
		if it.Kind == domain.InvoiceItemPlanSubscription && it.Plan != nil && it.Interval != nil {
			return it
		}
	}
	return nil
}
