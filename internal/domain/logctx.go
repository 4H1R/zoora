package domain

import "context"

// Correlation identifiers carried on the request/task context so the logger can
// tag every line without callers threading them through each log call.

type requestIDKey struct{}
type taskIDKey struct{}

// WithRequestID stores an HTTP request correlation id on the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromCtx returns the request correlation id, or "" if unset.
func RequestIDFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// WithTaskID stores an Asynq task correlation id on the context.
func WithTaskID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, taskIDKey{}, id)
}

// TaskIDFromCtx returns the task correlation id, or "" if unset.
func TaskIDFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(taskIDKey{}).(string)
	return id
}
