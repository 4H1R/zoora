package chathub

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	convChannelPrefix = "chat:conversation:"
	userChannelPrefix = "chat:user:"

	// pubsubRestartDelay throttles resubscribe attempts after the Redis
	// PubSub channel closes unexpectedly (e.g. connection drop that go-redis
	// gave up reconnecting), so a persistently-down Redis doesn't spin.
	pubsubRestartDelay = 1 * time.Second
)

type envelope struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// Bridge publishes broadcasts to Redis and forwards inbound Redis messages to
// the local hub. It implements the conversations.broadcaster port.
type Bridge struct {
	hub    *Hub
	rdb    *redis.Client
	logger *slog.Logger

	mu     sync.Mutex
	subs   map[uuid.UUID]int // local join refcount per conversation
	pubsub *redis.PubSub
}

func NewBridge(hub *Hub, rdb *redis.Client, logger *slog.Logger) *Bridge {
	b := &Bridge{hub: hub, rdb: rdb, logger: logger, subs: map[uuid.UUID]int{}}
	// dynamic (un)subscribe hooks fired by the hub on first-join / last-leave.
	hub.onFirstJoin = b.subscribeConversation
	hub.onLastLeave = b.unsubscribeConversation
	return b
}

// Run starts the subscriber loop and blocks until ctx is done. The user
// channels are subscribed eagerly for every conversation this instance cares
// about; simpler: subscribe to a per-instance pattern. Here we PSUBSCRIBE the
// conversation + user spaces.
//
// go-redis's PubSub.Channel() closes its output channel if the underlying
// connection dies and internal reconnect attempts are exhausted, which would
// otherwise silently kill realtime delivery for the life of the process. To
// survive a transient Redis blip, treat channel closure as a signal to
// re-PSUBSCRIBE and keep going until ctx is canceled.
func (b *Bridge) Run(ctx context.Context) {
	// go-redis's Channel() reader does NOT observe the ctx passed to PSubscribe;
	// the output channel only unblocks on Close(). So close the active PubSub
	// when ctx is canceled, which unblocks the range loop below and lets Run
	// return instead of leaking its goroutine on shutdown.
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		if b.pubsub != nil {
			_ = b.pubsub.Close()
		}
		b.mu.Unlock()
	}()

	for {
		if ctx.Err() != nil {
			return
		}
		ps := b.rdb.PSubscribe(ctx, convChannelPrefix+"*", userChannelPrefix+"*")
		b.mu.Lock()
		b.pubsub = ps
		b.mu.Unlock()
		// Cover the race where ctx was canceled between the guard above and the
		// assignment: the watcher may have already fired against the old pubsub.
		if ctx.Err() != nil {
			_ = ps.Close()
			return
		}

		for msg := range ps.Channel() {
			b.dispatch(msg.Channel, []byte(msg.Payload))
		}
		_ = ps.Close()

		if ctx.Err() != nil {
			return
		}
		b.logger.Warn("chathub redis pubsub loop exited unexpectedly, resubscribing")
		select {
		case <-ctx.Done():
			return
		case <-time.After(pubsubRestartDelay):
		}
	}
}

func (b *Bridge) dispatch(channel string, payload []byte) {
	switch {
	case strings.HasPrefix(channel, convChannelPrefix):
		if id, err := uuid.Parse(strings.TrimPrefix(channel, convChannelPrefix)); err == nil {
			b.hub.deliverToRoom(id, payload)
		} else {
			b.logger.Warn("chathub.dispatch bad conversation channel", "channel", channel)
		}
	case strings.HasPrefix(channel, userChannelPrefix):
		if id, err := uuid.Parse(strings.TrimPrefix(channel, userChannelPrefix)); err == nil {
			b.hub.deliverToUser(id, payload)
		} else {
			b.logger.Warn("chathub.dispatch bad user channel", "channel", channel)
		}
	}
}

// With PSUBSCRIBE covering the whole space, per-conversation (un)subscribe is a
// no-op refcount (kept for future switch to exact SUBSCRIBE). Left as counters.
func (b *Bridge) subscribeConversation(convID uuid.UUID) {
	b.mu.Lock()
	b.subs[convID]++
	b.mu.Unlock()
}
func (b *Bridge) unsubscribeConversation(convID uuid.UUID) {
	b.mu.Lock()
	if b.subs[convID] <= 1 {
		delete(b.subs, convID) // drop at zero so the map doesn't grow unbounded
	} else {
		b.subs[convID]--
	}
	b.mu.Unlock()
}

func (b *Bridge) publish(ctx context.Context, channel string, eventType string, data any) {
	payload, err := json.Marshal(envelope{Type: eventType, Data: data})
	if err != nil {
		b.logger.Error("chathub.publish marshal", "event", eventType, "error", err)
		return
	}
	if err := b.rdb.Publish(ctx, channel, payload).Err(); err != nil {
		b.logger.Error("chathub.publish", "channel", channel, "error", err)
	}
}

// ---- conversations.broadcaster implementation ----

func (b *Bridge) ToConversation(ctx context.Context, convID uuid.UUID, eventType string, data any) {
	b.publish(ctx, convChannelPrefix+convID.String(), eventType, data)
}

func (b *Bridge) ToUser(ctx context.Context, userID uuid.UUID, eventType string, data any) {
	b.publish(ctx, userChannelPrefix+userID.String(), eventType, data)
}

// PublishTyping is passed into the connection read-pump.
func (b *Bridge) PublishTyping(convID, userID uuid.UUID) {
	b.publish(context.Background(), convChannelPrefix+convID.String(), "user_typing",
		map[string]any{"conversation_id": convID.String(), "user_id": userID.String()})
}
