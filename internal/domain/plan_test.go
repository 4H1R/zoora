package domain

import (
	"testing"
	"time"
)

func TestPlanCatalogHasThreeTiers(t *testing.T) {
	for _, p := range []Plan{PlanFree, PlanPro, PlanEnterprise} {
		if _, ok := PlanCatalog[p]; !ok {
			t.Fatalf("catalog missing plan %q", p)
		}
	}
}

func TestFreePlanGates(t *testing.T) {
	ent := PlanCatalog[PlanFree]
	if ent.Can(FeatureRecording) {
		t.Fatal("Free must not have recording")
	}
	if got := ent.Limit(LimitMaxUsers); got != 10 {
		t.Fatalf("Free max users = %d, want 10", got)
	}
}

func TestEnterpriseUnlimitedUsers(t *testing.T) {
	ent := PlanCatalog[PlanEnterprise]
	if got := ent.Limit(LimitMaxUsers); got != 0 {
		t.Fatalf("Enterprise max users = %d, want 0 (unlimited)", got)
	}
	if !ent.Unlimited(LimitMaxUsers) {
		t.Fatal("Enterprise users must be unlimited")
	}
}

func TestZeroValueEntitlementsFailClosed(t *testing.T) {
	// A Caller built without middleware (tests, missed wiring) carries a
	// zero-value Entitlements. With the 0-means-unlimited convention that
	// would fail OPEN on limits — guard against it: zero value behaves as Free.
	var ent Entitlements
	if ent.Can(FeatureRecording) {
		t.Fatal("zero-value must not grant features")
	}
	if ent.Unlimited(LimitMaxUsers) {
		t.Fatal("zero-value must not be unlimited")
	}
	if ent.Within(LimitMaxUsers, 10) {
		t.Fatal("zero-value must enforce the Free ceiling (10)")
	}
	if !ent.Within(LimitMaxUsers, 5) {
		t.Fatal("zero-value under the Free ceiling must allow")
	}
}

func TestValidPlan(t *testing.T) {
	if !PlanFree.Valid() {
		t.Fatal("free is valid")
	}
	if Plan("bogus").Valid() {
		t.Fatal("bogus is invalid")
	}
}

func TestEffectiveEntitlementsExpiredDowngradesToFree(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	ent := EffectiveEntitlements(PlanPro, &past, now)
	if ent.Plan != PlanFree {
		t.Fatalf("expired Pro should be Free, got %s", ent.Plan)
	}
}

func TestEffectiveEntitlementsActiveKeepsPlan(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	ent := EffectiveEntitlements(PlanPro, &future, now)
	if ent.Plan != PlanPro || !ent.Can(FeatureRecording) {
		t.Fatal("active Pro should keep recording")
	}
}

func TestEffectiveEntitlementsNilExpiryIsPerpetual(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	ent := EffectiveEntitlements(PlanEnterprise, nil, now)
	if ent.Plan != PlanEnterprise {
		t.Fatal("nil expiry = perpetual, must keep plan")
	}
}

func TestEffectiveEntitlementsUnknownPlanFallsBackToFree(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	ent := EffectiveEntitlements(Plan("bogus"), nil, now)
	if ent.Plan != PlanFree {
		t.Fatal("unknown plan must resolve to Free")
	}
}
