package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/platform/cache"
)

// EntitlementsCacheBuster adapts the package-level cache.InvalidateOrgPlan into
// the service's entitlementsCacheBuster interface, so billing's interface stays
// free of the redis client type while wiring injects the concrete client.
type EntitlementsCacheBuster struct {
	rdb *redis.Client
}

func NewEntitlementsCacheBuster(rdb *redis.Client) *EntitlementsCacheBuster {
	return &EntitlementsCacheBuster{rdb: rdb}
}

// Invalidate drops the cached (plan, expiry) snapshot for an org after a plan
// change, mirroring how organizations' admin SetPlan busts the cache.
func (c *EntitlementsCacheBuster) Invalidate(ctx context.Context, orgID uuid.UUID) error {
	return cache.InvalidateOrgPlan(ctx, c.rdb, orgID)
}

var _ entitlementsCacheBuster = (*EntitlementsCacheBuster)(nil)
