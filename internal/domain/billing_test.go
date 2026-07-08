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
			buy: PlanPro, interval: BillingIntervalMonthly,
			wantPlan: PlanPro, wantExpiry: now.AddDate(0, 1, 0),
		},
		{
			name: "same tier active extends from expiry", curPlan: PlanPro, curExpiry: &future,
			buy: PlanPro, interval: BillingIntervalMonthly,
			wantPlan: PlanPro, wantExpiry: future.AddDate(0, 1, 0),
		},
		{
			name: "same tier expired extends from now", curPlan: PlanPro, curExpiry: &past,
			buy: PlanPro, interval: BillingIntervalYearly,
			wantPlan: PlanPro, wantExpiry: now.AddDate(1, 0, 0),
		},
		{
			name: "upgrade active resets from now", curPlan: PlanPro, curExpiry: &future,
			buy: PlanEnterprise, interval: BillingIntervalMonthly,
			wantPlan: PlanEnterprise, wantExpiry: now.AddDate(0, 1, 0),
		},
		{
			name: "downgrade while active is blocked", curPlan: PlanEnterprise, curExpiry: &future,
			buy: PlanPro, interval: BillingIntervalMonthly,
			wantErr: true,
		},
		{
			name: "downgrade allowed once expired", curPlan: PlanEnterprise, curExpiry: &past,
			buy: PlanPro, interval: BillingIntervalMonthly,
			wantPlan: PlanPro, wantExpiry: now.AddDate(0, 1, 0),
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
