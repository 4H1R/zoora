package chathub

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestPresence_MarkOnline_TTL(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	ttl := 5 * time.Second
	p := NewPresence(rdb, ttl)

	userA := uuid.New()
	userB := uuid.New()

	if err := p.MarkOnline(ctx, userA); err != nil {
		t.Fatalf("MarkOnline() error = %v", err)
	}

	statuses, err := p.Get(ctx, []uuid.UUID{userA, userB})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !statuses[userA].Online {
		t.Fatalf("userA Online = false, want true")
	}
	if statuses[userB].Online {
		t.Fatalf("userB Online = true, want false")
	}

	server.FastForward(ttl + time.Second)

	statuses, err = p.Get(ctx, []uuid.UUID{userA, userB})
	if err != nil {
		t.Fatalf("Get() after TTL error = %v", err)
	}
	if statuses[userA].Online {
		t.Fatalf("userA Online = true after TTL expiry, want false")
	}
	if statuses[userA].LastSeen.IsZero() {
		t.Fatalf("userA LastSeen is zero after TTL expiry, want non-zero (seen key should survive)")
	}
}

func TestPresence_MarkOffline(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	p := NewPresence(rdb, 30*time.Second)
	userA := uuid.New()

	if err := p.MarkOnline(ctx, userA); err != nil {
		t.Fatalf("MarkOnline() error = %v", err)
	}

	statuses, err := p.Get(ctx, []uuid.UUID{userA})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	firstSeen := statuses[userA].LastSeen
	if firstSeen.IsZero() {
		t.Fatalf("LastSeen is zero after MarkOnline, want non-zero")
	}

	server.FastForward(1 * time.Second)

	if err := p.MarkOffline(ctx, userA); err != nil {
		t.Fatalf("MarkOffline() error = %v", err)
	}

	statuses, err = p.Get(ctx, []uuid.UUID{userA})
	if err != nil {
		t.Fatalf("Get() after MarkOffline error = %v", err)
	}
	if statuses[userA].Online {
		t.Fatalf("Online = true after MarkOffline, want false")
	}
	if !statuses[userA].LastSeen.After(firstSeen) {
		t.Fatalf("LastSeen = %v, want after %v (advanced on MarkOffline)", statuses[userA].LastSeen, firstSeen)
	}
}
