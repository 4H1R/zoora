package chathub

import (
	"context"
	"errors"
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

// Presence tracks per-user online state in Redis as a cross-instance refcount
// of live sockets. presence:online:<uid> holds the number of sockets currently
// connected for the user across ALL app instances; the user is online while it
// is > 0. A TTL is (re)armed on connect and every heartbeat so a crashed
// instance's sockets cannot pin a user online forever — the count self-heals
// when the key expires. presence:seen:<uid> is a durable last-seen timestamp,
// updated when a user comes online and when their last socket goes.
//
// This is what makes presence correct across multiple app instances: earlier
// each instance deleted the online key when ITS last socket for the user
// dropped, marking a user offline even while they were still connected on
// another instance. Counting sockets instead of "last socket on this instance"
// removes that false-offline.
type Presence struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewPresence(rdb *redis.Client, ttl time.Duration) *Presence {
	return &Presence{rdb: rdb, ttl: ttl}
}

func onlineKey(uid uuid.UUID) string { return presenceOnlinePrefix + uid.String() }
func seenKey(uid uuid.UUID) string   { return presenceSeenPrefix + uid.String() }

// connectScript increments the socket count, (re)arms the TTL, and stamps
// last-seen — atomically, so a crash between INCR and EXPIRE can't leave a
// TTL-less key that pins the user online forever. Returns the new count so the
// caller can detect the 0->1 online transition.
var connectScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
redis.call('EXPIRE', KEYS[1], ARGV[1])
redis.call('SET', KEYS[2], ARGV[2])
return n
`)

// disconnectScript decrements the socket count. When it reaches zero (or below,
// e.g. after a partial crash-recovery left the counter stale) it deletes the
// key and stamps last-seen, returning 0 to signal the user is now fully
// offline. Otherwise it refreshes the TTL for the still-connected sockets and
// returns the remaining count.
var disconnectScript = redis.NewScript(`
local n = redis.call('DECR', KEYS[1])
if n <= 0 then
  redis.call('DEL', KEYS[1])
  redis.call('SET', KEYS[2], ARGV[1])
  return 0
end
redis.call('EXPIRE', KEYS[1], ARGV[2])
return n
`)

func (p *Presence) ttlSeconds() int { return int(p.ttl.Seconds()) }

// Connect registers a newly-opened socket for uid and refreshes the online TTL.
// It returns true when this was the user's first live socket anywhere (0->1),
// i.e. they just came online.
func (p *Presence) Connect(ctx context.Context, uid uuid.UUID) (online bool, err error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	n, err := connectScript.Run(ctx, p.rdb,
		[]string{onlineKey(uid), seenKey(uid)}, p.ttlSeconds(), now).Int64()
	if err != nil {
		return false, fmt.Errorf("presence.Connect: %w", err)
	}
	return n == 1, nil
}

// Refresh extends the online TTL on a heartbeat without changing the socket
// count. A missing key (already expired) is left absent: Refresh never
// resurrects a user the count logic considers offline.
func (p *Presence) Refresh(ctx context.Context, uid uuid.UUID) error {
	if err := p.rdb.Expire(ctx, onlineKey(uid), p.ttl).Err(); err != nil {
		return fmt.Errorf("presence.Refresh: %w", err)
	}
	return nil
}

// Disconnect deregisters a closed socket for uid. It returns true when the
// user's last socket across all instances has gone — i.e. they are now offline,
// at which point last-seen is stamped. While other sockets remain it refreshes
// the TTL and returns false.
func (p *Presence) Disconnect(ctx context.Context, uid uuid.UUID) (offline bool, err error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	n, err := disconnectScript.Run(ctx, p.rdb,
		[]string{onlineKey(uid), seenKey(uid)}, now, p.ttlSeconds()).Int64()
	if err != nil {
		return false, fmt.Errorf("presence.Disconnect: %w", err)
	}
	return n == 0, nil
}

// Get returns the presence Status for each requested user id. Users with no
// recorded last-seen timestamp get a zero-value LastSeen. A present online key
// means at least one live socket (the counter is deleted at zero), so its mere
// existence is the online signal.
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
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
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
