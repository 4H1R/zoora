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
