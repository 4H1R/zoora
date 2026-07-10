package chathub

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 8192
)

// PresenceTTL is how long a user is considered online after their last
// heartbeat: one full ping interval plus the pong grace window, so a user
// whose socket missed a single ping is not prematurely marked offline. The WS
// handler refreshes it on connect and on every pong; main uses it to size the
// Presence tracker's Redis key TTL.
const PresenceTTL = pingPeriod + pongWait

type clientMsg struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id"`
}

// presenceHooks carries the presence side effects the socket lifecycle fires.
// They are wired in the WS handler (where both the Presence tracker and the
// Redis Bridge are available) so chathub never imports the conversations
// package. Any hook may be nil (e.g. in tests), in which case it is skipped.
type presenceHooks struct {
	// onConnect is called once when the socket registers, to count it toward the
	// user's cross-instance live-socket total (0->1 means they came online).
	onConnect func(userID uuid.UUID)
	// onHeartbeat is called on every pong to refresh the user's online TTL
	// without changing the socket count.
	onHeartbeat func(userID uuid.UUID)
	// onJoin is called after a successful room join so the rest of the room
	// learns this user is present.
	onJoin func(userID, convID uuid.UUID)
	// onDisconnect is called for EVERY socket close (not just the instance's
	// last) so the socket is decremented from the cross-instance count; rooms
	// are the conversations that socket had joined, for offline fan-out when the
	// count reaches zero.
	onDisconnect func(userID uuid.UUID, rooms []uuid.UUID)
}

// serve registers the socket with the hub and runs its pumps until the
// connection closes. It blocks (running the read pump) so callers should
// invoke it from the goroutine handling the upgraded connection.
func (h *Hub) serve(ctx context.Context, ws *websocket.Conn, userID uuid.UUID, publishTyping func(convID, uid uuid.UUID), hooks presenceHooks) {
	c := &conn{userID: userID, send: make(chan outbound, 64), rooms: map[uuid.UUID]bool{}}
	h.addSocket(c)
	if hooks.onConnect != nil {
		hooks.onConnect(userID)
	}

	go h.writePump(ws, c)
	h.readPump(ctx, ws, c, publishTyping, hooks)
}

func (h *Hub) readPump(ctx context.Context, ws *websocket.Conn, c *conn, publishTyping func(convID, uid uuid.UUID), hooks presenceHooks) {
	defer func() {
		// Every socket must decrement the cross-instance count, so fire
		// onDisconnect regardless of whether this was the instance's last socket
		// for the user; the presence layer decides when the user is truly offline.
		rooms, _ := h.removeSocket(c)
		if hooks.onDisconnect != nil {
			hooks.onDisconnect(c.userID, rooms)
		}
		_ = ws.Close()
	}()
	ws.SetReadLimit(maxMessageSize)
	_ = ws.SetReadDeadline(time.Now().Add(pingPeriod + pongWait))
	ws.SetPongHandler(func(string) error {
		if hooks.onHeartbeat != nil {
			hooks.onHeartbeat(c.userID) // refresh online TTL on heartbeat
		}
		return ws.SetReadDeadline(time.Now().Add(pingPeriod + pongWait))
	})
	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			return
		}
		var m clientMsg
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		convID, perr := uuid.Parse(m.ConversationID)
		if perr != nil {
			continue
		}
		switch m.Type {
		case "join":
			ok, mErr := h.members.IsMember(ctx, convID, c.userID)
			if mErr != nil || !ok {
				continue // silently ignore unauthorized joins
			}
			h.joinRoom(c, convID)
			if hooks.onJoin != nil {
				hooks.onJoin(c.userID, convID) // tell the room this user is present
			}
		case "leave":
			h.leaveRoom(c, convID)
		case "typing":
			if c.rooms[convID] {
				publishTyping(convID, c.userID) // fanned out via redis to the room
			}
		}
	}
}

func (h *Hub) writePump(ws *websocket.Conn, c *conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = ws.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, msg.data); err != nil {
				return
			}
		case <-ticker.C:
			_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
