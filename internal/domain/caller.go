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

// HasAny reports whether the caller is an admin or holds at least one of the
// given permissions. Admins always pass.
func (c Caller) HasAny(perms ...PermissionName) bool {
	if c.IsAdmin {
		return true
	}
	for _, p := range perms {
		if c.HasPermission(p) {
			return true
		}
	}
	return false
}

// CanManage implements the owner-or-any authz tier: admins and holders of the
// org-wide anyPerm always pass; otherwise the caller must own the resource.
func (c Caller) CanManage(ownerID uuid.UUID, anyPerm PermissionName) bool {
	if c.IsAdmin || c.HasPermission(anyPerm) {
		return true
	}
	return c.UserID == ownerID
}

// CanManageOwned is the stricter tier used where owning a resource is not
// enough on its own: admins and anyPerm holders pass org-wide, while everyone
// else needs the scoped basePerm AND ownership of the resource.
func (c Caller) CanManageOwned(ownerID uuid.UUID, basePerm, anyPerm PermissionName) bool {
	if c.IsAdmin || c.HasPermission(anyPerm) {
		return true
	}
	return c.HasPermission(basePerm) && c.UserID == ownerID
}

type callerKey struct{}

func WithCaller(ctx context.Context, c Caller) context.Context {
	return context.WithValue(ctx, callerKey{}, c)
}

func CallerFromCtx(ctx context.Context) (Caller, bool) {
	c, ok := ctx.Value(callerKey{}).(Caller)
	return c, ok
}
