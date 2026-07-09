package chathub

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func newTestPresence(t *testing.T, ttl time.Duration) (*Presence, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewPresence(rdb, ttl), server
}

func TestPresence_Connect_MarksOnlineWithTTL(t *testing.T) {
	ctx := context.Background()
	p, server := newTestPresence(t, 5*time.Second)

	userA := uuid.New()
	userB := uuid.New()

	online, err := p.Connect(ctx, userA)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if !online {
		t.Fatal("first Connect should report online (0->1)")
	}

	statuses, err := p.Get(ctx, []uuid.UUID{userA, userB})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !statuses[userA].Online {
		t.Fatal("userA Online = false, want true")
	}
	if statuses[userB].Online {
		t.Fatal("userB Online = true, want false")
	}

	server.FastForward(6 * time.Second)

	statuses, err = p.Get(ctx, []uuid.UUID{userA})
	if err != nil {
		t.Fatalf("Get() after TTL error = %v", err)
	}
	if statuses[userA].Online {
		t.Fatal("userA Online = true after TTL expiry, want false")
	}
	if statuses[userA].LastSeen.IsZero() {
		t.Fatal("userA LastSeen is zero after TTL expiry, want non-zero (seen key survives)")
	}
}

// TestPresence_MultiSocket_StaysOnlineUntilLastDisconnect is the core
// multi-instance fix: a user with two live sockets stays online after one
// disconnects, and only goes offline when the last socket goes. Two Connects
// stand in for the same user connected to two app instances sharing Redis.
func TestPresence_MultiSocket_StaysOnlineUntilLastDisconnect(t *testing.T) {
	ctx := context.Background()
	p, _ := newTestPresence(t, 30*time.Second)
	user := uuid.New()

	if online, _ := p.Connect(ctx, user); !online {
		t.Fatal("first Connect should report online")
	}
	if online, _ := p.Connect(ctx, user); online {
		t.Fatal("second Connect must NOT report a fresh online transition")
	}

	// First socket drops: user still online on the second.
	offline, err := p.Disconnect(ctx, user)
	if err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if offline {
		t.Fatal("Disconnect with a socket remaining must not report offline")
	}
	statuses, _ := p.Get(ctx, []uuid.UUID{user})
	if !statuses[user].Online {
		t.Fatal("user should still be Online while one socket remains")
	}

	// Last socket drops: now offline.
	offline, err = p.Disconnect(ctx, user)
	if err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if !offline {
		t.Fatal("Disconnect of the last socket must report offline")
	}
	statuses, _ = p.Get(ctx, []uuid.UUID{user})
	if statuses[user].Online {
		t.Fatal("user should be offline after last socket disconnect")
	}
	if statuses[user].LastSeen.IsZero() {
		t.Fatal("LastSeen should be stamped on going offline")
	}
}

func TestPresence_Disconnect_BelowZeroSelfHeals(t *testing.T) {
	ctx := context.Background()
	p, _ := newTestPresence(t, 30*time.Second)
	user := uuid.New()

	// Disconnect with no prior Connect (e.g. crash-recovery skew): must report
	// offline and not leave a negative counter pinning odd state.
	offline, err := p.Disconnect(ctx, user)
	if err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if !offline {
		t.Fatal("Disconnect below zero should report offline")
	}
	statuses, _ := p.Get(ctx, []uuid.UUID{user})
	if statuses[user].Online {
		t.Fatal("user should be offline after self-healing disconnect")
	}
}

func TestPresence_Refresh_ExtendsTTL(t *testing.T) {
	ctx := context.Background()
	p, server := newTestPresence(t, 10*time.Second)
	user := uuid.New()

	if _, err := p.Connect(ctx, user); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	server.FastForward(8 * time.Second) // within TTL
	if err := p.Refresh(ctx, user); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	server.FastForward(8 * time.Second) // would have expired without the refresh

	statuses, _ := p.Get(ctx, []uuid.UUID{user})
	if !statuses[user].Online {
		t.Fatal("user should still be online after a heartbeat refresh")
	}
}
