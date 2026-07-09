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
	// markOnline is called when the socket registers and on every pong, to set
	// and refresh the user's online TTL.
	markOnline func(userID uuid.UUID)
	// onJoin is called after a successful room join so the rest of the room
	// learns this user is present.
	onJoin func(userID, convID uuid.UUID)
	// markOffline is called only when the user's LAST socket on this instance
	// disconnects; rooms are the conversations that socket had joined, so the
	// caller can publish a presence_update to each.
	markOffline func(userID uuid.UUID, rooms []uuid.UUID)
}

// serve registers the socket with the hub and runs its pumps until the
// connection closes. It blocks (running the read pump) so callers should
// invoke it from the goroutine handling the upgraded connection.
func (h *Hub) serve(ctx context.Context, ws *websocket.Conn, userID uuid.UUID, publishTyping func(convID, uid uuid.UUID), hooks presenceHooks) {
	c := &conn{userID: userID, send: make(chan outbound, 64), rooms: map[uuid.UUID]bool{}}
	h.addSocket(c)
	if hooks.markOnline != nil {
		hooks.markOnline(userID)
	}

	go h.writePump(ws, c)
	h.readPump(ctx, ws, c, publishTyping, hooks)
}

func (h *Hub) readPump(ctx context.Context, ws *websocket.Conn, c *conn, publishTyping func(convID, uid uuid.UUID), hooks presenceHooks) {
	defer func() {
		rooms, lastSocket := h.removeSocket(c)
		if lastSocket && hooks.markOffline != nil {
			hooks.markOffline(c.userID, rooms)
		}
		_ = ws.Close()
	}()
	ws.SetReadLimit(maxMessageSize)
	_ = ws.SetReadDeadline(time.Now().Add(pingPeriod + pongWait))
	ws.SetPongHandler(func(string) error {
		if hooks.markOnline != nil {
			hooks.markOnline(c.userID) // refresh online TTL on heartbeat
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
