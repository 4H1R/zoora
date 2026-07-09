package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Plan is a code-defined subscription key combining a tier and a member
// capacity: "<tier>_<size>", e.g. "pro_200". The catalog (PlanCatalog) is the
// single source of truth for what each plan includes; orgs store only the plan
// key + an expiry.
type Plan string

// PlanTier is the feature level of a plan, independent of member capacity.
type PlanTier string

const (
	TierFree PlanTier = "free"
	TierPlus PlanTier = "plus"
	TierPro  PlanTier = "pro"
	TierMax  PlanTier = "max"
)

// PlanTiers is ordered lowest to highest.
var PlanTiers = []PlanTier{TierFree, TierPlus, TierPro, TierMax}

// PlanSizes are the member capacities each tier is sold at, ascending.
var PlanSizes = []int64{50, 100, 200, 500, 1000}

// PlanFree is the fallback plan: new orgs start here and expired or unknown
// plans resolve to it.
const PlanFree Plan = "free_50"

// PlanKey builds the catalog key for a tier at a given member capacity.
func PlanKey(tier PlanTier, size int64) Plan {
	return Plan(fmt.Sprintf("%s_%d", tier, size))
}

// Tier extracts the tier part of the plan key ("" if malformed).
func (p Plan) Tier() PlanTier {
	tier, _ := p.parts()
	return tier
}

// Size extracts the member capacity of the plan key (0 if malformed).
func (p Plan) Size() int64 {
	_, size := p.parts()
	return size
}

func (p Plan) parts() (PlanTier, int64) {
	i := strings.LastIndexByte(string(p), '_')
	if i < 0 {
		return "", 0
	}
	size, err := strconv.ParseInt(string(p[i+1:]), 10, 64)
	if err != nil {
		return "", 0
	}
	return PlanTier(p[:i]), size
}

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
	FeatureWhiteboard        Feature = "whiteboard"
	FeatureChat              Feature = "chat"
	FeatureConnectors        Feature = "connectors"
	FeatureAutoGrading       Feature = "auto_grading"
	FeatureAI                Feature = "ai"
)

// AllFeatures keeps PlanInfo JSON complete: every flag appears explicitly per
// plan even when false.
var AllFeatures = []Feature{
	FeatureRecording, FeatureOfflineRooms, FeatureAdvancedAntiCheat,
	FeatureCustomRoles, FeatureSSO, FeatureWhiteboard, FeatureChat,
	FeatureConnectors, FeatureAutoGrading, FeatureAI,
}

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

// tierSpec defines a tier at the reference 50-member size. Numeric limits
// (except recording retention) scale linearly with size/50; features do not.
type tierSpec struct {
	features      map[Feature]bool
	participants  int64 // concurrent live participants
	storageGB     int64
	rooms         int64 // concurrent live rooms
	retentionDays int64 // recording retention, not scaled
}

// tierSpecs is the source of truth for what each tier includes. Editing a tier
// = editing this map; the full PlanCatalog is generated from it per size.
var tierSpecs = map[PlanTier]tierSpec{
	TierFree: {
		features: map[Feature]bool{
			FeatureOfflineRooms: true,
		},
		participants: 5, storageGB: 2, rooms: 1,
	},
	TierPlus: {
		features: map[Feature]bool{
			FeatureOfflineRooms: true,
		},
		participants: 10, storageGB: 10, rooms: 2,
	},
	TierPro: {
		features: map[Feature]bool{
			FeatureOfflineRooms: true, FeatureWhiteboard: true,
			FeatureAdvancedAntiCheat: true, FeatureConnectors: true,
			FeatureChat: true, FeatureAI: true, FeatureCustomRoles: true,
		},
		participants: 20, storageGB: 25, rooms: 5,
	},
	TierMax: {
		features: map[Feature]bool{
			FeatureOfflineRooms: true, FeatureWhiteboard: true,
			FeatureAdvancedAntiCheat: true, FeatureConnectors: true,
			FeatureChat: true, FeatureAI: true, FeatureCustomRoles: true,
			FeatureRecording: true, FeatureAutoGrading: true, FeatureSSO: true,
		},
		participants: 50, storageGB: 50, rooms: 5,
		retentionDays: 365,
	},
}

// PlanCatalog is the source of truth for plan entitlements, generated as
// tierSpecs × PlanSizes.
var PlanCatalog = buildPlanCatalog()

func buildPlanCatalog() map[Plan]Entitlements {
	out := make(map[Plan]Entitlements, len(PlanTiers)*len(PlanSizes))
	for _, tier := range PlanTiers {
		spec := tierSpecs[tier]
		for _, size := range PlanSizes {
			factor := size / 50
			features := make(map[Feature]bool, len(AllFeatures))
			for _, f := range AllFeatures {
				features[f] = spec.features[f]
			}
			p := PlanKey(tier, size)
			out[p] = Entitlements{
				Plan:     p,
				features: features,
				limits: map[Limit]int64{
					LimitMaxUsers:               size,
					LimitMaxParticipants:        spec.participants * factor,
					LimitStorageGB:              spec.storageGB * factor,
					LimitConcurrentRooms:        spec.rooms * factor,
					LimitRecordingRetentionDays: spec.retentionDays,
				},
			}
		}
	}
	return out
}

// PlanInfo is the public, serializable shape of a plan's entitlements, for the
// admin UI plan picker. The internal Entitlements maps are unexported, so this
// projects them into a stable JSON contract.
type PlanInfo struct {
	Plan     Plan             `json:"plan"`
	Tier     PlanTier         `json:"tier"`
	Size     int64            `json:"size"`
	Features map[Feature]bool `json:"features"`
	Limits   map[Limit]int64  `json:"limits"`
}

// Public projects a resolved Entitlements snapshot into the serializable
// PlanInfo contract (same shape as the catalog), for the
// /users/me/entitlements endpoint the SPA reads to gate plan-locked UI
// (nav visibility, paywalls) without duplicating the tier→feature rule.
func (e Entitlements) Public() PlanInfo {
	r := e.resolved()
	return PlanInfo{
		Plan:     r.Plan,
		Tier:     r.Plan.Tier(),
		Size:     r.Plan.Size(),
		Features: r.features,
		Limits:   r.limits,
	}
}

// PublicCatalog returns the plan catalog for API responses, grouped by size
// (ascending), tiers low to high within each size.
func PublicCatalog() []PlanInfo {
	out := make([]PlanInfo, 0, len(PlanSizes)*len(PlanTiers))
	for _, size := range PlanSizes {
		for _, tier := range PlanTiers {
			p := PlanKey(tier, size)
			ent := PlanCatalog[p]
			out = append(out, PlanInfo{
				Plan: p, Tier: tier, Size: size,
				Features: ent.features, Limits: ent.limits,
			})
		}
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
