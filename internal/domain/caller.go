package domain

import (
	"context"

	"github.com/google/uuid"
)

type Caller struct {
	UserID      uuid.UUID
	OrgID       *uuid.UUID
	IsAdmin     bool
	RoleID      *uuid.UUID
	Permissions []string
	Username    string
	Name        string
}

func (c Caller) HasPermission(perm PermissionName) bool {
	for _, p := range c.Permissions {
		if p == string(perm) {
			return true
		}
	}
	return false
}

type callerKey struct{}

func WithCaller(ctx context.Context, c Caller) context.Context {
	return context.WithValue(ctx, callerKey{}, c)
}

func CallerFromCtx(ctx context.Context) (Caller, bool) {
	c, ok := ctx.Value(callerKey{}).(Caller)
	return c, ok
}
