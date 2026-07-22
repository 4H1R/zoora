package orgsettings_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/orgsettings"
)

type mockSettingsRepo struct{ mock.Mock }

func (m *mockSettingsRepo) Create(ctx context.Context, s *domain.OrganizationSettings) error {
	return m.Called(ctx, s).Error(0)
}

func (m *mockSettingsRepo) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrganizationSettings), args.Error(1)
}

func (m *mockSettingsRepo) Update(ctx context.Context, s *domain.OrganizationSettings) error {
	return m.Called(ctx, s).Error(0)
}

// fakeTransactor runs fn inline with no real DB — unit tests exercise the audit
// same-tx wiring without a database.
type fakeTransactor struct{}

func (fakeTransactor) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// auditSpy captures the records a service emits so tests can assert on them.
type auditSpy struct{ records []domain.AuditRecord }

func (a *auditSpy) Record(_ context.Context, r domain.AuditRecord) error {
	a.records = append(a.records, r)
	return nil
}

func (a *auditSpy) RecordDenied(_ context.Context, _ domain.AuditRecord) error { return nil }

func newService(repo domain.OrganizationSettingsRepository) domain.OrganizationSettingsService {
	return orgsettings.NewService(repo, fakeTransactor{}, &auditSpy{}, slog.Default())
}

func newServiceWithAudit(repo domain.OrganizationSettingsRepository, audit domain.AuditRecorder) domain.OrganizationSettingsService {
	return orgsettings.NewService(repo, fakeTransactor{}, audit, slog.Default())
}

func TestGet_RejectsOtherOrg(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()

	repo := &mockSettingsRepo{}
	svc := newService(repo)
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgA})

	_, err := svc.Get(ctx, orgB)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "FindByOrgID")
}

func TestGet_AllowsOwnOrg(t *testing.T) {
	orgA := uuid.New()

	repo := &mockSettingsRepo{}
	repo.On("FindByOrgID", mock.Anything, orgA).Return(&domain.OrganizationSettings{OrganizationID: orgA}, nil)
	svc := newService(repo)
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgA})

	settings, err := svc.Get(ctx, orgA)
	assert.NoError(t, err)
	assert.Equal(t, orgA, settings.OrganizationID)
}

func TestGet_RejectsNoCaller(t *testing.T) {
	repo := &mockSettingsRepo{}
	svc := newService(repo)

	_, err := svc.Get(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdate_RejectsOtherOrg(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()

	repo := &mockSettingsRepo{}
	svc := newService(repo)
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgA})

	_, err := svc.Update(ctx, orgB, domain.UpdateOrganizationSettingsDTO{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "FindByOrgID")
}

func TestUpdate_AllowsOwnOrg(t *testing.T) {
	orgA := uuid.New()

	repo := &mockSettingsRepo{}
	repo.On("FindByOrgID", mock.Anything, orgA).Return(&domain.OrganizationSettings{OrganizationID: orgA}, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.OrganizationSettings")).Return(nil)
	svc := newService(repo)
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgA})

	settings, err := svc.Update(ctx, orgA, domain.UpdateOrganizationSettingsDTO{})
	assert.NoError(t, err)
	assert.Equal(t, orgA, settings.OrganizationID)
}

// GetByOrgID is the internal provider port; it must return settings even when
// the ctx carries no caller (regression: must NOT 403).
func TestGetByOrgID_ProviderNotGuarded(t *testing.T) {
	orgA := uuid.New()

	repo := &mockSettingsRepo{}
	repo.On("FindByOrgID", mock.Anything, orgA).Return(&domain.OrganizationSettings{OrganizationID: orgA}, nil)

	var provider domain.OrganizationSettingsProvider = orgsettings.NewService(repo, fakeTransactor{}, &auditSpy{}, slog.Default())

	settings, err := provider.GetByOrgID(context.Background(), orgA)
	assert.NoError(t, err)
	assert.Equal(t, orgA, settings.OrganizationID)
}
