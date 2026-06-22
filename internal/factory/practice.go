package factory

import (
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewPracticeRoom(orgID, classID, sessionID, userID uuid.UUID, opts ...func(*domain.PracticeRoom)) *domain.PracticeRoom {
	id := nextID()
	start := time.Now().Add(-24 * time.Hour)
	end := start.Add(7 * 24 * time.Hour)
	r := &domain.PracticeRoom{
		OrganizationID: orgID,
		ClassID:        classID,
		ClassSessionID: sessionID,
		UserID:         userID,
		Title:          fakePracticeTitle(id),
		Content:        fakeSentence(12),
		MaxScore:       float64(fake.IntRange(10, 100)),
		StartTime:      start,
		EndTime:        end,
		Attachments:    []uuid.UUID{},
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func NewPracticeSubmission(roomID, userID uuid.UUID, opts ...func(*domain.PracticeSubmission)) *domain.PracticeSubmission {
	score := float64(fake.IntRange(0, 100))
	s := &domain.PracticeSubmission{
		PracticeRoomID: roomID,
		UserID:         userID,
		Content:        fakeSentence(10),
		Score:          &score,
		TeacherComment: fakeSentence(6),
		SubmittedAt:    time.Now(),
		Attachments:    []uuid.UUID{},
	}
	for _, o := range opts {
		o(s)
	}
	return s
}
