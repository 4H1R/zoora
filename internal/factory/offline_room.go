package factory

import (
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewOfflineRoom(orgID, classID, sessionID, creatorID uuid.UUID, opts ...func(*domain.OfflineRoom)) *domain.OfflineRoom {
	id := nextID()
	now := time.Now()
	r := &domain.OfflineRoom{
		OrganizationID: orgID,
		ClassID:        classID,
		ClassSessionID: sessionID,
		CreatorID:      creatorID,
		Title:          fakeOfflineRoomTitle(id),
		Description:    fakeSentence(8),
		PublishedAt:    &now,
		ViewCount:      int64(fake.IntRange(0, 500)),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func NewOfflineRoomView(roomID, userID uuid.UUID) *domain.OfflineRoomView {
	return &domain.OfflineRoomView{
		OfflineRoomID:   roomID,
		UserID:          userID,
		ViewedAt:        time.Now(),
		DurationSeconds: fake.IntRange(30, 1800),
	}
}
