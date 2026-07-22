package organizations_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
)

// TestOrganizationUpdate_CrossTenant_Forbidden verifies a non-admin holding
// organizations:update cannot rewrite an org that is not their own (the slug is
// the subdomain routing key, so this is a tenant-takeover vector).
func TestOrganizationUpdate_CrossTenant_Forbidden(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)

	orgID := uuid.New()
	otherOrgID := uuid.New()
	newName := "Hijacked"

	// Non-admin whose OrgID does not match the target id.
	ctx := orgCaller(uuid.New(), &otherOrgID, false)
	_, err := svc.Update(ctx, orgID, domain.UpdateOrganizationDTO{Name: &newName})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	// Non-admin with no org at all is likewise rejected.
	ctxNoOrg := orgCaller(uuid.New(), nil, false)
	_, err = svc.Update(ctxNoOrg, orgID, domain.UpdateOrganizationDTO{Name: &newName})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestOrganizationUpdate_Admin_AnyOrg verifies an admin bypasses tenant scoping.
func TestOrganizationUpdate_Admin_AnyOrg(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)

	orgID := uuid.New()
	newName := "New"
	repo.On("FindByID", mock.Anything, orgID).Return(&domain.Organization{ID: orgID, Name: "Old"}, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(o *domain.Organization) bool {
		return o.ID == orgID && o.Name == "New"
	})).Return(nil)

	adminCtx := orgCaller(uuid.New(), nil, true)
	updated, err := svc.Update(adminCtx, orgID, domain.UpdateOrganizationDTO{Name: &newName})
	assert.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	repo.AssertExpectations(t)
}
