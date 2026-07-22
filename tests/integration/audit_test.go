//go:build integration

package integration

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/audit"
	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/users"
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

func discardLoggerIT() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// failingAuditRepo wraps a real AuditRepository but forces Create to error,
// simulating an audit insert failure that occurs inside the caller's
// transaction. List delegates to the inner repo so reads still work.
type failingAuditRepo struct {
	inner domain.AuditRepository
}

func (f failingAuditRepo) Create(_ context.Context, _ *domain.AuditEntry) error {
	return errors.New("forced audit insert failure")
}

func (f failingAuditRepo) List(ctx context.Context, orgID uuid.UUID, q domain.AuditListQuery) ([]domain.AuditEntry, int64, error) {
	return f.inner.List(ctx, orgID, q)
}

// TestAuditEndToEndSameTxRollback drives the REAL classes service against a
// testcontainer Postgres to prove the audit log's same-transaction hard-fail
// guarantee end-to-end:
//
//  1. Deleting a class writes exactly one deleted/class entry labelled with the
//     class name.
//  2. If the audit insert is forced to fail, the class delete is rolled back —
//     the class still exists and no entry is written (same-tx hard-fail).
//  3. A Manager (audit:view_any) can List the entry; a Student cannot
//     (ErrForbidden).
func TestAuditEndToEndSameTxRollback(t *testing.T) {
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.ClassMember{},
		&domain.AuditEntry{},
	))

	log := discardLoggerIT()
	classRepo := classes.NewRepository(db)
	sessionRepo := classes.NewSessionRepository(db)
	memberRepo := classes.NewMemberRepository(db)
	userRepo := users.NewRepository(db)
	orgRepo := organizations.NewRepository(db)
	auditRepo := audit.NewRepository(db)

	transactor := database.NewTransactor(db)
	auditService := audit.NewService(auditRepo, log)
	classService := classes.NewService(classRepo, sessionRepo, memberRepo, nil, transactor, auditService, log)

	ctx := context.Background()
	org := seedOrg(t, orgRepo, "Acme")
	teacher := seedTeacher(t, userRepo, org.ID, "teacher")
	manager := seedTeacher(t, userRepo, org.ID, "manager")

	managerCtx := domain.WithCaller(ctx, domain.Caller{
		UserID:      manager.ID,
		OrgID:       &org.ID,
		Name:        manager.Name,
		Username:    manager.Username,
		Permissions: []string{string(domain.PermClassesDeleteAny), string(domain.PermAuditViewAny)},
	})

	// --- 1. Successful delete writes exactly one deleted/class entry. ---
	physics := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Physics 101"}
	require.NoError(t, classRepo.Create(ctx, physics))

	require.NoError(t, classService.Delete(managerCtx, physics.ID))

	entries, total, err := auditRepo.List(ctx, org.ID, domain.AuditListQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, entries, 1)
	require.Equal(t, domain.AuditDeleted, entries[0].Action)
	require.Equal(t, domain.AuditTargetClass, entries[0].TargetType)
	require.Equal(t, "Physics 101", entries[0].TargetLabel)
	require.NotNil(t, entries[0].TargetID)
	require.Equal(t, physics.ID, *entries[0].TargetID)
	// Actor snapshot came from the Caller.
	require.NotNil(t, entries[0].ActorID)
	require.Equal(t, manager.ID, *entries[0].ActorID)

	// The class is actually gone.
	_, err = classRepo.FindByID(ctx, physics.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)

	// --- 2. Forced audit failure rolls the delete back (same-tx hard-fail). ---
	chem := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Chemistry 201"}
	require.NoError(t, classRepo.Create(ctx, chem))

	failingSvc := audit.NewService(failingAuditRepo{inner: auditRepo}, log)
	failingClassService := classes.NewService(classRepo, sessionRepo, memberRepo, nil, transactor, failingSvc, log)

	err = failingClassService.Delete(managerCtx, chem.ID)
	require.Error(t, err) // the forced audit insert failure surfaces

	// The class must still exist — the delete was rolled back with the audit insert.
	got, err := classRepo.FindByID(ctx, chem.ID)
	require.NoError(t, err)
	require.Equal(t, "Chemistry 201", got.Name)

	// No audit entry was written for the rolled-back delete.
	chemTarget := chem.ID
	_, chemTotal, err := auditRepo.List(ctx, org.ID, domain.AuditListQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
		TargetID:   &chemTarget,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), chemTotal)

	// Total org entries unchanged (still just the Physics deletion).
	_, total, err = auditRepo.List(ctx, org.ID, domain.AuditListQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)

	// --- 3. Manager can List; Student is forbidden. ---
	listed, total, err := auditService.List(managerCtx, domain.AuditListQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, listed, 1)
	require.Equal(t, "Physics 101", listed[0].TargetLabel)

	studentCtx := domain.WithCaller(ctx, domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &org.ID,
		Permissions: []string{}, // no audit:view_any
	})
	_, _, err = auditService.List(studentCtx, domain.AuditListQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.ErrorIs(t, err, domain.ErrForbidden)
}
