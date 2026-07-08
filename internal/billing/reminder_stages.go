package billing

import (
	"time"

	"github.com/4H1R/zoora/internal/domain"
)

// renewalStageFor returns the reminder stage due today for an expiry date, or
// ("",false) if today isn't a reminder day. Compares whole days (UTC date).
func renewalStageFor(expiry, now time.Time) (domain.BillingReminderKind, bool) {
	days := daysBetween(dateOnly(now), dateOnly(expiry)) // >0 = expiry ahead
	switch days {
	case 7:
		return domain.ReminderRenewal7d, true
	case 3:
		return domain.ReminderRenewal3d, true
	case 0:
		return domain.ReminderRenewalDue, true
	case -3:
		return domain.ReminderRenewalGrace, true
	default:
		return "", false
	}
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func daysBetween(from, to time.Time) int {
	return int(to.Sub(from).Hours() / 24)
}

// reminderPeriodKey scopes dedup to this expiry instant so a renewed plan (new
// expiry) can re-fire the same stage next cycle.
func reminderPeriodKey(expiry time.Time) string {
	return expiry.UTC().Format("2006-01-02")
}
