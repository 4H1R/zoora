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
	class := &domain.Class{UserID: owner}
	const anyPerm = domain.PermGradebookViewAny

	tests := []struct {
		name     string
		caller   domain.Caller
		isMember bool
		want     Scope
	}{
		{"admin -> all", domain.Caller{UserID: stranger, IsAdmin: true}, false, ScopeAll},
		{"any-perm holder -> all", domain.Caller{UserID: stranger, Permissions: []string{string(anyPerm)}}, false, ScopeAll},
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
