package domain

import "github.com/google/uuid"

// Asynq queue names. Weighted priority is configured on the worker server. Media
// isolates slow S3 / headless-Chrome work (cleanup, org purge, retention sweep,
// invoice PDF) so a burst of it can't starve latency-sensitive default tasks.
const (
	QueueCritical      = "critical"
	QueueDefault       = "default"
	QueueNotifications = "notifications"
	QueueMedia         = "media"
)

const (
	TypeLiveSessionAutoClose     = "livesession:auto-close"
	TypeLiveSessionCloseIfNoHost = "livesession:close-if-no-host"
	TypeAttendanceAutoMark       = "attendance:auto-mark"
	TypeMediaCleanup             = "media:cleanup"
	TypeOrganizationCleanup      = "organization:cleanup"
	TypeRecordingRetentionSweep  = "recording:retention-sweep"
	TypeNotificationFanout       = "notification:fanout"
	TypeNotificationDeliverBot   = "notification:deliver-bot"
	TypeNotificationDeliverSMS   = "notification:deliver-sms"
	TypeNotificationDeliverPush  = "notification:deliver-push"

	TypeInvoiceGeneratePDF   = "invoice:generate-pdf"
	TypeBillingReminderSweep = "billing:reminder-sweep"
	TypeBillingExpireSweep   = "billing:expire-sweep"

	TypeImportProcess = "import:process"

	TypeQuestionRenderImages = "question:render-images"
)

// QuestionRenderImagesPayload drives anti-cheat image generation for one
// question: the worker (re)renders the body text and every option value to
// PNGs, stores them as media, and flips the question's ImageRenderStatus. Also
// runs when RenderAsImage was just turned off, in which case it only purges the
// previously generated media. Idempotent: it deletes the prior system media
// before regenerating, so a retry or re-enqueue is safe.
type QuestionRenderImagesPayload struct {
	QuestionID uuid.UUID `json:"question_id"`
}

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

// OrganizationCleanupPayload is the Asynq payload for purging an org's S3
// objects after an admin hard-deletes it. Media rows are removed by the DB
// org FK cascade; this task drops the underlying storage under the org's
// key prefix (orgs/{org_id}/), which no FK covers.
type OrganizationCleanupPayload struct {
	OrganizationID uuid.UUID `json:"organization_id"`
}

// InvoiceGeneratePDFPayload renders and stores an invoice receipt PDF.
type InvoiceGeneratePDFPayload struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
}

// ImportProcessPayload snapshots the enqueueing caller's identity, permissions
// and plan — the worker ctx has no auth middleware, so authorization guards
// (Manager-role guard, seat limit) replay against this snapshot. The window
// between enqueue and processing is seconds; a permission revoked in that
// window is an accepted race.
type ImportProcessPayload struct {
	JobID       uuid.UUID `json:"job_id"`
	UserID      uuid.UUID `json:"user_id"`
	OrgID       uuid.UUID `json:"org_id"`
	IsAdmin     bool      `json:"is_admin"`
	Permissions []string  `json:"permissions"`
	Plan        Plan      `json:"plan"`
}
