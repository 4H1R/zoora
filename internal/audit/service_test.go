package audit

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
)

// fakeRepo captures the entry passed to Create and serves canned List results.
type fakeRepo struct {
	created *domain.AuditEntry
}

func (f *fakeRepo) Create(_ context.Context, e *domain.AuditEntry) error { f.created = e; return nil }
func (f *fakeRepo) List(_ context.Context, _ uuid.UUID, _ domain.AuditListQuery) ([]domain.AuditEntry, int64, error) {
	return nil, 0, nil
}

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestRecordDerivesActorOrgAndIP(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, discardLogger())

	orgID := uuid.New()
	actorID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID: actorID, OrgID: &orgID, Name: "Ali", Username: "ali",
	})
	ctx = domain.WithRequestInfo(ctx, domain.RequestInfo{IP: "1.2.3.4", UserAgent: "curl"})

	target := uuid.New()
	err := svc.Record(ctx, domain.AuditRecord{
		Action: domain.AuditDeleted, TargetType: domain.AuditTargetClass,
		TargetID: &target, TargetLabel: "Physics 101",
	})
	require.NoError(t, err)
	require.NotNil(t, repo.created)
	require.Equal(t, orgID, repo.created.OrganizationID)
	require.Equal(t, &actorID, repo.created.ActorID)
	require.Equal(t, "Ali", repo.created.ActorName)
	require.Equal(t, domain.AuditOutcomeSuccess, repo.created.Outcome)
	require.Equal(t, "1.2.3.4", repo.created.Metadata["ip"])
	require.Equal(t, "curl", repo.created.Metadata["user_agent"])
}

func TestRecordSystemActorWhenNoCaller(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, discardLogger())

	orgID := uuid.New()
	// No caller in ctx (worker path); OrgID passed explicitly on the record.
	err := svc.Record(context.Background(), domain.AuditRecord{
		Action: domain.AuditUpdated, TargetType: domain.AuditTargetBilling,
		OrgID: &orgID, TargetLabel: "plan downgrade",
	})
	require.NoError(t, err)
	require.Nil(t, repo.created.ActorID)
	require.Equal(t, domain.AuditActorSystemName, repo.created.ActorName)
	require.Equal(t, orgID, repo.created.OrganizationID)
}

func TestRecordErrorsWhenNoOrgResolvable(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, discardLogger())
	// No caller, no OrgID override → cannot file the entry.
	err := svc.Record(context.Background(), domain.AuditRecord{
		Action: domain.AuditCreated, TargetType: domain.AuditTargetClass,
	})
	require.Error(t, err)
	require.Nil(t, repo.created)
}
