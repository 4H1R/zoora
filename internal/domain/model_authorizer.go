package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrUnsupportedModelType is returned when a polymorphic feature is asked to
// authorize against a model_type it has no resolver for.
var ErrUnsupportedModelType = errors.New("unsupported model type")

// ModelAuthorizer answers authorization questions about a caller's relationship
// to a polymorphic model (identified by model_type + model_id). It lets
// polymorphic features (e.g. qa) make teacher-vs-participant decisions without
// importing the feature package that owns the model.
//
// Implementations resolve model_type to the owning entity and reuse that
// feature's existing authz rules. Unknown model_type -> ErrUnsupportedModelType.
type ModelAuthorizer interface {
	// CanParticipate reports whether the caller may read/interact with the model
	// (e.g. an enrolled student or the owning teacher of a live room).
	CanParticipate(ctx context.Context, caller Caller, modelType string, modelID uuid.UUID) (bool, error)
	// CanModerate reports whether the caller may perform host/teacher actions on
	// the model (e.g. resolve/dismiss questions in their own live room).
	CanModerate(ctx context.Context, caller Caller, modelType string, modelID uuid.UUID) (bool, error)
}
