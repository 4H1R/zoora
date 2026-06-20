package domain

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("resource already exists")
	ErrValidation   = errors.New("validation failed")
	ErrInternal     = errors.New("internal server error")
	ErrUserDisabled = errors.New("account is disabled")
)

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
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest, "VALIDATION_ERROR"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
