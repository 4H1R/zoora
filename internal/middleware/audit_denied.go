package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// mutatingMethods are the HTTP methods whose denial is worth auditing. GET/HEAD
// denials are reads — excluded by decision.
var mutatingMethods = map[string]struct{}{
	http.MethodPost:   {},
	http.MethodPut:    {},
	http.MethodPatch:  {},
	http.MethodDelete: {},
}

// routeSegments maps the leading gin route segment (after "v1") to a closed-set
// AuditTargetType. Denied entries need a security-signal target, not a precise
// resource id — an unmapped segment falls through to AuditTargetType(segment).
//
// Built from the same resources as the AuditTargetType enum; keys are the real
// URL segments used by each feature's RegisterRoutes.
var routeSegments = map[string]domain.AuditTargetType{
	"classes":                  domain.AuditTargetClass,
	"members":                  domain.AuditTargetEnrollment,
	"users":                    domain.AuditTargetUser,
	"roles":                    domain.AuditTargetRole,
	"role":                     domain.AuditTargetRole,
	"quizzes":                  domain.AuditTargetQuiz,
	"question-banks":           domain.AuditTargetQuestionBank,
	"question_banks":           domain.AuditTargetQuestionBank,
	"gradebook":                domain.AuditTargetGradebook,
	"billing":                  domain.AuditTargetBilling,
	"invoices":                 domain.AuditTargetBilling,
	"live-rooms":               domain.AuditTargetLiveSession,
	"offlines":                 domain.AuditTargetOffline,
	"practices":                domain.AuditTargetPractice,
	"attendance":               domain.AuditTargetAttendance,
	"settings":                 domain.AuditTargetOrgSettings,
	"organizations":            domain.AuditTargetOrganization,
	"custom-field-definitions": domain.AuditTargetCustomField,
	"custom-fields":            domain.AuditTargetCustomField,
	"connectors":               domain.AuditTargetConnector,
	"tickets":                  domain.AuditTargetTicket,
	"calendar":                 domain.AuditTargetCalendarEvent,
	"polls":                    domain.AuditTargetPoll,
	"qa":                       domain.AuditTargetQA,
	"imports":                  domain.AuditTargetImport,
	"media":                    domain.AuditTargetMedia,
}

// AuditDenied records a best-effort 'denied' audit entry when a mutating
// request resolves to 403 — whether the 403 came from a permission middleware
// (writes the response directly) or a service returning ErrForbidden (mapped by
// ErrorHandler). It runs AFTER the handler chain, but the outer ErrorHandler's
// post-Next phase runs LATER than this inner middleware's, so a service-layer
// ErrForbidden is still only an attached c.Error here — the status is not yet
// 403. To stay ordering-independent it treats the request as denied when the
// final status is 403 OR an attached error maps to 403. Soft-fail: a recorder
// error is logged, never surfaced to the client.
func AuditDenied(recorder domain.AuditRecorder, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if !requestDenied(c) {
			return
		}
		if _, ok := mutatingMethods[c.Request.Method]; !ok {
			return
		}
		caller, ok := domain.CallerFromCtx(c.Request.Context())
		if !ok || caller.OrgID == nil {
			// Unauthenticated or admin-host denials have no org log to file under.
			return
		}

		targetType := routeSegmentToTargetType(c.FullPath())
		var targetID *uuid.UUID
		if raw := c.Param("id"); raw != "" {
			if id, err := uuid.Parse(raw); err == nil {
				targetID = &id
			}
		}

		rec := domain.AuditRecord{
			Action:     denyActionForMethod(c.Request.Method),
			TargetType: targetType,
			TargetID:   targetID,
			Metadata: map[string]any{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			},
		}
		// RecordDenied is best-effort; no tx here (the action never ran).
		if err := recorder.RecordDenied(c.Request.Context(), rec); err != nil {
			logger.WarnContext(c.Request.Context(), "audit: failed to record denied attempt",
				"err", err, "path", c.Request.URL.Path)
		}
	}
}

// requestDenied reports whether the finished request was a 403 denial. It is
// ordering-independent: it returns true if the final status is already 403
// (permission middleware wrote it inline) OR any attached c.Error maps to a 403
// (a service returned domain.ErrForbidden / ErrUserDisabled, which the OUTER
// ErrorHandler will map to 403 in its own — later — post-Next phase). The two
// signals can overlap on a single request; callers must record only once.
func requestDenied(c *gin.Context) bool {
	if c.Writer.Status() == http.StatusForbidden {
		return true
	}
	for _, e := range c.Errors {
		if errors.Is(e.Err, domain.ErrForbidden) {
			return true
		}
		if status, _ := domain.MapError(e.Err); status == http.StatusForbidden {
			return true
		}
	}
	return false
}

// routeSegmentToTargetType maps a gin route like "/api/v1/classes/:id" to a
// best-effort target type via an explicit lookup on the first segment after
// "v1". Unmapped segments are stored as-is (AuditTargetType(segment)) so the
// entry still carries a usable security signal.
func routeSegmentToTargetType(fullPath string) domain.AuditTargetType {
	parts := strings.Split(strings.Trim(fullPath, "/"), "/")
	for i, p := range parts {
		if p == "v1" && i+1 < len(parts) {
			seg := parts[i+1]
			if t, ok := routeSegments[seg]; ok {
				return t
			}
			return domain.AuditTargetType(seg)
		}
	}
	return domain.AuditTargetType("unknown")
}

func denyActionForMethod(method string) domain.AuditAction {
	switch method {
	case http.MethodDelete:
		return domain.AuditDeleted
	case http.MethodPost:
		return domain.AuditCreated
	default:
		return domain.AuditUpdated
	}
}
