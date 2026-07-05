package domain

import (
	"context"
	"slices"

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
	// Ent is the effective entitlement snapshot for the caller's org, resolved
	// at auth-middleware time (Free for admins / unauthenticated). Mirrors how
	// Permissions is loaded — see internal/auth/middleware.go.
	Ent Entitlements
}

// HasFeature reports whether the caller's plan includes feature f.
func (c Caller) HasFeature(f Feature) bool { return c.Ent.Can(f) }

// Limit returns the ceiling for l on the caller's plan (0 = unlimited).
func (c Caller) Limit(l Limit) int64 { return c.Ent.Limit(l) }

func (c Caller) HasPermission(perm PermissionName) bool {
	return slices.Contains(c.Permissions, string(perm))
}

// HasAny reports whether the caller is an admin or holds at least one of the
// given permissions. Admins always pass.
func (c Caller) HasAny(perms ...PermissionName) bool {
	if c.IsAdmin {
		return true
	}
	return slices.ContainsFunc(perms, c.HasPermission)
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
