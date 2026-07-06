package domain

import "github.com/google/uuid"

const (
	TypeLiveSessionAutoClose     = "livesession:auto-close"
	TypeLiveSessionCloseIfNoHost = "livesession:close-if-no-host"
	TypeAttendanceAutoMark       = "attendance:auto-mark"
	TypeMediaCleanup             = "media:cleanup"
	TypeRecordingRetentionSweep  = "recording:retention-sweep"
	TypeNotificationFanout       = "notification:fanout"
	TypeNotificationDeliverBot   = "notification:deliver-bot"
	TypeNotificationDeliverSMS   = "notification:deliver-sms"
	TypeNotificationDeliverPush  = "notification:deliver-push"
)

// NotificationFanoutPayload resolves a notification's audience to user IDs
// and inserts inbox rows. Retry-safe: recipient insert ignores conflicts.
type NotificationFanoutPayload struct {
	NotificationID uuid.UUID `json:"notification_id"`
}

// NotificationDeliverBotPayload sends one telegram/bale message (one chat per
// API call, so one task per delivery row).
type NotificationDeliverBotPayload struct {
	DeliveryID uuid.UUID `json:"delivery_id"`
}

// NotificationDeliverBatchPayload covers SMS (bulk provider call) and push
// (FCM multicast) — one task per batch of delivery rows.
type NotificationDeliverBatchPayload struct {
	NotificationID uuid.UUID   `json:"notification_id"`
	DeliveryIDs    []uuid.UUID `json:"delivery_ids"`
}

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

// MediaCleanupPayload is the Asynq payload for purging a polymorphic media
// collection (rows + underlying S3 objects). Enqueued, for example, when a
// live room finishes to drop the slides the host shared.
type MediaCleanupPayload struct {
	ModelType      string    `json:"model_type"`
	ModelID        uuid.UUID `json:"model_id"`
	CollectionName string    `json:"collection_name"`
}
