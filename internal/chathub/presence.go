package chathub

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	presenceOnlinePrefix = "presence:online:"
	presenceSeenPrefix   = "presence:seen:"
)

// Status reports whether a user is currently online and, regardless of
// current online state, when they were last seen.
type Status struct {
	Online   bool      `json:"online"`
	LastSeen time.Time `json:"last_seen"`
}

// Presence tracks per-user online/offline state in Redis: a volatile key
// with a TTL (existence = online, refreshed on every heartbeat) and a
// durable last-seen timestamp updated whenever the user goes online or
// offline.
type Presence struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewPresence(rdb *redis.Client, ttl time.Duration) *Presence {
	return &Presence{rdb: rdb, ttl: ttl}
}

func onlineKey(uid uuid.UUID) string { return presenceOnlinePrefix + uid.String() }
func seenKey(uid uuid.UUID) string   { return presenceSeenPrefix + uid.String() }

// MarkOnline marks uid online for the configured TTL and records the
// current time as last-seen.
func (p *Presence) MarkOnline(ctx context.Context, uid uuid.UUID) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pipe := p.rdb.Pipeline()
	pipe.Set(ctx, onlineKey(uid), "1", p.ttl)
	pipe.Set(ctx, seenKey(uid), now, 0)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("presence.MarkOnline: %w", err)
	}
	return nil
}

// MarkOffline clears uid's online state and records the current time as
// last-seen.
func (p *Presence) MarkOffline(ctx context.Context, uid uuid.UUID) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pipe := p.rdb.Pipeline()
	pipe.Del(ctx, onlineKey(uid))
	pipe.Set(ctx, seenKey(uid), now, 0)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("presence.MarkOffline: %w", err)
	}
	return nil
}

// Get returns the presence Status for each requested user id. Users with no
// recorded last-seen timestamp get a zero-value LastSeen.
func (p *Presence) Get(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]Status, error) {
	result := make(map[uuid.UUID]Status, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	pipe := p.rdb.Pipeline()
	existsCmds := make(map[uuid.UUID]*redis.IntCmd, len(ids))
	seenCmds := make(map[uuid.UUID]*redis.StringCmd, len(ids))
	for _, id := range ids {
		existsCmds[id] = pipe.Exists(ctx, onlineKey(id))
		seenCmds[id] = pipe.Get(ctx, seenKey(id))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("presence.Get: %w", err)
	}

	for _, id := range ids {
		status := Status{Online: existsCmds[id].Val() > 0}
		if raw, err := seenCmds[id].Result(); err == nil {
			if ts, err := time.Parse(time.RFC3339, raw); err == nil {
				status.LastSeen = ts.UTC()
			}
		}
		result[id] = status
	}
	return result, nil
}
