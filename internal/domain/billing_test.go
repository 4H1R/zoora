package domain

import (
	"testing"
	"time"
)

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestNextPlanState(t *testing.T) {
	now := mustTime("2026-07-08T00:00:00Z")
	future := mustTime("2026-08-01T00:00:00Z") // active expiry ahead of now
	past := mustTime("2026-06-01T00:00:00Z")   // already expired

	pro50 := PlanKey(TierPro, 50)
	pro200 := PlanKey(TierPro, 200)
	max50 := PlanKey(TierMax, 50)
	plus1000 := PlanKey(TierPlus, 1000)

	tests := []struct {
		name       string
		curPlan    Plan
		curExpiry  *time.Time
		buy        Plan
		interval   BillingInterval
		wantPlan   Plan
		wantExpiry time.Time
		wantErr    bool
	}{
		{
			name: "free to pro monthly", curPlan: PlanFree, curExpiry: nil,
			buy: pro50, interval: BillingIntervalMonthly,
			wantPlan: pro50, wantExpiry: now.AddDate(0, 1, 0),
		},
		{
			name: "same plan active extends from expiry", curPlan: pro50, curExpiry: &future,
			buy: pro50, interval: BillingIntervalMonthly,
			wantPlan: pro50, wantExpiry: future.AddDate(0, 1, 0),
		},
		{
			name: "same plan expired extends from now", curPlan: pro50, curExpiry: &past,
			buy: pro50, interval: BillingIntervalYearly,
			wantPlan: pro50, wantExpiry: now.AddDate(1, 0, 0),
		},
		{
			name: "tier upgrade active resets from now", curPlan: pro50, curExpiry: &future,
			buy: max50, interval: BillingIntervalMonthly,
			wantPlan: max50, wantExpiry: now.AddDate(0, 1, 0),
		},
		{
			name: "size upgrade within tier resets from now", curPlan: pro50, curExpiry: &future,
			buy: pro200, interval: BillingIntervalMonthly,
			wantPlan: pro200, wantExpiry: now.AddDate(0, 1, 0),
		},
		{
			name: "higher tier outranks bigger size of lower tier", curPlan: plus1000, curExpiry: &future,
			buy: pro50, interval: BillingIntervalMonthly,
			wantPlan: pro50, wantExpiry: now.AddDate(0, 1, 0),
		},
		{
			name: "tier downgrade while active is blocked", curPlan: max50, curExpiry: &future,
			buy: pro50, interval: BillingIntervalMonthly,
			wantErr: true,
		},
		{
			name: "size downgrade while active is blocked", curPlan: pro200, curExpiry: &future,
			buy: pro50, interval: BillingIntervalMonthly,
			wantErr: true,
		},
		{
			name: "downgrade allowed once expired", curPlan: max50, curExpiry: &past,
			buy: pro50, interval: BillingIntervalMonthly,
			wantPlan: pro50, wantExpiry: now.AddDate(0, 1, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, expiry, err := NextPlanState(tt.curPlan, tt.curExpiry, tt.buy, tt.interval, now)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if plan != tt.wantPlan {
				t.Errorf("plan = %s, want %s", plan, tt.wantPlan)
			}
			if !expiry.Equal(tt.wantExpiry) {
				t.Errorf("expiry = %s, want %s", expiry, tt.wantExpiry)
			}
		})
	}
}
