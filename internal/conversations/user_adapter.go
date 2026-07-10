package conversations

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// UserOrgLookup adapts domain.UserRepository to the service's userDirectory
// port, used to enforce the cross-org DM/member guards (Phase 3 Step 6).
type UserOrgLookup struct {
	users domain.UserRepository
}

func NewUserOrgLookup(users domain.UserRepository) *UserOrgLookup {
	return &UserOrgLookup{users: users}
}

// FilterSameOrg returns the subset of ids belonging to users in orgID, in one
// query — the batch form of the cross-org guard.
func (a *UserOrgLookup) FilterSameOrg(ctx context.Context, orgID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
	return a.users.FilterIDsInOrg(ctx, orgID, ids)
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
