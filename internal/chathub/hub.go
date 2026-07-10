// Package chathub implements the in-process WebSocket hub for realtime chat
// fan-out: per-instance registries of connected sockets, room membership, and
// non-blocking delivery to local connections. The Redis bridge (P2b) wires
// onFirstJoin/onLastLeave to keep cross-instance pub/sub subscriptions in
// sync with local room membership.
package chathub

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// membership is the narrow read port the hub uses to authorize joins.
type membership interface {
	IsMember(ctx context.Context, convID, userID uuid.UUID) (bool, error)
	ListUserIDs(ctx context.Context, convID uuid.UUID) ([]uuid.UUID, error)
}

type outbound struct {
	data []byte
}

type conn struct {
	userID uuid.UUID
	send   chan outbound
	// rooms is owned by this conn's readPump goroutine: written under h.mu in
	// joinRoom/leaveRoom/removeSocket (all called only from that goroutine) and
	// read there unlocked (the typing case). Do not touch it from elsewhere.
	rooms map[uuid.UUID]bool // conversations this socket joined
}

type Hub struct {
	mu          sync.RWMutex
	userSockets map[uuid.UUID]map[*conn]bool
	rooms       map[uuid.UUID]map[*conn]bool
	members     membership
	logger      *slog.Logger

	// onFirstJoin/onLastLeave notify the redis bridge to (un)subscribe the
	// conversation channel on the 0<->1 room-membership transition;
	// onUserFirstSocket/onUserLastSocket do the same for the user channel on the
	// 0<->1 socket transition. All four are invoked while h.mu is held, so they
	// MUST NOT call back into the Hub (RWMutex is not reentrant — doing so would
	// deadlock) or perform blocking I/O; the bridge only enqueues.
	onFirstJoin       func(convID uuid.UUID)
	onLastLeave       func(convID uuid.UUID)
	onUserFirstSocket func(userID uuid.UUID)
	onUserLastSocket  func(userID uuid.UUID)
}

func NewHub(members membership, logger *slog.Logger) *Hub {
	return &Hub{
		userSockets: make(map[uuid.UUID]map[*conn]bool),
		rooms:       make(map[uuid.UUID]map[*conn]bool),
		members:     members,
		logger:      logger,
	}
}

func (h *Hub) addSocket(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.userSockets[c.userID] == nil {
		h.userSockets[c.userID] = make(map[*conn]bool)
		if h.onUserFirstSocket != nil {
			h.onUserFirstSocket(c.userID) // first socket for this user here: subscribe user channel
		}
	}
	h.userSockets[c.userID][c] = true
}

// removeSocket detaches c from every room it joined and from its user's socket
// set, returning the conversations it had joined (for presence fan-out) and
// whether this was the user's LAST socket on this instance (so the caller can
// mark the user offline only once, supporting multi-device). Both are computed
// under a single lock so they cannot race a concurrent join/leave on c.
func (h *Hub) removeSocket(c *conn) (rooms []uuid.UUID, lastSocket bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rooms = make([]uuid.UUID, 0, len(c.rooms))
	for convID := range c.rooms {
		rooms = append(rooms, convID)
		if set := h.rooms[convID]; set != nil {
			delete(set, c)
			if len(set) == 0 {
				delete(h.rooms, convID)
				if h.onLastLeave != nil {
					h.onLastLeave(convID)
				}
			}
		}
	}
	if set := h.userSockets[c.userID]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.userSockets, c.userID)
			lastSocket = true
			if h.onUserLastSocket != nil {
				h.onUserLastSocket(c.userID) // last socket for this user here: unsubscribe user channel
			}
		}
	}
	close(c.send)
	return rooms, lastSocket
}

func (h *Hub) joinRoom(c *conn, convID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[convID] == nil {
		h.rooms[convID] = make(map[*conn]bool)
		if h.onFirstJoin != nil {
			h.onFirstJoin(convID)
		}
	}
	h.rooms[convID][c] = true
	c.rooms[convID] = true
}

func (h *Hub) leaveRoom(c *conn, convID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set := h.rooms[convID]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.rooms, convID)
			if h.onLastLeave != nil {
				h.onLastLeave(convID)
			}
		}
	}
	delete(c.rooms, convID)
}

// deliverToRoom sends to every local socket that joined convID.
func (h *Hub) deliverToRoom(convID uuid.UUID, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[convID] {
		select {
		case c.send <- outbound{data: data}:
		default: // slow consumer: drop (client re-syncs via poll)
		}
	}
}

// deliverToUser sends to every local socket of userID (all devices).
func (h *Hub) deliverToUser(userID uuid.UUID, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.userSockets[userID] {
		select {
		case c.send <- outbound{data: data}:
		default:
		}
	}
}
