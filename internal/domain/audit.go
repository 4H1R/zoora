package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditActorSystemName is the reserved actor_name for mutations with no human
// Caller (worker jobs, plan-expiry downgrade, cascades). Such entries carry a
// nil actor_id.
const AuditActorSystemName = "System"

// AuditAction is a closed-set verb on an audit entry. Extend deliberately —
// adding a constant is the reviewed choke point for new auditable verbs.
type AuditAction string

const (
	AuditCreated AuditAction = "created"
	AuditUpdated AuditAction = "updated"
	AuditDeleted AuditAction = "deleted"
	// Domain verbs (add as needed by instrumentation):
	AuditEnrolled   AuditAction = "enrolled"
	AuditUnenrolled AuditAction = "unenrolled"
	AuditGraded     AuditAction = "graded"
	AuditDisabled   AuditAction = "disabled"
	AuditEnabled    AuditAction = "enabled"
)

var auditActions = map[AuditAction]struct{}{
	AuditCreated: {}, AuditUpdated: {}, AuditDeleted: {},
	AuditEnrolled: {}, AuditUnenrolled: {}, AuditGraded: {},
	AuditDisabled: {}, AuditEnabled: {},
}

func (a AuditAction) Valid() bool { _, ok := auditActions[a]; return ok }

// AuditTargetType is the closed-set kind of resource an entry concerns —
// one constant per auditable resource. AuditTargetTypes() drives the guard
// test that asserts every declared type is emitted somewhere.
type AuditTargetType string

const (
	AuditTargetClass         AuditTargetType = "class"
	AuditTargetEnrollment    AuditTargetType = "enrollment"
	AuditTargetUser          AuditTargetType = "user"
	AuditTargetRole          AuditTargetType = "role"
	AuditTargetQuiz          AuditTargetType = "quiz"
	AuditTargetQuestionBank  AuditTargetType = "question_bank"
	AuditTargetGradebook     AuditTargetType = "gradebook"
	AuditTargetBilling       AuditTargetType = "billing"
	AuditTargetLiveSession   AuditTargetType = "live_session"
	AuditTargetOffline       AuditTargetType = "offline_room"
	AuditTargetPractice      AuditTargetType = "practice"
	AuditTargetAttendance    AuditTargetType = "attendance"
	AuditTargetOrgSettings   AuditTargetType = "org_settings"
	AuditTargetOrganization  AuditTargetType = "organization"
	AuditTargetCustomField   AuditTargetType = "custom_field"
	AuditTargetConnector     AuditTargetType = "connector"
	AuditTargetTicket        AuditTargetType = "ticket"
	AuditTargetCalendarEvent AuditTargetType = "calendar_event"
	AuditTargetPoll          AuditTargetType = "poll"
	AuditTargetQA            AuditTargetType = "qa"
	AuditTargetImport        AuditTargetType = "import"
	AuditTargetMedia         AuditTargetType = "media"
)

// AuditTargetTypes is the authoritative list of every auditable target type.
// The guard test asserts each is actually emitted by some service path.
func AuditTargetTypes() []AuditTargetType {
	return []AuditTargetType{
		AuditTargetClass, AuditTargetEnrollment, AuditTargetUser, AuditTargetRole,
		AuditTargetQuiz, AuditTargetQuestionBank, AuditTargetGradebook, AuditTargetBilling,
		AuditTargetLiveSession, AuditTargetOffline, AuditTargetPractice, AuditTargetAttendance,
		AuditTargetOrgSettings, AuditTargetOrganization, AuditTargetCustomField,
		AuditTargetConnector, AuditTargetTicket, AuditTargetCalendarEvent, AuditTargetPoll,
		AuditTargetQA, AuditTargetImport, AuditTargetMedia,
	}
}

func (t AuditTargetType) Valid() bool {
	for _, v := range AuditTargetTypes() {
		if v == t {
			return true
		}
	}
	return false
}

// AuditOutcome distinguishes a committed change from a blocked attempt.
type AuditOutcome string

const (
	AuditOutcomeSuccess AuditOutcome = "success"
	AuditOutcomeDenied  AuditOutcome = "denied"
)

func (o AuditOutcome) Valid() bool {
	return o == AuditOutcomeSuccess || o == AuditOutcomeDenied
}

// AuditEntry is one immutable record in an Organization's audit log.
type AuditEntry struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null;column:organization_id" json:"organization_id"`
	ActorID        *uuid.UUID      `gorm:"type:uuid;column:actor_id" json:"actor_id,omitempty"`
	ActorName      string          `gorm:"type:varchar(255);not null;default:''" json:"actor_name"`
	ActorUsername  string          `gorm:"type:varchar(255);not null;default:''" json:"actor_username"`
	Action         AuditAction     `gorm:"type:varchar(40);not null" json:"action"`
	TargetType     AuditTargetType `gorm:"type:varchar(64);not null" json:"target_type"`
	TargetID       *uuid.UUID      `gorm:"type:uuid;column:target_id" json:"target_id,omitempty"`
	TargetLabel    string          `gorm:"type:varchar(512);not null;default:''" json:"target_label"`
	Outcome        AuditOutcome    `gorm:"type:varchar(16);not null;default:'success'" json:"outcome"`
	Metadata       map[string]any  `gorm:"type:jsonb;serializer:json" json:"metadata"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (AuditEntry) TableName() string { return "audit_entries" }

// AuditRecord is the input a service passes to the recorder. Keeping it a
// struct keeps call sites to a single readable line and lets fields be added
// without breaking every caller.
type AuditRecord struct {
	Action      AuditAction
	TargetType  AuditTargetType
	TargetID    *uuid.UUID
	TargetLabel string
	// OrgID overrides the org the entry is filed under. Leave nil for the
	// common case (recorder uses the Caller's org). Set it for Platform Admin
	// cross-tenant actions and worker/System actions, where the caller's org is
	// absent or wrong — pass the TARGET's org.
	OrgID *uuid.UUID
	// Metadata holds forensic extras: field diffs, cascade counts, flags.
	Metadata map[string]any
}

// AuditRecorder is the narrow interface feature services depend on. Implemented
// by audit.service; injected via constructor so features never import audit.
type AuditRecorder interface {
	// Record writes a success entry. It reads actor + org from the Caller in
	// ctx and client IP/UA from RequestInfo in ctx, and joins the caller's DB
	// transaction via TxFromCtx. Callers MUST invoke it inside the same
	// RunInTx block as the change so the entry commits or rolls back with it.
	Record(ctx context.Context, r AuditRecord) error
}

// AuditListQuery filters the Manager read endpoint. All filters are optional.
type AuditListQuery struct {
	ListParams
	ActorID    *uuid.UUID
	Action     *AuditAction
	TargetType *AuditTargetType
	TargetID   *uuid.UUID
	Outcome    *AuditOutcome
	From       *time.Time
	To         *time.Time
}

type AuditRepository interface {
	Create(ctx context.Context, e *AuditEntry) error
	// List returns entries for the caller's org, newest first, paginated.
	List(ctx context.Context, orgID uuid.UUID, q AuditListQuery) ([]AuditEntry, int64, error)
}

type AuditService interface {
	AuditRecorder
	// List enforces the audit:view_any permission and org scoping.
	List(ctx context.Context, q AuditListQuery) ([]AuditEntry, int64, error)
}
