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

// subCmd is a dynamic (un)subscribe request handed from the hub's membership
// hooks to the Run loop, which owns the PubSub connection.
type subCmd struct {
	channel   string
	subscribe bool
}

// Bridge publishes broadcasts to Redis and forwards inbound Redis messages to
// the local hub. It implements the conversations.broadcaster port.
//
// Fan-out is scoped: instead of one wildcard PSUBSCRIBE over chat:conversation:*
// / chat:user:* (which made every app instance receive every message in the
// whole system and filter locally), the Bridge SUBSCRIBEs to a conversation
// channel only while this instance holds a socket in that conversation, and to
// a user channel only while this instance holds a socket for that user. The hub
// fires the first-join/last-leave and first-socket/last-socket transitions;
// the Bridge turns them into exact (un)subscribes.
type Bridge struct {
	hub    *Hub
	rdb    *redis.Client
	logger *slog.Logger

	// pending carries (un)subscribe requests from the hub hooks (which fire
	// under the hub lock and so must not perform Redis I/O) to the Run loop. It
	// is an UNBOUNDED, order-preserving handoff guarded by cmdMu, with wake as a
	// coalesced wakeup signal. Unbounded + non-blocking is what breaks the
	// hub-lock/Run-loop deadlock: a hook firing under the hub write lock appends
	// and returns without ever parking, so the write lock is never held while
	// blocked, so Run's deliverToRoom/deliverToUser (which take the hub RLock)
	// can always proceed and keep draining. Ordering is preserved because the
	// hub write lock serializes the enqueue callers, so appends land in
	// transition order and are never dropped or reordered.
	cmdMu   sync.Mutex
	pending []subCmd
	wake    chan struct{} // buffered(1); a non-blocking nudge to drain pending

	mu     sync.Mutex
	pubsub *redis.PubSub // guarded for the ctx-cancel watcher; nil until Run starts
}

func NewBridge(hub *Hub, rdb *redis.Client, logger *slog.Logger) *Bridge {
	b := &Bridge{hub: hub, rdb: rdb, logger: logger, wake: make(chan struct{}, 1)}
	// Dynamic (un)subscribe hooks fired by the hub on the local membership
	// transitions that change what this instance needs to receive.
	hub.onFirstJoin = func(convID uuid.UUID) { b.enqueue(convChannelPrefix+convID.String(), true) }
	hub.onLastLeave = func(convID uuid.UUID) { b.enqueue(convChannelPrefix+convID.String(), false) }
	hub.onUserFirstSocket = func(userID uuid.UUID) { b.enqueue(userChannelPrefix+userID.String(), true) }
	hub.onUserLastSocket = func(userID uuid.UUID) { b.enqueue(userChannelPrefix+userID.String(), false) }
	return b
}

// enqueue hands a (un)subscribe request to the Run loop. It is called under the
// hub write lock, so it MUST NOT block: it appends to the unbounded pending
// slice and fires a coalesced wakeup, then returns immediately. Never blocking
// under the hub lock is precisely what prevents the deadlock where a full
// buffer would park the lock holder while Run — blocked on the hub RLock in
// deliverTo* — can no longer drain. Appends happen in transition order (the hub
// lock serializes callers) so subscribe/unsubscribe commands are neither
// dropped nor reordered, keeping the refcount correct.
func (b *Bridge) enqueue(channel string, subscribe bool) {
	b.cmdMu.Lock()
	b.pending = append(b.pending, subCmd{channel: channel, subscribe: subscribe})
	b.cmdMu.Unlock()
	select {
	case b.wake <- struct{}{}:
	default: // a wakeup is already pending; Run will drain all of pending
	}
}

// takePending atomically swaps out the queued commands for the Run loop to
// apply. Returns them in enqueue (transition) order.
func (b *Bridge) takePending() []subCmd {
	b.cmdMu.Lock()
	cmds := b.pending
	b.pending = nil
	b.cmdMu.Unlock()
	return cmds
}

// Run owns the PubSub connection and blocks until ctx is done. It applies
// dynamic (un)subscribes from the command queue and dispatches inbound messages
// to the local hub.
//
// go-redis's PubSub.Channel() reader does NOT observe the ctx passed to
// Subscribe; its output channel only unblocks on Close(). So a watcher closes
// the active PubSub when ctx is canceled, which unblocks the range below and
// lets Run return instead of leaking its goroutine on shutdown. If the channel
// closes for any other reason (connection died and internal reconnect was
// exhausted), Run rebuilds the PubSub and replays the still-active
// subscriptions so realtime delivery survives a transient Redis blip.
func (b *Bridge) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		if b.pubsub != nil {
			_ = b.pubsub.Close()
		}
		b.mu.Unlock()
	}()

	// refs is the per-channel local subscription refcount, owned solely by this
	// goroutine (mutated only via applyCmd below), so it needs no lock. It also
	// drives replay of active subscriptions after a reconnect.
	refs := map[string]int{}

	for {
		if ctx.Err() != nil {
			return
		}
		ps := b.rdb.Subscribe(ctx)
		b.mu.Lock()
		b.pubsub = ps
		b.mu.Unlock()
		// Cover the race where ctx was canceled between the guard above and the
		// assignment: the watcher may have already fired against the old pubsub.
		if ctx.Err() != nil {
			_ = ps.Close()
			return
		}
		// Replay current subscriptions after a (re)connect so delivery resumes
		// for every conversation/user this instance still holds.
		if len(refs) > 0 {
			channels := make([]string, 0, len(refs))
			for ch := range refs {
				channels = append(channels, ch)
			}
			if err := ps.Subscribe(ctx, channels...); err != nil {
				b.logger.Warn("chathub resubscribe on reconnect failed", "error", err)
			}
		}

		ch := ps.Channel()
	drain:
		for {
			// Prioritize pending (un)subscribes: apply the full queued batch
			// before handling any inbound message so subscription/refcount state
			// never lags behind local membership while messages are flowing, and
			// so a backed-up queue can always be drained (the deadlock guard).
			for _, cmd := range b.takePending() {
				b.applyCmd(ctx, ps, refs, cmd)
			}
			select {
			case <-ctx.Done():
				_ = ps.Close()
				return
			case <-b.wake:
				// New commands queued: loop to drain pending at the top.
			case msg, ok := <-ch:
				if !ok {
					break drain
				}
				b.dispatch(msg.Channel, []byte(msg.Payload))
			}
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

// applyCmd applies one (un)subscribe against the live PubSub, using the local
// refcount so a channel is really SUBSCRIBEd only on the first joiner and
// UNSUBSCRIBEd only on the last leaver. The hub already fires clean
// first/last transitions, so refs normally toggles 0<->1; the counter is a
// safety net against hook races and reconnect replay.
func (b *Bridge) applyCmd(ctx context.Context, ps *redis.PubSub, refs map[string]int, cmd subCmd) {
	if cmd.subscribe {
		refs[cmd.channel]++
		if refs[cmd.channel] == 1 {
			if err := ps.Subscribe(ctx, cmd.channel); err != nil {
				b.logger.Warn("chathub subscribe failed", "channel", cmd.channel, "error", err)
			}
		}
		return
	}
	refs[cmd.channel]--
	if refs[cmd.channel] <= 0 {
		delete(refs, cmd.channel)
		if err := ps.Unsubscribe(ctx, cmd.channel); err != nil {
			b.logger.Warn("chathub unsubscribe failed", "channel", cmd.channel, "error", err)
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
