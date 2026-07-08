package domain

import (
	"testing"
	"time"
)

func TestPlanCatalogHasAllTierSizeCombos(t *testing.T) {
	if got, want := len(PlanCatalog), len(PlanTiers)*len(PlanSizes); got != want {
		t.Fatalf("catalog has %d plans, want %d", got, want)
	}
	for _, tier := range PlanTiers {
		for _, size := range PlanSizes {
			if _, ok := PlanCatalog[PlanKey(tier, size)]; !ok {
				t.Fatalf("catalog missing plan %q", PlanKey(tier, size))
			}
		}
	}
}

func TestPlanTierAndSize(t *testing.T) {
	p := PlanKey(TierPlus, 1000)
	if p != Plan("plus_1000") {
		t.Fatalf("key = %q, want plus_1000", p)
	}
	if p.Tier() != TierPlus || p.Size() != 1000 {
		t.Fatalf("parsed tier=%q size=%d", p.Tier(), p.Size())
	}
	if Plan("legacy").Tier() != "" || Plan("legacy").Size() != 0 {
		t.Fatal("malformed plan must parse to zero values")
	}
}

func TestFreePlanGates(t *testing.T) {
	ent := PlanCatalog[PlanFree]
	for _, f := range []Feature{FeatureRecording, FeatureWhiteboard, FeatureChat, FeatureConnectors, FeatureAutoGrading, FeatureAI, FeatureAdvancedAntiCheat} {
		if ent.Can(f) {
			t.Fatalf("Free must not have %s", f)
		}
	}
	if !ent.Can(FeatureOfflineRooms) {
		t.Fatal("Free must have offline rooms")
	}
	if got := ent.Limit(LimitMaxUsers); got != 50 {
		t.Fatalf("free_50 max users = %d, want 50", got)
	}
	if got := ent.Limit(LimitStorageGB); got != 2 {
		t.Fatalf("free_50 storage = %d, want 2", got)
	}
	if got := ent.Limit(LimitMaxParticipants); got != 5 {
		t.Fatalf("free_50 participants = %d, want 5", got)
	}
	if got := ent.Limit(LimitConcurrentRooms); got != 1 {
		t.Fatalf("free_50 rooms = %d, want 1", got)
	}
}

func TestTierFeatureMatrix(t *testing.T) {
	plus := PlanCatalog[PlanKey(TierPlus, 50)]
	pro := PlanCatalog[PlanKey(TierPro, 50)]
	max := PlanCatalog[PlanKey(TierMax, 50)]

	if plus.Can(FeatureWhiteboard) || plus.Can(FeatureChat) || plus.Can(FeatureRecording) {
		t.Fatal("Plus must not have whiteboard/chat/recording")
	}
	for _, f := range []Feature{FeatureWhiteboard, FeatureChat, FeatureConnectors, FeatureAI, FeatureAdvancedAntiCheat} {
		if !pro.Can(f) {
			t.Fatalf("Pro must have %s", f)
		}
	}
	if pro.Can(FeatureRecording) || pro.Can(FeatureAutoGrading) {
		t.Fatal("Pro must not have recording/auto-grading")
	}
	for _, f := range AllFeatures {
		if !max.Can(f) {
			t.Fatalf("Max must have %s", f)
		}
	}
}

func TestLimitsScaleWithSize(t *testing.T) {
	pro200 := PlanCatalog[PlanKey(TierPro, 200)]
	if got := pro200.Limit(LimitMaxUsers); got != 200 {
		t.Fatalf("pro_200 users = %d, want 200", got)
	}
	if got := pro200.Limit(LimitStorageGB); got != 100 {
		t.Fatalf("pro_200 storage = %d, want 100 (25 × 200/50)", got)
	}
	if got := pro200.Limit(LimitMaxParticipants); got != 80 {
		t.Fatalf("pro_200 participants = %d, want 80", got)
	}
	if got := pro200.Limit(LimitConcurrentRooms); got != 20 {
		t.Fatalf("pro_200 rooms = %d, want 20", got)
	}
	// Retention does not scale.
	max1000 := PlanCatalog[PlanKey(TierMax, 1000)]
	if got := max1000.Limit(LimitRecordingRetentionDays); got != 365 {
		t.Fatalf("max_1000 retention = %d, want 365", got)
	}
}

func TestZeroCeilingMeansUnlimited(t *testing.T) {
	ent := Entitlements{Plan: "synthetic", features: map[Feature]bool{}, limits: map[Limit]int64{LimitMaxUsers: 0}}
	if !ent.Unlimited(LimitMaxUsers) || !ent.Within(LimitMaxUsers, 1_000_000) {
		t.Fatal("0 ceiling must mean unlimited")
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
	if ent.Within(LimitMaxUsers, 50) {
		t.Fatal("zero-value must enforce the free_50 ceiling (50)")
	}
	if !ent.Within(LimitMaxUsers, 5) {
		t.Fatal("zero-value under the free_50 ceiling must allow")
	}
}

func TestValidPlan(t *testing.T) {
	if !PlanFree.Valid() {
		t.Fatal("free_50 is valid")
	}
	if !PlanKey(TierMax, 1000).Valid() {
		t.Fatal("max_1000 is valid")
	}
	if Plan("bogus").Valid() {
		t.Fatal("bogus is invalid")
	}
	if Plan("pro").Valid() {
		t.Fatal("legacy sizeless key is invalid")
	}
	if Plan("pro_75").Valid() {
		t.Fatal("unknown size is invalid")
	}
}

func TestPublicCatalogOrderedBySizeThenTier(t *testing.T) {
	cat := PublicCatalog()
	if len(cat) != len(PlanTiers)*len(PlanSizes) {
		t.Fatalf("public catalog has %d entries, want %d", len(cat), len(PlanTiers)*len(PlanSizes))
	}
	if cat[0].Plan != PlanFree || cat[1].Plan != PlanKey(TierPlus, 50) {
		t.Fatalf("unexpected order: %s, %s", cat[0].Plan, cat[1].Plan)
	}
	last := cat[len(cat)-1]
	if last.Plan != PlanKey(TierMax, 1000) || last.Tier != TierMax || last.Size != 1000 {
		t.Fatalf("unexpected last entry: %+v", last)
	}
}

func TestEffectiveEntitlementsExpiredDowngradesToFree(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	ent := EffectiveEntitlements(PlanKey(TierPro, 200), &past, now)
	if ent.Plan != PlanFree {
		t.Fatalf("expired pro_200 should be Free, got %s", ent.Plan)
	}
}

func TestEffectiveEntitlementsActiveKeepsPlan(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	ent := EffectiveEntitlements(PlanKey(TierMax, 100), &future, now)
	if ent.Plan != PlanKey(TierMax, 100) || !ent.Can(FeatureRecording) {
		t.Fatal("active max_100 should keep recording")
	}
}

func TestEffectiveEntitlementsNilExpiryIsPerpetual(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	ent := EffectiveEntitlements(PlanKey(TierMax, 500), nil, now)
	if ent.Plan != PlanKey(TierMax, 500) {
		t.Fatal("nil expiry = perpetual, must keep plan")
	}
}

func TestEffectiveEntitlementsUnknownPlanFallsBackToFree(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	ent := EffectiveEntitlements(Plan("enterprise"), nil, now)
	if ent.Plan != PlanFree {
		t.Fatal("unknown plan must resolve to Free")
	}
}
