package audit_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
)

// emittedTargetTypes is the set of target types the codebase actually records on
// a service success path. UPDATE THIS DELIBERATELY when you instrument (or
// intentionally drop) a target type. It is the reviewed choke point that keeps
// coverage from silently rotting.
//
// Note: calendar_event is intentionally absent — see readOnlyExceptions below.
var emittedTargetTypes = map[domain.AuditTargetType]bool{
	domain.AuditTargetClass: true, domain.AuditTargetEnrollment: true,
	domain.AuditTargetUser: true, domain.AuditTargetRole: true,
	domain.AuditTargetQuiz: true, domain.AuditTargetQuestionBank: true,
	domain.AuditTargetGradebook: true, domain.AuditTargetBilling: true,
	domain.AuditTargetLiveSession: true, domain.AuditTargetOffline: true,
	domain.AuditTargetPractice: true, domain.AuditTargetAttendance: true,
	domain.AuditTargetOrgSettings: true, domain.AuditTargetOrganization: true,
	domain.AuditTargetCustomField: true, domain.AuditTargetConnector: true,
	domain.AuditTargetTicket: true, domain.AuditTargetPoll: true,
	domain.AuditTargetQA: true, domain.AuditTargetImport: true,
	domain.AuditTargetMedia: true,
}

// readOnlyExceptions are target types that are declared (and accepted by
// AuditTargetType.Valid() so the denied-attempt middleware can file 403s against
// them) but are never emitted on a success path because the feature is a
// read-only projection with no create/update/delete.
//
// calendar_event: the calendar is a read-only projection over classes,
// livesessions, quizzes, etc. Nothing creates/updates/deletes a calendar event
// directly, so no service success path emits it. The constant is kept because
// the denied-attempt middleware's route map uses it to file denied 403s against
// calendar routes.
var readOnlyExceptions = map[domain.AuditTargetType]bool{
	domain.AuditTargetCalendarEvent: true,
}

func TestEveryTargetTypeIsEmitted(t *testing.T) {
	for _, tt := range domain.AuditTargetTypes() {
		require.Truef(t, emittedTargetTypes[tt] || readOnlyExceptions[tt],
			"AuditTargetType %q is declared but neither emitted by a service success path nor listed as a read-only exception — instrument it, or add it to readOnlyExceptions with a justification (see docs/audit-coverage.md)", tt)
	}

	// No stale entries: every key in emittedTargetTypes must be a real member of
	// AuditTargetTypes(), otherwise the golden list has drifted from the enum.
	valid := make(map[domain.AuditTargetType]bool, len(domain.AuditTargetTypes()))
	for _, tt := range domain.AuditTargetTypes() {
		valid[tt] = true
	}
	for tt := range emittedTargetTypes {
		require.Truef(t, valid[tt],
			"emittedTargetTypes has entry %q that is not in AuditTargetTypes() — remove the stale entry (see docs/audit-coverage.md)", tt)
	}
}
