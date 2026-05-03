package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("resource already exists")
	ErrValidation   = errors.New("validation failed")
	ErrInternal     = errors.New("internal server error")
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
