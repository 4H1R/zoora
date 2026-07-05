package domain

import "time"

// Plan is a code-defined subscription tier. The catalog (PlanCatalog) is the
// single source of truth for what each tier includes; orgs store only the plan
// key + an expiry.
type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanEnterprise Plan = "enterprise"
)

func (p Plan) Valid() bool {
	_, ok := PlanCatalog[p]
	return ok
}

// Feature is a boolean capability gate key.
type Feature string

const (
	FeatureRecording         Feature = "recording"
	FeatureOfflineRooms      Feature = "offline_rooms"
	FeatureAdvancedAntiCheat Feature = "advanced_anticheat"
	FeatureCustomRoles       Feature = "custom_roles"
	FeatureSSO               Feature = "sso"
)

// Limit is a numeric quota key. A stored value of 0 means unlimited.
type Limit string

const (
	LimitMaxUsers               Limit = "max_users"
	LimitMaxParticipants        Limit = "max_participants"
	LimitStorageGB              Limit = "storage_gb"
	LimitConcurrentRooms        Limit = "concurrent_rooms"
	LimitRecordingRetentionDays Limit = "recording_retention_days"
)

// Entitlements is the resolved capability set for a plan. Value type — cheap to
// copy into Caller.
type Entitlements struct {
	Plan     Plan
	features map[Feature]bool
	limits   map[Limit]int64
}

func (e Entitlements) Can(f Feature) bool { return e.resolved().features[f] }

// resolved guards against a zero-value Entitlements: with the 0-means-unlimited
// convention, nil maps would make every limit check fail OPEN (Limit()=0 →
// unlimited). A zero value (Caller built without middleware) resolves to Free.
func (e Entitlements) resolved() Entitlements {
	if e.limits == nil {
		return PlanCatalog[PlanFree]
	}
	return e
}

// Limit returns the ceiling for l (0 = unlimited).
func (e Entitlements) Limit(l Limit) int64 { return e.resolved().limits[l] }

// Unlimited reports whether limit l has no ceiling.
func (e Entitlements) Unlimited(l Limit) bool { return e.resolved().limits[l] == 0 }

// Within reports whether adding one more (current -> current+1) is allowed.
// current is the live count of existing resources.
func (e Entitlements) Within(l Limit, current int64) bool {
	ceiling := e.resolved().limits[l]
	if ceiling == 0 {
		return true // unlimited
	}
	return current < ceiling
}

// PlanCatalog is the source of truth. Editing a tier = editing this map.
var PlanCatalog = map[Plan]Entitlements{
	PlanFree: {
		Plan: PlanFree,
		features: map[Feature]bool{
			FeatureRecording: false, FeatureOfflineRooms: false,
			FeatureAdvancedAntiCheat: false, FeatureCustomRoles: false, FeatureSSO: false,
		},
		limits: map[Limit]int64{
			LimitMaxUsers: 10, LimitMaxParticipants: 25, LimitStorageGB: 2,
			LimitConcurrentRooms: 1, LimitRecordingRetentionDays: 0,
		},
	},
	PlanPro: {
		Plan: PlanPro,
		features: map[Feature]bool{
			FeatureRecording: true, FeatureOfflineRooms: true,
			FeatureAdvancedAntiCheat: true, FeatureCustomRoles: true, FeatureSSO: false,
		},
		limits: map[Limit]int64{
			LimitMaxUsers: 150, LimitMaxParticipants: 100, LimitStorageGB: 100,
			LimitConcurrentRooms: 10, LimitRecordingRetentionDays: 30,
		},
	},
	PlanEnterprise: {
		Plan: PlanEnterprise,
		features: map[Feature]bool{
			FeatureRecording: true, FeatureOfflineRooms: true,
			FeatureAdvancedAntiCheat: true, FeatureCustomRoles: true, FeatureSSO: true,
		},
		limits: map[Limit]int64{
			LimitMaxUsers: 0, LimitMaxParticipants: 300, LimitStorageGB: 0,
			LimitConcurrentRooms: 0, LimitRecordingRetentionDays: 365,
		},
	},
}

// PlanInfo is the public, serializable shape of a plan's entitlements, for the
// admin UI plan picker. The internal Entitlements maps are unexported, so this
// projects them into a stable JSON contract.
type PlanInfo struct {
	Plan     Plan             `json:"plan"`
	Features map[Feature]bool `json:"features"`
	Limits   map[Limit]int64  `json:"limits"`
}

// PublicCatalog returns the plan catalog in tier order for API responses.
func PublicCatalog() []PlanInfo {
	order := []Plan{PlanFree, PlanPro, PlanEnterprise}
	out := make([]PlanInfo, 0, len(order))
	for _, p := range order {
		ent := PlanCatalog[p]
		out = append(out, PlanInfo{Plan: p, Features: ent.features, Limits: ent.limits})
	}
	return out
}

// EffectiveEntitlements resolves the entitlements an org actually gets right now:
// the stored plan while active, downgraded to Free once expiresAt has passed.
// A nil expiresAt means perpetual. An unknown plan resolves to Free.
func EffectiveEntitlements(plan Plan, expiresAt *time.Time, now time.Time) Entitlements {
	ent, ok := PlanCatalog[plan]
	if !ok {
		return PlanCatalog[PlanFree]
	}
	if expiresAt != nil && now.After(*expiresAt) {
		return PlanCatalog[PlanFree]
	}
	return ent
}
