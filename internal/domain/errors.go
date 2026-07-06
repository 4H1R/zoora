package domain

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrForbidden     = errors.New("forbidden")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrConflict      = errors.New("resource already exists")
	ErrValidation    = errors.New("validation failed")
	ErrInternal      = errors.New("internal server error")
	ErrUserDisabled  = errors.New("account is disabled")
	ErrAccountLocked = errors.New("account is locked")
	ErrRateLimited   = errors.New("rate limit exceeded")

	ErrInvalidParticipantRole = errors.New("invalid participant role")
	ErrParticipantNotFound    = errors.New("participant not found")
	ErrCannotChangeHostRole   = errors.New("cannot change the host role")
	ErrCannotRemoveHost       = errors.New("cannot remove a host from the room")
	ErrCannotRemoveSelf       = errors.New("cannot remove yourself from the room")
	ErrWhiteboardNotFound     = errors.New("whiteboard not found")

	// Plan / entitlement gates (mapped to HTTP 402 Payment Required).
	ErrFeatureNotInPlan = errors.New("feature not available on current plan")
	ErrPlanLimitReached = errors.New("plan limit reached")
)

// PlanError carries machine-readable context for a plan/entitlement gate so the
// frontend can render an upgrade prompt. Unwraps to its sentinel so
// errors.Is(err, ErrFeatureNotInPlan) / ErrPlanLimitReached works.
type PlanError struct {
	Sentinel error  `json:"-"` // ErrFeatureNotInPlan or ErrPlanLimitReached
	Plan     string `json:"plan"`
	Feature  string `json:"feature,omitempty"` // set for feature gates
	Limit    string `json:"limit,omitempty"`   // set for limit gates
	Current  int64  `json:"current,omitempty"`
	Ceiling  int64  `json:"ceiling,omitempty"`
}

func (e *PlanError) Error() string {
	if e.Feature != "" {
		return fmt.Sprintf("feature %q not available on plan %q", e.Feature, e.Plan)
	}
	return fmt.Sprintf("limit %q reached on plan %q (%d/%d)", e.Limit, e.Plan, e.Current, e.Ceiling)
}

func (e *PlanError) Unwrap() error { return e.Sentinel }

// NewFeatureError builds a 402 feature-gate error.
func NewFeatureError(plan Plan, f Feature) *PlanError {
	return &PlanError{Sentinel: ErrFeatureNotInPlan, Plan: string(plan), Feature: string(f)}
}

// NewLimitError builds a 402 limit-gate error.
func NewLimitError(plan Plan, l Limit, current, ceiling int64) *PlanError {
	return &PlanError{Sentinel: ErrPlanLimitReached, Plan: string(plan), Limit: string(l), Current: current, Ceiling: ceiling}
}

// ValidationError carries per-field validation failures.
// Unwraps to ErrValidation so errors.Is(err, ErrValidation) works.
type ValidationError struct {
	Fields map[string]string
	Cause  error
}

func NewValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{Fields: fields}
}

func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		if e.Cause != nil {
			return e.Cause.Error()
		}
		return ErrValidation.Error()
	}
	parts := make([]string, 0, len(e.Fields))
	for k, v := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

func (e *ValidationError) Unwrap() error { return ErrValidation }

// MapError maps a domain sentinel error to an HTTP status and stable error code.
func MapError(err error) (int, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, ErrUserDisabled):
		return http.StatusForbidden, "USER_DISABLED"
	case errors.Is(err, ErrAccountLocked):
		return http.StatusTooManyRequests, "ACCOUNT_LOCKED"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, ErrInvalidSlug):
		return http.StatusUnprocessableEntity, "INVALID_SLUG"
	case errors.Is(err, ErrParticipantNotFound):
		return http.StatusNotFound, "PARTICIPANT_NOT_FOUND"
	case errors.Is(err, ErrInvalidParticipantRole):
		return http.StatusUnprocessableEntity, "INVALID_PARTICIPANT_ROLE"
	case errors.Is(err, ErrCannotChangeHostRole):
		return http.StatusConflict, "CANNOT_CHANGE_HOST_ROLE"
	case errors.Is(err, ErrCannotRemoveHost):
		return http.StatusConflict, "CANNOT_REMOVE_HOST"
	case errors.Is(err, ErrCannotRemoveSelf):
		return http.StatusConflict, "CANNOT_REMOVE_SELF"
	case errors.Is(err, ErrWhiteboardNotFound):
		return http.StatusNotFound, "WHITEBOARD_NOT_FOUND"
	case errors.Is(err, ErrFeatureNotInPlan):
		return http.StatusPaymentRequired, "FEATURE_NOT_IN_PLAN"
	case errors.Is(err, ErrPlanLimitReached):
		return http.StatusPaymentRequired, "PLAN_LIMIT_REACHED"
	case errors.Is(err, ErrRateLimited):
		return http.StatusTooManyRequests, "RATE_LIMITED"
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest, "VALIDATION_ERROR"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
