package customfields_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/customfields"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/tests/testutil"
)

func setupRepo(t *testing.T) (domain.CustomFieldRepository, uuid.UUID, *gorm.DB) {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.UserCustomFieldDefinition{},
	))
	org := &domain.Organization{Name: "Acme", Slug: "acme-" + uuid.NewString()[:8], Status: domain.OrganizationStatusActive}
	require.NoError(t, db.Create(org).Error)
	return customfields.NewRepository(db), org.ID, db
}

func TestCreateAndListDefinitions(t *testing.T) {
	repo, orgID, _ := setupRepo(t)
	ctx := context.Background()

	def := &domain.UserCustomFieldDefinition{
		OrganizationID: orgID, Label: "Student ID", FieldType: domain.CustomFieldTypeText,
	}
	require.NoError(t, repo.CreateDefinition(ctx, def))
	require.NotEqual(t, uuid.Nil, def.ID)

	list, err := repo.ListDefinitions(ctx, orgID, false)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Student ID", list[0].Label)

	count, err := repo.CountActiveDefinitions(ctx, orgID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestUserValuesRoundTripAndUniqueness(t *testing.T) {
	repo, orgID, db := setupRepo(t)
	ctx := context.Background()
	fieldID := uuid.New()

	u1 := &domain.User{OrganizationID: &orgID, Username: "u1", Name: "U1", Password: "x"}
	require.NoError(t, db.Create(u1).Error)

	require.NoError(t, repo.SetUserCustomFields(ctx, u1.ID, map[string]any{fieldID.String(): "12345"}))

	got, gotOrg, err := repo.GetUserCustomFields(ctx, u1.ID)
	require.NoError(t, err)
	require.Equal(t, "12345", got[fieldID.String()])
	require.Equal(t, orgID, gotOrg)

	n, err := repo.CountUsersWithFieldValue(ctx, orgID, fieldID, "12345", uuid.Nil)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	nExcl, err := repo.CountUsersWithFieldValue(ctx, orgID, fieldID, "12345", u1.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), nExcl)
}

func TestHasDuplicateFieldValues(t *testing.T) {
	repo, orgID, db := setupRepo(t)
	ctx := context.Background()
	fieldID := uuid.New()

	u1 := &domain.User{OrganizationID: &orgID, Username: "u1", Name: "U1", Password: "x"}
	u2 := &domain.User{OrganizationID: &orgID, Username: "u2", Name: "U2", Password: "x"}
	require.NoError(t, db.Create(u1).Error)
	require.NoError(t, db.Create(u2).Error)

	require.NoError(t, repo.SetUserCustomFields(ctx, u1.ID, map[string]any{fieldID.String(): "dup"}))
	dup, err := repo.HasDuplicateFieldValues(ctx, orgID, fieldID)
	require.NoError(t, err)
	require.False(t, dup)

	require.NoError(t, repo.SetUserCustomFields(ctx, u2.ID, map[string]any{fieldID.String(): "dup"}))
	dup, err = repo.HasDuplicateFieldValues(ctx, orgID, fieldID)
	require.NoError(t, err)
	require.True(t, dup)
}
