package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CalendarEventType identifies which schedulable room produced an event.
type CalendarEventType string

const (
	CalendarEventLive     CalendarEventType = "live"
	CalendarEventQuiz     CalendarEventType = "quiz"
	CalendarEventPractice CalendarEventType = "practice"
	CalendarEventOffline  CalendarEventType = "offline"
)

// CalendarEvent is a flattened, read-only view of one schedulable room for the
// calendar. ID is the room's own id (unique React key). EntityID is the id the
// frontend deep-links to for this type (live room, quiz, offline room, or the
// parent class session for practice). Times are UTC; the client buckets into
// days using its local timezone.
type CalendarEvent struct {
	ID             uuid.UUID         `json:"id"`
	Type           CalendarEventType `json:"type"`
	Title          string            `json:"title"`
	ClassID        uuid.UUID         `json:"class_id"`
	ClassName      string            `json:"class_name"`
	ClassSessionID uuid.UUID         `json:"class_session_id"`
	EntityID       uuid.UUID         `json:"entity_id"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        *time.Time        `json:"end_time,omitempty"`
}

// CalendarEventsData is the payload wrapped by Response.Data for the calendar
// endpoint. A named struct (not a bare slice) keeps the generated TS client
// shape stable and self-describing.
type CalendarEventsData struct {
	Events []CalendarEvent `json:"events"`
}

// CalendarRange is the validated, inclusive UTC window a calendar query covers.
type CalendarRange struct {
	From time.Time
	To   time.Time
}

// CalendarRepository reads schedulable rooms across features for the calendar.
// It reuses the shared ClassListScope (resolved via authz.ListScope) rather
// than a calendar-specific scope type.
type CalendarRepository interface {
	ListEvents(ctx context.Context, scope ClassListScope, r CalendarRange) ([]CalendarEvent, error)
}

// CalendarService resolves the caller's scope and returns events in range.
type CalendarService interface {
	ListEvents(ctx context.Context, r CalendarRange) ([]CalendarEvent, error)
}
