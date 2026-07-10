package conversations

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// UserOrgLookup adapts domain.UserRepository to the service's userLookup
// port, used to enforce the cross-org DM/member guards (Phase 3 Step 6).
type UserOrgLookup struct {
	users domain.UserRepository
}

func NewUserOrgLookup(users domain.UserRepository) *UserOrgLookup {
	return &UserOrgLookup{users: users}
}

// OrgID returns the user's organization id, or nil if the user has none
// (e.g. a platform admin). Returns an error if the user cannot be found.
func (a *UserOrgLookup) OrgID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	u, err := a.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.OrganizationID, nil
}

// DirectorySearch returns member-safe projections of active org users matching
// query, for chat discovery.
func (a *UserOrgLookup) DirectorySearch(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]domain.DirectoryUser, error) {
	users, err := a.users.SearchActiveInOrg(ctx, orgID, query, limit)
	if err != nil {
		return nil, err
	}
	out := make([]domain.DirectoryUser, 0, len(users))
	for _, u := range users {
		out = append(out, domain.DirectoryUser{ID: u.ID, Name: u.Name, Username: u.Username})
	}
	return out, nil
}

// DirectoryByUsername resolves one active org user by exact username. findOne
// does NOT filter disabled_at, so we drop disabled users here to keep the
// directory consistent with search.
func (a *UserOrgLookup) DirectoryByUsername(ctx context.Context, orgID uuid.UUID, username string) (*domain.DirectoryUser, error) {
	u, err := a.users.FindByUsernameAndOrg(ctx, username, orgID)
	if err != nil {
		return nil, err
	}
	if u.DisabledAt != nil {
		return nil, domain.ErrNotFound
	}
	return &domain.DirectoryUser{ID: u.ID, Name: u.Name, Username: u.Username}, nil
}
