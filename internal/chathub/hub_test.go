package chathub

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type fakeMembers struct{ ok bool }

func (f fakeMembers) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) { return f.ok, nil }
func (f fakeMembers) ListUserIDs(context.Context, uuid.UUID) ([]uuid.UUID, error)  { return nil, nil }

func TestDeliverToRoom_ReachesJoinedSocket(t *testing.T) {
	h := NewHub(fakeMembers{ok: true}, testLogger())
	convID := uuid.New()
	c := &conn{userID: uuid.New(), send: make(chan outbound, 4), rooms: map[uuid.UUID]bool{}}
	h.addSocket(c)
	h.joinRoom(c, convID)

	h.deliverToRoom(convID, []byte(`{"type":"new_message"}`))

	select {
	case msg := <-c.send:
		if string(msg.data) != `{"type":"new_message"}` {
			t.Fatalf("unexpected payload: %s", msg.data)
		}
	default:
		t.Fatal("expected delivery to joined socket")
	}
}

func TestDeliverToRoom_SkipsNonJoiner(t *testing.T) {
	h := NewHub(fakeMembers{}, testLogger())
	c := &conn{userID: uuid.New(), send: make(chan outbound, 4), rooms: map[uuid.UUID]bool{}}
	h.addSocket(c)
	h.deliverToRoom(uuid.New(), []byte(`x`))
	select {
	case <-c.send:
		t.Fatal("non-joiner should not receive")
	default:
	}
}

// TestRemoveSocket_ReportsRoomsAndLastSocket covers the presence lifecycle
// signal: removeSocket returns the rooms the socket had joined and whether it
// was the user's last socket (so the caller marks offline exactly once).
func TestRemoveSocket_ReportsRoomsAndLastSocket(t *testing.T) {
	h := NewHub(fakeMembers{ok: true}, testLogger())
	userID := uuid.New()
	convA, convB := uuid.New(), uuid.New()

	c1 := &conn{userID: userID, send: make(chan outbound, 4), rooms: map[uuid.UUID]bool{}}
	c2 := &conn{userID: userID, send: make(chan outbound, 4), rooms: map[uuid.UUID]bool{}}
	h.addSocket(c1)
	h.addSocket(c2)
	h.joinRoom(c1, convA)
	h.joinRoom(c1, convB)

	// First socket leaves: reports its rooms but is NOT the last socket.
	rooms, last := h.removeSocket(c1)
	if last {
		t.Fatal("removing c1 while c2 remains must not report lastSocket")
	}
	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms reported, got %d", len(rooms))
	}
	got := map[uuid.UUID]bool{}
	for _, r := range rooms {
		got[r] = true
	}
	if !got[convA] || !got[convB] {
		t.Fatalf("expected rooms convA and convB, got %v", rooms)
	}

	// Second socket leaves: it joined no rooms, and IS the last socket.
	rooms, last = h.removeSocket(c2)
	if !last {
		t.Fatal("removing the final socket must report lastSocket=true")
	}
	if len(rooms) != 0 {
		t.Fatalf("c2 joined no rooms, expected 0 reported, got %d", len(rooms))
	}
}
