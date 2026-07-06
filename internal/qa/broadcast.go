package qa

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

const (
	qaEventCreated = "qa_question_created"
	qaEventVoted   = "qa_question_voted"
	qaEventStatus  = "qa_status_changed"
)

// roomDataSender is the narrow LiveKit surface used for realtime fanout.
// *livekit.Client satisfies it. Kept local so qa depends on one method.
type roomDataSender interface {
	SendData(ctx context.Context, roomName string, payload []byte, destinationIdentities []string) error
}

// broadcaster pushes QA events over the LiveKit reliable data channel of the
// room backing a live_session question, so participants receive updates
// instantly instead of polling. Best-effort: the DB row is already the source
// of truth, so failures are only logged. No-op when wiring is absent or the
// model is not a live session.
type broadcaster struct {
	livekit   roomDataSender
	liveRooms domain.LiveRoomRepository
	logger    *slog.Logger
}

// NewBroadcaster builds the realtime broadcaster. Pass nil deps (worker/tests)
// to disable broadcasting.
func NewBroadcaster(livekit roomDataSender, liveRooms domain.LiveRoomRepository, logger *slog.Logger) *broadcaster {
	return &broadcaster{livekit: livekit, liveRooms: liveRooms, logger: logger}
}

func (b *broadcaster) send(ctx context.Context, modelType string, modelID uuid.UUID, eventType string, data any) {
	if b == nil || b.livekit == nil || b.liveRooms == nil || modelType != domain.QAModelLiveSession {
		return
	}
	// modelID is a LiveRoom ID for live_session questions.
	room, err := b.liveRooms.FindByID(ctx, modelID)
	if err != nil {
		b.logger.Error("qa.broadcast: resolve room", "error", err)
		return
	}
	payload, err := json.Marshal(map[string]any{"type": eventType, "data": data})
	if err != nil {
		b.logger.Error("qa.broadcast: marshal", "event", eventType, "error", err)
		return
	}
	if err := b.livekit.SendData(ctx, room.LiveKitRoomName, payload, nil); err != nil {
		b.logger.Error("qa.broadcast: send", "room", room.LiveKitRoomName, "event", eventType, "error", err)
	}
}

func (s *service) broadcastCreated(ctx context.Context, q *domain.QAQuestion, caller domain.Caller) {
	if s.broadcaster == nil {
		return
	}
	s.broadcaster.send(ctx, q.ModelType, q.ModelID, qaEventCreated, map[string]any{
		"id":          q.ID,
		"user_id":     q.UserID,
		"author_name": caller.Name,
		"text":        q.Text,
		"status":      q.Status,
		"vote_count":  0,
		"created_at":  q.CreatedAt,
	})
}

func (s *service) broadcastVoted(ctx context.Context, q *domain.QAQuestion, count int64) {
	if s.broadcaster == nil {
		return
	}
	s.broadcaster.send(ctx, q.ModelType, q.ModelID, qaEventVoted, map[string]any{
		"id":         q.ID,
		"vote_count": count,
	})
}

func (s *service) broadcastStatus(ctx context.Context, q *domain.QAQuestion, status string) {
	if s.broadcaster == nil {
		return
	}
	s.broadcaster.send(ctx, q.ModelType, q.ModelID, qaEventStatus, map[string]any{
		"id":     q.ID,
		"status": status,
	})
}
