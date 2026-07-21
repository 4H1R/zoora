package organizations_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
)

// fakeEnqueuer records enqueued tasks so cleanup-scheduling can be asserted.
type fakeEnqueuer struct{ tasks []*asynq.Task }

func (f *fakeEnqueuer) Enqueue(t *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	f.tasks = append(f.tasks, t)
	return &asynq.TaskInfo{}, nil
}

type orgRepoMock struct{ mock.Mock }

func (m *orgRepoMock) Create(ctx context.Context, org *domain.Organization) error {
	return m.Called(ctx, org).Error(0)
}

func (m *orgRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *orgRepoMock) FindBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *orgRepoMock) Update(ctx context.Context, org *domain.Organization) error {
	return m.Called(ctx, org).Error(0)
}

func (m *orgRepoMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *orgRepoMock) List(ctx context.Context, f domain.OrganizationFilter) ([]domain.Organization, int64, error) {
	args := m.Called(ctx, f)
	items, _ := args.Get(0).([]domain.Organization)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *orgRepoMock) GetStats(ctx context.Context) (*domain.OrganizationStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrganizationStats), args.Error(1)
}

func (m *orgRepoMock) AdminList(ctx context.Context, q domain.AdminListOrganizationsQuery) ([]domain.Organization, int64, error) {
	args := m.Called(ctx, q)
	items, _ := args.Get(0).([]domain.Organization)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *orgRepoMock) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *orgRepoMock) Restore(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *orgRepoMock) UpdatePlan(ctx context.Context, id uuid.UUID, plan domain.Plan, expiresAt *time.Time) error {
	return m.Called(ctx, id, plan, expiresAt).Error(0)
}

// noopSettingsRepo satisfies domain.OrganizationSettingsRepository; org-create
// inserts a default settings row, which these tests don't assert on.
type noopSettingsRepo struct{}

func (noopSettingsRepo) Create(ctx context.Context, s *domain.OrganizationSettings) error { return nil }
func (noopSettingsRepo) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	return domain.NewDefaultOrganizationSettings(orgID), nil
}
func (noopSettingsRepo) Update(ctx context.Context, s *domain.OrganizationSettings) error { return nil }

func newOrganizationService(repo *orgRepoMock) domain.OrganizationService {
	return organizations.NewService(repo, nil, noopSettingsRepo{}, nil, nil, slog.Default())
}

func orgCaller(userID uuid.UUID, orgID *uuid.UUID, isAdmin bool) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: userID, OrgID: orgID, IsAdmin: isAdmin})
}

func TestAdminHardDelete_EnqueuesStorageCleanup(t *testing.T) {
	repo := &orgRepoMock{}
	q := &fakeEnqueuer{}
	svc := organizations.NewService(repo, nil, noopSettingsRepo{}, nil, q, slog.Default())

	orgID := uuid.New()
	repo.On("HardDelete", mock.Anything, orgID).Return(nil)

	adminCtx := orgCaller(uuid.New(), nil, true)
	require.NoError(t, svc.AdminHardDelete(adminCtx, orgID))

	require.Len(t, q.tasks, 1)
	task := q.tasks[0]
	assert.Equal(t, domain.TypeOrganizationCleanup, task.Type())
	var payload domain.OrganizationCleanupPayload
	require.NoError(t, json.Unmarshal(task.Payload(), &payload))
	assert.Equal(t, orgID, payload.OrganizationID)
	repo.AssertExpectations(t)
}

func TestAdminHardDelete_RepoError_NoCleanupEnqueued(t *testing.T) {
	repo := &orgRepoMock{}
	q := &fakeEnqueuer{}
	svc := organizations.NewService(repo, nil, noopSettingsRepo{}, nil, q, slog.Default())

	orgID := uuid.New()
	repo.On("HardDelete", mock.Anything, orgID).Return(errors.New("boom"))

	adminCtx := orgCaller(uuid.New(), nil, true)
	assert.Error(t, svc.AdminHardDelete(adminCtx, orgID))
	assert.Empty(t, q.tasks)
}

func orgAdminCtx() context.Context {
	return orgCaller(uuid.New(), nil, true)
}

func TestOrganizationCreateDefaultsStatusAndRequiresCaller(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)

	_, err := svc.Create(context.Background(), domain.CreateOrganizationDTO{Name: "Zoora"})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	ctx := orgCaller(uuid.New(), nil, false)
	repo.On("Create", ctx, mock.MatchedBy(func(org *domain.Organization) bool {
		return org.Name == "Zoora" &&
			org.Slug == "zoora" &&
			org.Description == "Learning" &&
			org.Status == domain.OrganizationStatusActive
	})).Return(nil)

	org, err := svc.Create(ctx, domain.CreateOrganizationDTO{Name: "Zoora", Slug: "zoora", Description: "Learning"})

	assert.NoError(t, err)
	assert.Equal(t, domain.OrganizationStatusActive, org.Status)
	repo.AssertExpectations(t)
}

func TestCreateRejectsReservedSlug(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)
	ctx := orgCaller(uuid.New(), nil, false)

	_, err := svc.Create(ctx, domain.CreateOrganizationDTO{Name: "Acme", Slug: "api"})
	assert.ErrorIs(t, err, domain.ErrInvalidSlug)
	repo.AssertNotCalled(t, "Create")
}

func TestOrganizationGetByIDScopesNonAdminsToTheirOrganization(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)
	userID := uuid.New()
	orgID := uuid.New()
	otherOrgID := uuid.New()

	_, err := svc.GetByID(context.Background(), orgID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	_, err = svc.GetByID(orgCaller(userID, nil, false), orgID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	_, err = svc.GetByID(orgCaller(userID, &otherOrgID, false), orgID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	ctx := orgCaller(userID, &orgID, false)
	repo.On("FindByID", ctx, orgID).Return(&domain.Organization{ID: orgID, Name: "Own org"}, nil).Once()
	org, err := svc.GetByID(ctx, orgID)
	assert.NoError(t, err)
	assert.Equal(t, "Own org", org.Name)

	adminCtx := orgAdminCtx()
	repo.On("FindByID", adminCtx, orgID).Return(&domain.Organization{ID: orgID, Name: "Admin view"}, nil).Once()
	org, err = svc.GetByID(adminCtx, orgID)
	assert.NoError(t, err)
	assert.Equal(t, "Admin view", org.Name)
}

func TestOrganizationUpdateAppliesOnlyProvidedFields(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)
	orgID := uuid.New()
	ctx := orgCaller(uuid.New(), &orgID, false)
	org := &domain.Organization{ID: orgID, Name: "Old", Description: "old", Status: domain.OrganizationStatusTrial}
	newName := "New"

	repo.On("FindByID", ctx, orgID).Return(org, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(updated *domain.Organization) bool {
		return updated.ID == orgID &&
			updated.Name == "New" &&
			updated.Description == "old" &&
			updated.Status == domain.OrganizationStatusTrial
	})).Return(nil)

	updated, err := svc.Update(ctx, orgID, domain.UpdateOrganizationDTO{Name: &newName})

	assert.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, "old", updated.Description)
}

func TestOrganizationListAndStatsRequireCallerAndWrapRepositoryErrors(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)
	filter := domain.OrganizationFilter{Search: "school", Limit: 10}
	repoErr := errors.New("db down")

	_, _, err := svc.List(context.Background(), filter)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	_, err = svc.GetStats(context.Background())
	assert.ErrorIs(t, err, domain.ErrForbidden)

	ctx := orgCaller(uuid.New(), nil, false)
	repo.On("List", ctx, filter).Return([]domain.Organization(nil), int64(0), repoErr)
	_, _, err = svc.List(ctx, filter)
	assert.ErrorIs(t, err, repoErr)

	repo.On("GetStats", ctx).Return(nil, repoErr)
	_, err = svc.GetStats(ctx)
	assert.ErrorIs(t, err, repoErr)
}

func TestOrganizationAdminMethodsRequireAdminAndDefaultPagination(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)
	nonAdmin := orgCaller(uuid.New(), nil, false)
	orgID := uuid.New()

	_, _, err := svc.AdminList(nonAdmin, domain.AdminListOrganizationsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	_, err = svc.AdminCreate(nonAdmin, domain.AdminCreateOrganizationDTO{Name: "Nope"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	_, err = svc.AdminUpdate(nonAdmin, orgID, domain.AdminUpdateOrganizationDTO{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.ErrorIs(t, svc.AdminHardDelete(nonAdmin, orgID), domain.ErrForbidden)
	assert.ErrorIs(t, svc.AdminRestore(nonAdmin, orgID), domain.ErrForbidden)

	ctx := orgAdminCtx()
	repo.On("AdminList", ctx, mock.MatchedBy(func(q domain.AdminListOrganizationsQuery) bool {
		return q.ListParams.Page == 1 && q.ListParams.PageSize == domain.DefaultPageSize
	})).Return([]domain.Organization{}, int64(0), nil)
	_, _, err = svc.AdminList(ctx, domain.AdminListOrganizationsQuery{ListParams: domain.ListParams{Page: -5}})
	assert.NoError(t, err)

	repo.On("Create", ctx, mock.MatchedBy(func(org *domain.Organization) bool {
		return org.Name == "Admin org" && org.Slug == "admin-org" && org.Status == domain.OrganizationStatusActive
	})).Return(nil)
	created, err := svc.AdminCreate(ctx, domain.AdminCreateOrganizationDTO{Name: "Admin org", Slug: "admin-org"})
	assert.NoError(t, err)
	assert.Equal(t, domain.OrganizationStatusActive, created.Status)

	status := domain.OrganizationStatusSuspended
	name := "Suspended org"
	repo.On("FindByID", ctx, orgID).Return(&domain.Organization{ID: orgID, Name: "Old", Status: domain.OrganizationStatusActive}, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(org *domain.Organization) bool {
		return org.Name == name && org.Status == status
	})).Return(nil)
	updated, err := svc.AdminUpdate(ctx, orgID, domain.AdminUpdateOrganizationDTO{Name: &name, Status: &status})
	assert.NoError(t, err)
	assert.Equal(t, status, updated.Status)

	repo.On("HardDelete", ctx, orgID).Return(nil)
	assert.NoError(t, svc.AdminHardDelete(ctx, orgID))
	repo.On("Restore", ctx, orgID).Return(nil)
	assert.NoError(t, svc.AdminRestore(ctx, orgID))
}

func TestSetPlanRequiresAdmin(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)

	orgID := uuid.New()
	nonAdmin := orgCaller(uuid.New(), &orgID, false)
	_, err := svc.SetPlan(nonAdmin, orgID, domain.SetPlanDTO{Plan: domain.PlanKey(domain.TierPro, 50)})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "UpdatePlan", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSetPlanRejectsUnknownPlan(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)

	id := uuid.New()
	_, err := svc.SetPlan(orgAdminCtx(), id, domain.SetPlanDTO{Plan: domain.Plan("bogus")})
	assert.ErrorIs(t, err, domain.ErrValidation)
	repo.AssertNotCalled(t, "UpdatePlan", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSetPlanPersistsAndReturnsOrg(t *testing.T) {
	repo := &orgRepoMock{}
	svc := newOrganizationService(repo)

	id := uuid.New()
	exp := time.Unix(1_800_000_000, 0)
	existing := &domain.Organization{ID: id, Name: "o", Slug: "s", Plan: domain.PlanFree}
	updated := &domain.Organization{ID: id, Name: "o", Slug: "s", Plan: domain.PlanKey(domain.TierPro, 50), PlanExpiresAt: &exp}

	repo.On("FindByID", mock.Anything, id).Return(existing, nil).Once()
	repo.On("UpdatePlan", mock.Anything, id, domain.PlanKey(domain.TierPro, 50), &exp).Return(nil).Once()
	repo.On("FindByID", mock.Anything, id).Return(updated, nil).Once()

	got, err := svc.SetPlan(orgAdminCtx(), id, domain.SetPlanDTO{Plan: domain.PlanKey(domain.TierPro, 50), ExpiresAt: &exp})
	assert.NoError(t, err)
	assert.Equal(t, domain.PlanKey(domain.TierPro, 50), got.Plan)
	repo.AssertExpectations(t)
}
