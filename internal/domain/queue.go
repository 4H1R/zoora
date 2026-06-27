package domain

import "github.com/google/uuid"

const (
	TypeLiveSessionAutoClose = "livesession:auto-close"
	TypeAttendanceAutoMark   = "attendance:auto-mark"
)

// AttendanceAutoMarkPayload is the Asynq payload for session-scoped live
// auto-mark, enqueued when a live room finishes.
type AttendanceAutoMarkPayload struct {
	ClassID   uuid.UUID `json:"class_id"`
	SessionID uuid.UUID `json:"session_id"`
}
