package domain

import "github.com/google/uuid"

const (
	TypeLiveSessionAutoClose     = "livesession:auto-close"
	TypeLiveSessionCloseIfNoHost = "livesession:close-if-no-host"
	TypeAttendanceAutoMark       = "attendance:auto-mark"
)

// LiveSessionCloseIfNoHostPayload is the Asynq payload for the delayed,
// webhook-triggered auto-close. Enqueued (with a room-scoped TaskID so it is
// idempotent and cancelable) when a room's last host leaves; when it fires the
// service re-checks host presence against LiveKit before closing.
type LiveSessionCloseIfNoHostPayload struct {
	RoomID uuid.UUID `json:"room_id"`
}

// AttendanceAutoMarkPayload is the Asynq payload for session-scoped live
// auto-mark, enqueued when a live room finishes.
type AttendanceAutoMarkPayload struct {
	ClassID   uuid.UUID `json:"class_id"`
	SessionID uuid.UUID `json:"session_id"`
}
