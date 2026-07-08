package billing

import (
	"testing"
	"time"

	"github.com/4H1R/zoora/internal/domain"
)

func TestRenewalStageFor(t *testing.T) {
	exp := mustDate2("2026-07-15") // expiry
	cases := []struct {
		now  string
		want domain.BillingReminderKind
		ok   bool
	}{
		{"2026-07-08", domain.ReminderRenewal7d, true},    // 7 days before
		{"2026-07-12", domain.ReminderRenewal3d, true},    // 3 days before
		{"2026-07-15", domain.ReminderRenewalDue, true},   // due day
		{"2026-07-18", domain.ReminderRenewalGrace, true}, // 3 days after
		{"2026-07-10", "", false},                         // no stage on a gap day
	}
	for _, c := range cases {
		kind, ok := renewalStageFor(exp, mustDate2(c.now))
		if ok != c.ok || kind != c.want {
			t.Errorf("now=%s got (%q,%v) want (%q,%v)", c.now, kind, ok, c.want, c.ok)
		}
	}
}

func mustDate2(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
