package authz

import (
	"testing"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func TestDecideScope(t *testing.T) {
	owner := uuid.New()
	student := uuid.New()
	stranger := uuid.New()
	classOrg := uuid.New()
	otherOrg := uuid.New()
	class := &domain.Class{UserID: owner, OrganizationID: classOrg}
	const anyPerm = domain.PermGradebookViewAny

	// perms builds a caller holding anyPerm, scoped to org (nil = no org).
	anyHolder := func(uid uuid.UUID, org *uuid.UUID) domain.Caller {
		return domain.Caller{UserID: uid, OrgID: org, Permissions: []string{string(anyPerm)}}
	}

	tests := []struct {
		name     string
		caller   domain.Caller
		isMember bool
		want     Scope
	}{
		{"admin -> all (org ignored)", domain.Caller{UserID: stranger, IsAdmin: true, OrgID: &otherOrg}, false, ScopeAll},
		{"any-perm holder same org -> all", anyHolder(stranger, &classOrg), false, ScopeAll},
		{"any-perm holder cross-org, stranger -> none", anyHolder(stranger, &otherOrg), false, ScopeNone},
		{"any-perm holder cross-org, owner -> class", anyHolder(owner, &otherOrg), false, ScopeClass},
		{"any-perm holder cross-org, member -> own", anyHolder(student, &otherOrg), true, ScopeOwn},
		{"any-perm holder nil org -> not elevated (none)", anyHolder(stranger, nil), false, ScopeNone},
		{"class owner -> class", domain.Caller{UserID: owner}, false, ScopeClass},
		{"enrolled student -> own", domain.Caller{UserID: student}, true, ScopeOwn},
		{"stranger -> none", domain.Caller{UserID: stranger}, false, ScopeNone},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := decideScope(tc.caller, class, tc.isMember, anyPerm); got != tc.want {
				t.Fatalf("decideScope = %v, want %v", got, tc.want)
			}
		})
	}
}
