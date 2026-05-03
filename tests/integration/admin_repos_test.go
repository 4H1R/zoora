//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

// setupAdminReposDB spins up Postgres + migrates schema. Returns both repos.
func setupAdminReposDB(t *testing.T) (domain.UserRepository, domain.OrganizationRepository) {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(&domain.Organization{}, &domain.User{}))
	return users.NewRepository(db), organizations.NewRepository(db)
}

// seedUser inserts a user with controlled flags.
func seedUser(t *testing.T, repo domain.UserRepository, orgID *uuid.UUID, username string, isAdmin bool) *domain.User {
	t.Helper()
	u := &domain.User{
		OrganizationID: orgID,
		Username:       username,
		Name:           username,
		Password:       "x",
		IsAdmin:        isAdmin,
	}
	require.NoError(t, repo.Create(context.Background(), u))
	return u
}

func TestIntegration_UserRepo_HardDelete_Removes(t *testing.T) {
	repo, _ := setupAdminReposDB(t)
	ctx := context.Background()
	u := seedUser(t, repo, nil, "hd1", false)

	require.NoError(t, repo.HardDelete(ctx, u.ID))

	_, err := repo.FindByIDIncludingDeleted(ctx, u.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// Double delete → ErrNotFound.
	err = repo.HardDelete(ctx, u.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_UserRepo_FindByIDIncludingDeleted_SeesSoftDeleted(t *testing.T) {
	repo, _ := setupAdminReposDB(t)
	ctx := context.Background()
	u := seedUser(t, repo, nil, "sd1", false)

	require.NoError(t, repo.Delete(ctx, u.ID))

	// Normal FindByID hides it.
	_, err := repo.FindByID(ctx, u.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// Unscoped find still returns it.
	got, err := repo.FindByIDIncludingDeleted(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, u.ID, got.ID)
	assert.True(t, got.DeletedAt.Valid)
}

func TestIntegration_UserRepo_AdminList_Filters(t *testing.T) {
	repo, orgRepo := setupAdminReposDB(t)
	ctx := context.Background()

	orgA := &domain.Organization{Name: "A"}
	orgB := &domain.Organization{Name: "B"}
	require.NoError(t, orgRepo.Create(ctx, orgA))
	require.NoError(t, orgRepo.Create(ctx, orgB))

	seedUser(t, repo, &orgA.ID, "alice", true)
	seedUser(t, repo, &orgA.ID, "bob", false)
	seedUser(t, repo, &orgB.ID, "carol", false)
	deletedUser := seedUser(t, repo, &orgB.ID, "dave", false)
	require.NoError(t, repo.Delete(ctx, deletedUser.ID))

	// Big page size so all seed rows fit on page 1 without clipping.
	bigPage := domain.ListParams{Page: 1, PageSize: 50}

	// Filter by org.
	usrs, total, err := repo.AdminList(ctx, domain.AdminListUsersQuery{
		OrganizationID: &orgA.ID, ListParams: bigPage,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, usrs, 2)

	// Filter by is_admin.
	adminTrue := true
	usrs, total, err = repo.AdminList(ctx, domain.AdminListUsersQuery{
		IsAdmin: &adminTrue, ListParams: bigPage,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "alice", usrs[0].Username)

	// Search substring match (case-insensitive via ILIKE). The handler sets
	// SearchFields from its white-list; simulate that here.
	searchParams := domain.ListParams{
		Page: 1, PageSize: 50,
		Search:       "CAR",
		SearchFields: []string{"username", "name"},
	}
	usrs, _, err = repo.AdminList(ctx, domain.AdminListUsersQuery{ListParams: searchParams})
	require.NoError(t, err)
	assert.Len(t, usrs, 1)
	assert.Equal(t, "carol", usrs[0].Username)

	// IncludeDeleted=false → dave hidden.
	_, total, err = repo.AdminList(ctx, domain.AdminListUsersQuery{ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)

	// IncludeDeleted=true → dave visible.
	_, total, err = repo.AdminList(ctx, domain.AdminListUsersQuery{
		IncludeDeleted: true, ListParams: bigPage,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(4), total)
}

func TestIntegration_UserRepo_CountAll_IncludesDeleted(t *testing.T) {
	repo, _ := setupAdminReposDB(t)
	ctx := context.Background()

	u1 := seedUser(t, repo, nil, "c1", false)
	u2 := seedUser(t, repo, nil, "c2", false)
	require.NoError(t, repo.Delete(ctx, u2.ID))

	n, err := repo.CountAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)
	_ = u1
}

func TestIntegration_OrgRepo_Restore_ClearsDeletedAt(t *testing.T) {
	_, orgRepo := setupAdminReposDB(t)
	ctx := context.Background()

	org := &domain.Organization{Name: "R-Org"}
	require.NoError(t, orgRepo.Create(ctx, org))
	require.NoError(t, orgRepo.Delete(ctx, org.ID))

	// Not visible via scoped find.
	_, err := orgRepo.FindByID(ctx, org.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	require.NoError(t, orgRepo.Restore(ctx, org.ID))

	got, err := orgRepo.FindByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, org.ID, got.ID)
}

func TestIntegration_OrgRepo_Restore_NotFound(t *testing.T) {
	_, orgRepo := setupAdminReposDB(t)
	err := orgRepo.Restore(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_OrgRepo_ListAll_IncludeDeleted(t *testing.T) {
	_, orgRepo := setupAdminReposDB(t)
	ctx := context.Background()

	active := &domain.Organization{Name: "Active"}
	soft := &domain.Organization{Name: "Soft"}
	require.NoError(t, orgRepo.Create(ctx, active))
	require.NoError(t, orgRepo.Create(ctx, soft))
	require.NoError(t, orgRepo.Delete(ctx, soft.ID))

	bigPage := domain.ListParams{Page: 1, PageSize: 50}

	_, total, err := orgRepo.ListAll(ctx, domain.AdminListOrganizationsQuery{ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)

	orgs, total, err := orgRepo.ListAll(ctx, domain.AdminListOrganizationsQuery{IncludeDeleted: true, ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, orgs, 2)
}

func TestIntegration_OrgRepo_CountActiveAndDeleted(t *testing.T) {
	_, orgRepo := setupAdminReposDB(t)
	ctx := context.Background()

	for i, name := range []string{"a", "b", "c"} {
		o := &domain.Organization{Name: name}
		require.NoError(t, orgRepo.Create(ctx, o))
		if i == 2 {
			require.NoError(t, orgRepo.Delete(ctx, o.ID))
		}
	}

	active, err := orgRepo.CountActive(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), active)

	deleted, err := orgRepo.CountDeleted(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)
}
