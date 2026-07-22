// Package authz centralizes relational access-scope resolution so feature
// services stop hand-rolling the "admin / _any / owner / member" ladder.
package authz

import (
	"context"
	"slices"

	"github.com/4H1R/zoora/internal/domain"
)

// Scope is the row-visibility a caller has over a class-scoped resource.
type Scope int

const (
	ScopeNone  Scope = iota // no access
	ScopeOwn                // only rows the caller owns (enrolled student)
	ScopeClass              // every row in the class (teacher/owner)
	ScopeAll                // org-wide (admin or _any holder)
)

// hasOrgScopedAny reports whether the caller may act org-wide on class: admin
// always, or an anyPerm holder whose org matches the class's org. This mirrors
// ListScope's tenancy guard so a cross-org _any holder never elevates to
// ScopeAll on a single object belonging to another org.
func hasOrgScopedAny(caller domain.Caller, class *domain.Class, anyPerm domain.PermissionName) bool {
	if caller.IsAdmin {
		return true
	}
	return caller.HasPermission(anyPerm) &&
		caller.OrgID != nil &&
		*caller.OrgID == class.OrganizationID
}

// decideScope is the pure decision over already-fetched data. anyPerm is the
// org-wide elevation permission for the resource being checked.
func decideScope(caller domain.Caller, class *domain.Class, isMember bool, anyPerm domain.PermissionName) Scope {
	if hasOrgScopedAny(caller, class, anyPerm) {
		return ScopeAll
	}
	if caller.UserID == class.UserID {
		return ScopeClass
	}
	if isMember {
		return ScopeOwn
	}
	return ScopeNone
}

// ListScope resolves the multi-class list filter for a caller — the list
// analogue of Scope, returning the domain.ClassListScope the class-scoped
// repositories understand. anyPerms are the org-wide elevation permissions for
// the resource. Org-wide access additionally requires a non-nil OrgID so an
// *_any holder without an org context never falls through to an unfiltered
// (cross-tenant) scan; such a caller degrades to the classes they teach or are
// enrolled in.
func ListScope(caller domain.Caller, anyPerms ...domain.PermissionName) domain.ClassListScope {
	if caller.IsAdmin {
		return domain.ClassListScope{All: true}
	}
	if caller.OrgID != nil && slices.ContainsFunc(anyPerms, caller.HasPermission) {
		return domain.ClassListScope{All: true, OrganizationID: caller.OrgID}
	}
	uid := caller.UserID
	return domain.ClassListScope{TeacherID: &uid, MemberUserID: &uid}
}

// Resolver wraps decideScope with the membership lookup it needs.
type Resolver struct {
	members domain.ClassMemberRepository
}

func NewResolver(members domain.ClassMemberRepository) *Resolver {
	return &Resolver{members: members}
}

// Scope resolves the caller's scope over class. It only hits the DB when the
// outcome actually depends on enrollment (i.e. not admin/_any/owner).
func (r *Resolver) Scope(ctx context.Context, caller domain.Caller, class *domain.Class, anyPerm domain.PermissionName) (Scope, error) {
	if hasOrgScopedAny(caller, class, anyPerm) {
		return ScopeAll, nil
	}
	if caller.UserID == class.UserID {
		return ScopeClass, nil
	}
	isMember, err := r.members.Exists(ctx, class.ID, caller.UserID)
	if err != nil {
		return ScopeNone, err
	}
	return decideScope(caller, class, isMember, anyPerm), nil
}
