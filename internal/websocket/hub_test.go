package websocket

import (
	"io"
	"log/slog"
	"testing"
)

func TestHubAddRemoveAndRoomClientCount(t *testing.T) {
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	client := &Client{send: make(chan *Message, 1), userID: "u1", room: "room-1"}

	hub.addClient(client)
	if got := hub.RoomClientCount("room-1"); got != 1 {
		t.Fatalf("RoomClientCount() = %d, want 1", got)
	}

	hub.removeClient(client)
	if got := hub.RoomClientCount("room-1"); got != 0 {
		t.Fatalf("RoomClientCount() after remove = %d, want 0", got)
	}
	if _, ok := <-client.send; ok {
		t.Fatal("client send channel is still open after remove")
	}
}

func TestHubEnforcesPerUserConnectionLimit(t *testing.T) {
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	clients := make([]*Client, 0, maxConnectionsPerUser+1)
	for i := 0; i < maxConnectionsPerUser+1; i++ {
		client := &Client{send: make(chan *Message, 1), userID: "u1", room: "room-1"}
		clients = append(clients, client)
		hub.addClient(client)
	}

	if got := hub.RoomClientCount("room-1"); got != maxConnectionsPerUser {
		t.Fatalf("RoomClientCount() = %d, want max %d", got, maxConnectionsPerUser)
	}
	if _, ok := <-clients[maxConnectionsPerUser].send; ok {
		t.Fatal("overflow client send channel is still open")
	}
}

func TestHubBroadcastToRoom(t *testing.T) {
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	c1 := &Client{send: make(chan *Message, 1), userID: "u1", room: "room-1"}
	c2 := &Client{send: make(chan *Message, 1), userID: "u2", room: "room-1"}
	otherRoom := &Client{send: make(chan *Message, 1), userID: "u3", room: "room-2"}
	hub.addClient(c1)
	hub.addClient(c2)
	hub.addClient(otherRoom)

	msg := &Message{Room: "room-1", Type: "chat", Payload: []byte(`{"text":"hello"}`)}
	hub.broadcastToRoom(msg)

	for _, client := range []*Client{c1, c2} {
		select {
		case got := <-client.send:
			if got != msg {
				t.Fatalf("broadcast message = %#v, want original message", got)
			}
		default:
			t.Fatalf("client %s did not receive broadcast", client.userID)
		}
	}

	select {
	case got := <-otherRoom.send:
		t.Fatalf("other room client received unexpected message %#v", got)
	default:
	}
}
