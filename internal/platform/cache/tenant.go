package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
)

const tenantTTL = 5 * time.Minute

func tenantKey(slug string) string { return "tenant:slug:" + slug }

type cachedTenant struct {
	OrgID  uuid.UUID                 `json:"org_id"`
	Status domain.OrganizationStatus `json:"status"`
}

// GetTenant returns the cached (org_id, status) for a slug, or redis.Nil miss.
func GetTenant(ctx context.Context, rdb *redis.Client, slug string) (uuid.UUID, domain.OrganizationStatus, error) {
	val, err := rdb.Get(ctx, tenantKey(slug)).Bytes()
	if err != nil {
		return uuid.Nil, "", err
	}
	var c cachedTenant
	if err := json.Unmarshal(val, &c); err != nil {
		return uuid.Nil, "", err
	}
	return c.OrgID, c.Status, nil
}

func SetTenant(ctx context.Context, rdb *redis.Client, slug string, orgID uuid.UUID, status domain.OrganizationStatus) error {
	b, err := json.Marshal(cachedTenant{OrgID: orgID, Status: status})
	if err != nil {
		return fmt.Errorf("cache.SetTenant: %w", err)
	}
	return rdb.Set(ctx, tenantKey(slug), b, tenantTTL).Err()
}

// BustTenant removes a slug's cache entry after a slug or status change.
func BustTenant(ctx context.Context, rdb *redis.Client, slug string) error {
	return rdb.Del(ctx, tenantKey(slug)).Err()
}
