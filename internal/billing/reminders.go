package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// RunReminderSweep sends renewal reminders (7d/3d/due/grace) and unpaid-invoice
// reminders (24h/72h), deduped via billing_reminders_sent. Called daily by the
// scheduler.
func (s *service) RunReminderSweep(ctx context.Context, now time.Time) error {
	// Renewal reminders: look at orgs expiring within [-3d, +7d] of now.
	from := now.AddDate(0, 0, -4)
	to := now.AddDate(0, 0, 8)
	orgs, err := s.repo.OrgsWithExpiryBetween(ctx, from, to)
	if err != nil {
		return err
	}
	for _, org := range orgs {
		if org.PlanExpiresAt == nil {
			continue
		}
		kind, ok := renewalStageFor(*org.PlanExpiresAt, now)
		if !ok {
			continue
		}
		periodKey := reminderPeriodKey(*org.PlanExpiresAt)
		if err := s.sendReminderOnce(ctx, kind, org.ID, periodKey, renewalTitle(kind), renewalBody(kind, org), s.renewActionURL()); err != nil {
			s.logger.Error("billing: renewal reminder failed", "org_id", org.ID, "kind", kind, "error", err)
		}
	}

	// Unpaid-invoice reminders: pending invoices issued ~24h and ~72h ago.
	if err := s.unpaidReminders(ctx, now, domain.ReminderInvoiceUnpaid24h, 24*time.Hour); err != nil {
		s.logger.Error("billing: unpaid-24h sweep failed", "error", err)
	}
	if err := s.unpaidReminders(ctx, now, domain.ReminderInvoiceUnpaid72h, 72*time.Hour); err != nil {
		s.logger.Error("billing: unpaid-72h sweep failed", "error", err)
	}
	return nil
}

func (s *service) unpaidReminders(ctx context.Context, now time.Time, kind domain.BillingReminderKind, age time.Duration) error {
	// A ~1d window around the target age so a daily/hourly sweep catches it once.
	target := now.Add(-age)
	from := target.Add(-12 * time.Hour)
	to := target.Add(12 * time.Hour)
	invs, err := s.repo.PendingInvoicesIssuedBetween(ctx, from, to)
	if err != nil {
		return err
	}
	for _, inv := range invs {
		if err := s.sendReminderOnce(ctx, kind, inv.ID, inv.ID.String(), unpaidTitle(), unpaidBody(inv), s.invoiceActionURL(inv)); err != nil {
			s.logger.Error("billing: unpaid reminder failed", "invoice_id", inv.ID, "error", err)
		}
	}
	return nil
}

// sendReminderOnce dedups then sends a SYSTEM notification to the org's members.
func (s *service) sendReminderOnce(ctx context.Context, kind domain.BillingReminderKind, subjectID uuid.UUID, periodKey, title, body, actionURL string) error {
	already, err := s.repo.ReminderAlreadySent(ctx, kind, subjectID, periodKey)
	if err != nil {
		return err
	}
	if already {
		return nil
	}
	// Audience = the org (fan-out targets org members). For renewal reminders the
	// subjectID is the org id; for unpaid-invoice reminders it is an invoice id,
	// so resolve its org for the audience.
	orgID := subjectID
	if kind == domain.ReminderInvoiceUnpaid24h || kind == domain.ReminderInvoiceUnpaid72h {
		inv, err := s.repo.FindInvoiceByID(ctx, subjectID)
		if err != nil {
			return err
		}
		orgID = inv.OrganizationID
	}
	action := actionURL
	if err := s.notifier.SendSystem(ctx, domain.SystemNotificationInput{
		OrganizationID: &orgID,
		Category:       domain.NotificationCategoryReminder,
		Title:          title,
		Body:           body,
		ActionURL:      &action,
		Audience:       domain.NotificationAudience{Type: domain.AudienceOrg, OrgID: &orgID},
	}); err != nil {
		return err
	}
	return s.repo.MarkReminderSent(ctx, &domain.BillingReminderSent{Kind: kind, SubjectID: subjectID, PeriodKey: periodKey})
}

// ExpireStaleInvoices flips pending invoices past their deadline to expired.
func (s *service) ExpireStaleInvoices(ctx context.Context, now time.Time) error {
	expired, err := s.repo.ExpirePendingInvoices(ctx, now)
	if err != nil {
		return err
	}
	if len(expired) > 0 {
		s.logger.Info("billing: expired stale pending invoices", "count", len(expired))
	}
	return nil
}

func (s *service) renewActionURL() string { return s.cfg.AppBaseURL + "/org/billing" }
func (s *service) invoiceActionURL(inv domain.Invoice) string {
	return fmt.Sprintf("%s/org/billing/invoices/%s", s.cfg.AppBaseURL, inv.ID)
}

func renewalTitle(kind domain.BillingReminderKind) string {
	if kind == domain.ReminderRenewalGrace {
		return "اشتراک شما منقضی شده است"
	}
	return "یادآوری تمدید اشتراک"
}

func renewalBody(kind domain.BillingReminderKind, org domain.Organization) string {
	switch kind {
	case domain.ReminderRenewal7d:
		return "اشتراک سازمان شما تا ۷ روز دیگر منقضی می‌شود. برای جلوگیری از قطع سرویس، آن را تمدید کنید."
	case domain.ReminderRenewal3d:
		return "تنها ۳ روز تا انقضای اشتراک سازمان شما باقی مانده است."
	case domain.ReminderRenewalDue:
		return "اشتراک سازمان شما امروز منقضی می‌شود."
	default:
		return "اشتراک سازمان شما منقضی شده و به پلن رایگان تغییر یافته است. برای بازیابی امکانات، تمدید کنید."
	}
}

func unpaidTitle() string { return "فاکتور پرداخت‌نشده" }
func unpaidBody(inv domain.Invoice) string {
	return "شما یک فاکتور پرداخت‌نشده دارید. برای فعال‌سازی اشتراک، پرداخت را تکمیل کنید."
}
