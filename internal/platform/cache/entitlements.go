package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const orgPlanCacheTTL = 5 * time.Minute

func orgPlanKey(orgID uuid.UUID) string {
	return fmt.Sprintf("org:%s:plan", orgID.String())
}

// CachedOrgPlan is the raw (plan, expiry) pair cached per org. Plan is kept as a
// plain string so this platform package stays free of domain types; the caller
// converts to domain.Plan. Effectiveness (expiry downgrade) is resolved
// in-process against `now`, never cached.
type CachedOrgPlan struct {
	Plan      string     `json:"plan"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// GetOrgPlan returns the cached (plan, expiry) for an org. Miss or decode
// failure returns an error — callers fall back to the repository.
func GetOrgPlan(ctx context.Context, rdb *redis.Client, orgID uuid.UUID) (CachedOrgPlan, error) {
	data, err := rdb.Get(ctx, orgPlanKey(orgID)).Bytes()
	if err != nil {
		return CachedOrgPlan{}, err
	}
	var cp CachedOrgPlan
	if err := json.Unmarshal(data, &cp); err != nil {
		return CachedOrgPlan{}, fmt.Errorf("unmarshaling cached org plan: %w", err)
	}
	return cp, nil
}

// SetOrgPlan caches the (plan, expiry) for an org.
func SetOrgPlan(ctx context.Context, rdb *redis.Client, orgID uuid.UUID, plan string, expiresAt *time.Time) error {
	data, err := json.Marshal(CachedOrgPlan{Plan: plan, ExpiresAt: expiresAt})
	if err != nil {
		return fmt.Errorf("marshaling org plan: %w", err)
	}
	return rdb.Set(ctx, orgPlanKey(orgID), data, orgPlanCacheTTL).Err()
}

// InvalidateOrgPlan drops the cached plan for an org (call after a plan change).
func InvalidateOrgPlan(ctx context.Context, rdb *redis.Client, orgID uuid.UUID) error {
	return rdb.Del(ctx, orgPlanKey(orgID)).Err()
}
