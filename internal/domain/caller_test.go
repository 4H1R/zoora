package domain

import "testing"

func TestCallerEntitlementHelpers(t *testing.T) {
	c := Caller{Ent: PlanCatalog[PlanFree]}
	if c.HasFeature(FeatureRecording) {
		t.Fatal("free caller cannot record")
	}
	if c.Limit(LimitMaxUsers) != 10 {
		t.Fatal("free caller max users = 10")
	}
}
