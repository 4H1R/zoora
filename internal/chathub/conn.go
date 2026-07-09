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

type clientMsg struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id"`
}

// serve registers the socket with the hub and runs its pumps until the
// connection closes. It blocks (running the read pump) so callers should
// invoke it from the goroutine handling the upgraded connection.
func (h *Hub) serve(ctx context.Context, ws *websocket.Conn, userID uuid.UUID, publishTyping func(convID, uid uuid.UUID)) {
	c := &conn{userID: userID, send: make(chan outbound, 64), rooms: map[uuid.UUID]bool{}}
	h.addSocket(c)

	go h.writePump(ws, c)
	h.readPump(ctx, ws, c, publishTyping)
}

func (h *Hub) readPump(ctx context.Context, ws *websocket.Conn, c *conn, publishTyping func(convID, uid uuid.UUID)) {
	defer func() {
		h.removeSocket(c)
		_ = ws.Close()
	}()
	ws.SetReadLimit(maxMessageSize)
	_ = ws.SetReadDeadline(time.Now().Add(pingPeriod + pongWait))
	ws.SetPongHandler(func(string) error {
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
