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
