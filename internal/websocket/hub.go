package websocket

import (
	"log/slog"
	"sync"
)

const maxConnectionsPerUser = 3

type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.RWMutex
	logger     *slog.Logger
	done       chan struct{}
}

type Message struct {
	Room    string `json:"room"`
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
	Sender  string `json:"sender,omitempty"`
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		logger:     logger,
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case message := <-h.broadcast:
			h.broadcastToRoom(message)
		case <-h.done:
			return
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.done)

	h.mu.Lock()
	defer h.mu.Unlock()

	for room, clients := range h.rooms {
		for client := range clients {
			close(client.send)
		}
		delete(h.rooms, room)
	}

	h.logger.Info("websocket hub shut down")
}

func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[client.room]; !ok {
		h.rooms[client.room] = make(map[*Client]bool)
	}

	count := 0
	for c := range h.rooms[client.room] {
		if c.userID == client.userID {
			count++
		}
	}
	if count >= maxConnectionsPerUser {
		h.logger.Warn("connection limit reached",
			"user_id", client.userID,
			"room", client.room,
		)
		close(client.send)
		return
	}

	h.rooms[client.room][client] = true
	h.logger.Info("client registered",
		"user_id", client.userID,
		"room", client.room,
	)
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.room]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)

			if len(clients) == 0 {
				delete(h.rooms, client.room)
			}

			h.logger.Info("client unregistered",
				"user_id", client.userID,
				"room", client.room,
			)
		}
	}
}

func (h *Hub) broadcastToRoom(msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[msg.Room]
	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.send <- msg:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

func (h *Hub) RoomClientCount(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[room])
}
