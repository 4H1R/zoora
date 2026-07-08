package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const userCacheTTL = 5 * time.Minute

func userKey(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id.String())
}

// CachedUser is the per-request auth snapshot of a user, cached by ID to spare
// the users+roles lookup the auth middleware runs on every request. It holds
// only the fields the middleware reads; permissions and entitlements are cached
// separately. Kept free of domain types so this platform package stays a leaf.
type CachedUser struct {
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	RoleID         *uuid.UUID `json:"role_id,omitempty"`
	IsAdmin        bool       `json:"is_admin"`
	Username       string     `json:"username"`
	Name           string     `json:"name"`
	DisabledAt     *time.Time `json:"disabled_at,omitempty"`
}

// GetUser returns the cached auth snapshot for a user. Miss or decode failure
// returns an error — callers fall back to the repository.
func GetUser(ctx context.Context, rdb *redis.Client, id uuid.UUID) (CachedUser, error) {
	data, err := rdb.Get(ctx, userKey(id)).Bytes()
	if err != nil {
		return CachedUser{}, err
	}
	var cu CachedUser
	if err := json.Unmarshal(data, &cu); err != nil {
		return CachedUser{}, fmt.Errorf("unmarshaling cached user: %w", err)
	}
	return cu, nil
}

// SetUser caches the auth snapshot for a user.
func SetUser(ctx context.Context, rdb *redis.Client, id uuid.UUID, cu CachedUser) error {
	data, err := json.Marshal(cu)
	if err != nil {
		return fmt.Errorf("marshaling user: %w", err)
	}
	return rdb.Set(ctx, userKey(id), data, userCacheTTL).Err()
}

// InvalidateUser drops the cached snapshot for a user (call after any change to
// the user's row: profile, role, admin flag, disable/enable, delete).
func InvalidateUser(ctx context.Context, rdb *redis.Client, id uuid.UUID) error {
	return rdb.Del(ctx, userKey(id)).Err()
}
