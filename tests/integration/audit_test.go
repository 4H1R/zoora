//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/audit"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
	"github.com/4H1R/zoora/tests/testutil"
)

func TestAuditRepositoryCreateAndList(t *testing.T) {
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(&domain.Organization{}, &domain.AuditEntry{}))

	repo := audit.NewRepository(db)
	ctx := context.Background()

	org := factory.NewOrganization()
	require.NoError(t, db.Create(org).Error)
	target := uuid.New()

	e := &domain.AuditEntry{
		OrganizationID: org.ID,
		ActorName:      "System",
		Action:         domain.AuditDeleted,
		TargetType:     domain.AuditTargetClass,
		TargetID:       &target,
		TargetLabel:    "Physics 101",
		Outcome:        domain.AuditOutcomeSuccess,
		Metadata:       map[string]any{"cascaded": map[string]any{"enrollments": 3}},
	}
	require.NoError(t, repo.Create(ctx, e))
	require.NotEqual(t, uuid.Nil, e.ID)

	got, total, err := repo.List(ctx, org.ID, domain.AuditListQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 20},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	require.Equal(t, "Physics 101", got[0].TargetLabel)

	// target_type filter narrows correctly.
	tt := domain.AuditTargetUser
	_, total, err = repo.List(ctx, org.ID, domain.AuditListQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 20},
		TargetType: &tt,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
}
