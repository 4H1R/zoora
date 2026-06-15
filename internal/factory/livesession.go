package factory

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewLiveRoom(sessionID uuid.UUID, opts ...func(*domain.LiveRoom)) *domain.LiveRoom {
	id := nextID()
	r := &domain.LiveRoom{
		ClassSessionID:  sessionID,
		LiveKitRoomName: fmt.Sprintf("room-%d-%s", id, uuid.New().String()[:8]),
		Status:          domain.LiveRoomStatusCreated,
		Config:          domain.DefaultLiveRoomConfig(),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func NewLiveParticipant(roomID, userID uuid.UUID, opts ...func(*domain.LiveParticipant)) *domain.LiveParticipant {
	p := &domain.LiveParticipant{
		LiveRoomID:           roomID,
		UserID:               userID,
		Identity:             fmt.Sprintf("user-%s", userID.String()[:8]),
		JoinedAt:             time.Now().Add(-time.Duration(fake.IntRange(10, 120)) * time.Minute),
		TotalDurationSeconds: fake.IntRange(60, 3600),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func NewLiveRecording(roomID uuid.UUID, opts ...func(*domain.LiveRecording)) *domain.LiveRecording {
	id := nextID()
	r := &domain.LiveRecording{
		LiveRoomID: roomID,
		EgressID:   fmt.Sprintf("EG_%d_%s", id, uuid.New().String()[:8]),
		Status:     domain.LiveRecordingStatusCompleted,
		FileURL:    fake.URL(),
		Duration:   fake.IntRange(300, 7200),
		Size:       int64(fake.IntRange(1024*1024, 1024*1024*500)),
		StartedAt:  time.Now().Add(-2 * time.Hour),
	}
	endedAt := r.StartedAt.Add(time.Duration(r.Duration) * time.Second)
	r.EndedAt = &endedAt
	for _, o := range opts {
		o(r)
	}
	return r
}
