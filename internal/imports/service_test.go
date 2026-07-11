package imports_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
	"github.com/4H1R/zoora/internal/imports"
)

// --- mocks exercised by these tests -----------------------------------

type mockJobRepo struct{ mock.Mock }

func (m *mockJobRepo) Create(ctx context.Context, job *domain.ImportJob) error {
	return m.Called(ctx, job).Error(0)
}

func (m *mockJobRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ImportJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ImportJob), args.Error(1)
}

func (m *mockJobRepo) Update(ctx context.Context, job *domain.ImportJob) error {
	return m.Called(ctx, job).Error(0)
}

func (m *mockJobRepo) UpdateProgress(ctx context.Context, id uuid.UUID, processed, created, skipped, failed int) error {
	return m.Called(ctx, id, processed, created, skipped, failed).Error(0)
}

func (m *mockJobRepo) Latest(ctx context.Context, orgID uuid.UUID, t domain.ImportType) (*domain.ImportJob, error) {
	args := m.Called(ctx, orgID, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ImportJob), args.Error(1)
}

type mockMediaRepo struct{ mock.Mock }

func (m *mockMediaRepo) Create(ctx context.Context, media *domain.Media) error {
	return m.Called(ctx, media).Error(0)
}

func (m *mockMediaRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Media), args.Error(1)
}

func (m *mockMediaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockMediaRepo) ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]domain.Media, error) {
	args := m.Called(ctx, modelType, modelID, collection)
	items, _ := args.Get(0).([]domain.Media)
	return items, args.Error(1)
}

func (m *mockMediaRepo) ListFolders(ctx context.Context, orgID uuid.UUID) ([]domain.MediaFolder, error) {
	args := m.Called(ctx, orgID)
	items, _ := args.Get(0).([]domain.MediaFolder)
	return items, args.Error(1)
}

func (m *mockMediaRepo) ListFiles(ctx context.Context, orgID uuid.UUID, modelType string, p domain.ListParams) ([]domain.Media, int64, error) {
	args := m.Called(ctx, orgID, modelType, p)
	items, _ := args.Get(0).([]domain.Media)
	return items, args.Get(1).(int64), args.Error(2)
}

type mockObjectStore struct{ mock.Mock }

func (m *mockObjectStore) GetObject(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	b, _ := args.Get(0).([]byte)
	return b, args.Error(1)
}

type mockEnqueuer struct{ mock.Mock }

func (m *mockEnqueuer) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	args := m.Called(task, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}

type mockResultStore struct{ mock.Mock }

func (m *mockResultStore) Set(ctx context.Context, jobID uuid.UUID, data []byte) error {
	return m.Called(ctx, jobID, data).Error(0)
}

func (m *mockResultStore) Get(ctx context.Context, jobID uuid.UUID) ([]byte, error) {
	args := m.Called(ctx, jobID)
	b, _ := args.Get(0).([]byte)
	return b, args.Error(1)
}

// --- stub repos: not exercised by Create/Get/Latest/Result, so plain
// zero-value returns (no mock.Mock bookkeeping needed) satisfy the wide
// domain interfaces the service constructor requires. ------------------

type stubUserRepo struct{}

func (stubUserRepo) Create(ctx context.Context, user *domain.User) error { return nil }
func (stubUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (stubUserRepo) FindByIDWithPermissions(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (stubUserRepo) FindByUsernameAndOrg(ctx context.Context, username string, orgID uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (stubUserRepo) FindByUsernames(ctx context.Context, orgID uuid.UUID, usernames []string) ([]domain.User, error) {
	return nil, nil
}
func (stubUserRepo) SearchActiveInOrg(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]domain.User, error) {
	return nil, nil
}
func (stubUserRepo) FilterIDsInOrg(ctx context.Context, orgID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (stubUserRepo) FindAdminByUsername(ctx context.Context, username string) (*domain.User, error) {
	return nil, nil
}
func (stubUserRepo) Update(ctx context.Context, user *domain.User) error { return nil }
func (stubUserRepo) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (stubUserRepo) List(ctx context.Context, scope domain.UserListScope, p domain.ListParams) ([]domain.User, int64, error) {
	return nil, 0, nil
}
func (stubUserRepo) StatusCounts(ctx context.Context, scope domain.UserListScope) (domain.UserStatusCounts, error) {
	return domain.UserStatusCounts{}, nil
}
func (stubUserRepo) HardDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (stubUserRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (stubUserRepo) AdminList(ctx context.Context, q domain.AdminListUsersQuery) ([]domain.User, int64, error) {
	return nil, 0, nil
}
func (stubUserRepo) CountAll(ctx context.Context) (int64, error) { return 0, nil }

type stubRoleRepo struct{}

func (stubRoleRepo) Create(ctx context.Context, role *domain.Role) error { return nil }
func (stubRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return nil, nil
}
func (stubRoleRepo) FindPresetByName(ctx context.Context, name string) (*domain.Role, error) {
	return nil, nil
}
func (stubRoleRepo) Update(ctx context.Context, role *domain.Role) error { return nil }
func (stubRoleRepo) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (stubRoleRepo) List(ctx context.Context, f domain.RoleFilter) ([]domain.Role, error) {
	return nil, nil
}
func (stubRoleRepo) AdminList(ctx context.Context, f domain.AdminRoleFilter) ([]domain.Role, int64, error) {
	return nil, 0, nil
}
func (stubRoleRepo) SetPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return nil
}
func (stubRoleRepo) Stats(ctx context.Context, orgID *uuid.UUID) (*domain.RoleStats, error) {
	return nil, nil
}
func (stubRoleRepo) GetPermissionNames(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	return nil, nil
}

type stubClassRepo struct{}

func (stubClassRepo) Create(ctx context.Context, class *domain.Class) error { return nil }
func (stubClassRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	return nil, nil
}
func (stubClassRepo) Update(ctx context.Context, class *domain.Class) error { return nil }
func (stubClassRepo) Delete(ctx context.Context, id uuid.UUID) error       { return nil }
func (stubClassRepo) List(ctx context.Context, scope domain.ClassListScope, p domain.ListParams) ([]domain.Class, int64, error) {
	return nil, 0, nil
}
func (stubClassRepo) ListByNames(ctx context.Context, orgID uuid.UUID, names []string) ([]domain.Class, error) {
	return nil, nil
}
func (stubClassRepo) HardDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (stubClassRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	return nil, nil
}
func (stubClassRepo) AdminList(ctx context.Context, q domain.AdminListClassesQuery) ([]domain.Class, int64, error) {
	return nil, 0, nil
}

type stubClassMemberRepo struct{}

func (stubClassMemberRepo) Create(ctx context.Context, m *domain.ClassMember) error { return nil }
func (stubClassMemberRepo) Delete(ctx context.Context, classID, userID uuid.UUID) error {
	return nil
}
func (stubClassMemberRepo) Exists(ctx context.Context, classID, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (stubClassMemberRepo) CountByClass(ctx context.Context, classID uuid.UUID) (int64, error) {
	return 0, nil
}
func (stubClassMemberRepo) ListByClass(ctx context.Context, classID uuid.UUID, p domain.ListParams) ([]domain.ClassMember, int64, error) {
	return nil, 0, nil
}
func (stubClassMemberRepo) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.ClassMember, error) {
	return nil, nil
}

// fakeEnt is an entitlements.Service stub with configurable errors. None of
// the Create/Get/Latest/Result paths call it today (it's wired in for Tasks
// 8-9's row processing), so the zero value (all nil errors) is the default.
type fakeEnt struct {
	userLimitErr  error
	userLimitNErr error
	storageErr    error
	roomsErr      error
}

func (f *fakeEnt) CheckUserLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements) error {
	return f.userLimitErr
}

func (f *fakeEnt) CheckUserLimitN(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements, n int64) error {
	return f.userLimitNErr
}

func (f *fakeEnt) CheckStorageLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements, addBytes int64) error {
	return f.storageErr
}

func (f *fakeEnt) CheckConcurrentRoomsLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements) error {
	return f.roomsErr
}

var _ entitlements.Service = (*fakeEnt)(nil)

// --- test wiring --------------------------------------------------------

type testDeps struct {
	job     *mockJobRepo
	users   domain.UserRepository
	roles   domain.RoleRepository
	classes domain.ClassRepository
	members domain.ClassMemberRepository
	media   *mockMediaRepo
	ent     entitlements.Service
	storage *mockObjectStore
	queue   *mockEnqueuer
	results *mockResultStore
}

func newTestService(t *testing.T, deps testDeps) domain.ImportService {
	t.Helper()
	if deps.job == nil {
		deps.job = &mockJobRepo{}
	}
	if deps.users == nil {
		deps.users = stubUserRepo{}
	}
	if deps.roles == nil {
		deps.roles = stubRoleRepo{}
	}
	if deps.classes == nil {
		deps.classes = stubClassRepo{}
	}
	if deps.members == nil {
		deps.members = stubClassMemberRepo{}
	}
	if deps.media == nil {
		deps.media = &mockMediaRepo{}
	}
	if deps.ent == nil {
		deps.ent = &fakeEnt{}
	}
	if deps.storage == nil {
		deps.storage = &mockObjectStore{}
	}
	if deps.queue == nil {
		deps.queue = &mockEnqueuer{}
	}
	if deps.results == nil {
		deps.results = &mockResultStore{}
	}
	return imports.NewService(
		deps.job, deps.users, deps.roles, deps.classes, deps.members,
		deps.media, deps.ent, deps.storage, deps.queue, deps.results,
		slog.Default(),
	)
}

func callerCtx(orgID uuid.UUID, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgID,
		Permissions: perms,
		Ent:         domain.PlanCatalog[domain.PlanFree],
	})
}

// --- tests ---------------------------------------------------------------

func TestCreateImport_Forbidden(t *testing.T) {
	orgID := uuid.New()
	svc := newTestService(t, testDeps{})
	_, err := svc.Create(callerCtx(orgID /* no perms */), domain.CreateImportJobDTO{
		Type: domain.ImportTypeUsers, MediaID: uuid.New(),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateImport_ClassTypeNeedsCreateAny(t *testing.T) {
	orgID := uuid.New()
	svc := newTestService(t, testDeps{})
	_, err := svc.Create(callerCtx(orgID, string(domain.PermClassesCreate)), domain.CreateImportJobDTO{
		Type: domain.ImportTypeClasses, MediaID: uuid.New(),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateImport_RejectsWrongOrgMedia(t *testing.T) {
	orgID, otherOrg := uuid.New(), uuid.New()
	mediaID := uuid.New()
	media := &mockMediaRepo{}
	media.On("FindByID", mock.Anything, mediaID).
		Return(&domain.Media{ID: mediaID, OrganizationID: &otherOrg, FileName: "u.xlsx", Size: 100}, nil)
	svc := newTestService(t, testDeps{media: media})
	_, err := svc.Create(callerCtx(orgID, string(domain.PermUsersCreate)), domain.CreateImportJobDTO{
		Type: domain.ImportTypeUsers, MediaID: mediaID,
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCreateImport_RejectsOversizeAndWrongExtension(t *testing.T) {
	orgID := uuid.New()

	t.Run("oversize", func(t *testing.T) {
		mediaID := uuid.New()
		media := &mockMediaRepo{}
		media.On("FindByID", mock.Anything, mediaID).
			Return(&domain.Media{ID: mediaID, OrganizationID: &orgID, FileName: "u.xlsx", Size: 11 << 20}, nil)
		svc := newTestService(t, testDeps{media: media})

		_, err := svc.Create(callerCtx(orgID, string(domain.PermUsersCreate)), domain.CreateImportJobDTO{
			Type: domain.ImportTypeUsers, MediaID: mediaID,
		})

		var ve *domain.ValidationError
		assert.ErrorAs(t, err, &ve)
		assert.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("wrong extension", func(t *testing.T) {
		mediaID := uuid.New()
		media := &mockMediaRepo{}
		media.On("FindByID", mock.Anything, mediaID).
			Return(&domain.Media{ID: mediaID, OrganizationID: &orgID, FileName: "u.csv", Size: 100}, nil)
		svc := newTestService(t, testDeps{media: media})

		_, err := svc.Create(callerCtx(orgID, string(domain.PermUsersCreate)), domain.CreateImportJobDTO{
			Type: domain.ImportTypeUsers, MediaID: mediaID,
		})

		var ve *domain.ValidationError
		assert.ErrorAs(t, err, &ve)
		assert.ErrorIs(t, err, domain.ErrValidation)
	})
}

func TestCreateImport_CreatesJobAndEnqueues(t *testing.T) {
	orgID := uuid.New()
	mediaID := uuid.New()
	userID := uuid.New()

	media := &mockMediaRepo{}
	media.On("FindByID", mock.Anything, mediaID).
		Return(&domain.Media{ID: mediaID, OrganizationID: &orgID, FileName: "u.xlsx", Size: 1 << 20}, nil)

	job := &mockJobRepo{}
	job.On("Create", mock.Anything, mock.MatchedBy(func(j *domain.ImportJob) bool {
		return j.OrganizationID == orgID &&
			j.UserID == userID &&
			j.MediaID == mediaID &&
			j.Type == domain.ImportTypeUsers &&
			j.Status == domain.ImportStatusPending
	})).Run(func(args mock.Arguments) {
		j := args.Get(1).(*domain.ImportJob)
		j.ID = uuid.New()
	}).Return(nil)

	var capturedTask *asynq.Task
	queue := &mockEnqueuer{}
	queue.On("Enqueue", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			capturedTask = args.Get(0).(*asynq.Task)
		}).
		Return(&asynq.TaskInfo{}, nil)

	svc := newTestService(t, testDeps{media: media, job: job, queue: queue})

	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		OrgID:       &orgID,
		Permissions: []string{string(domain.PermUsersCreate)},
		Ent:         domain.PlanCatalog[domain.PlanFree],
	})

	created, err := svc.Create(ctx, domain.CreateImportJobDTO{Type: domain.ImportTypeUsers, MediaID: mediaID})
	assert.NoError(t, err)
	if !assert.NotNil(t, created) {
		return
	}
	assert.Equal(t, domain.ImportStatusPending, created.Status)

	job.AssertExpectations(t)
	if !assert.NotNil(t, capturedTask) {
		return
	}
	assert.Equal(t, domain.TypeImportProcess, capturedTask.Type())

	var payload domain.ImportProcessPayload
	assert.NoError(t, json.Unmarshal(capturedTask.Payload(), &payload))
	assert.Equal(t, created.ID, payload.JobID)
	assert.Equal(t, userID, payload.UserID)
	assert.Equal(t, orgID, payload.OrgID)
	assert.False(t, payload.IsAdmin)
	assert.Equal(t, []string{string(domain.PermUsersCreate)}, payload.Permissions)
	assert.Equal(t, domain.PlanFree, payload.Plan)
}

func TestGetImport_WrongOrg(t *testing.T) {
	orgID, otherOrg := uuid.New(), uuid.New()
	jobID := uuid.New()

	job := &mockJobRepo{}
	job.On("FindByID", mock.Anything, jobID).
		Return(&domain.ImportJob{ID: jobID, OrganizationID: otherOrg, Type: domain.ImportTypeUsers}, nil)

	svc := newTestService(t, testDeps{job: job})
	_, err := svc.Get(callerCtx(orgID, string(domain.PermUsersCreate)), jobID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestLatest_NoJobReturnsNilNil(t *testing.T) {
	orgID := uuid.New()

	job := &mockJobRepo{}
	job.On("Latest", mock.Anything, orgID, domain.ImportTypeUsers).
		Return(nil, domain.ErrNotFound)

	svc := newTestService(t, testDeps{job: job})
	result, err := svc.Latest(callerCtx(orgID, string(domain.PermUsersCreate)), domain.ImportTypeUsers)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestResult_ChecksPermThenReadsStore(t *testing.T) {
	orgID := uuid.New()
	jobID := uuid.New()

	t.Run("forbidden without permission never reads the store", func(t *testing.T) {
		job := &mockJobRepo{}
		job.On("FindByID", mock.Anything, jobID).
			Return(&domain.ImportJob{ID: jobID, OrganizationID: orgID, Type: domain.ImportTypeUsers}, nil)
		results := &mockResultStore{}

		svc := newTestService(t, testDeps{job: job, results: results})
		_, err := svc.Result(callerCtx(orgID /* no perms */), jobID)

		assert.ErrorIs(t, err, domain.ErrForbidden)
		results.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	})

	t.Run("ok path reads store bytes", func(t *testing.T) {
		want := []byte("xlsx-bytes")
		job := &mockJobRepo{}
		job.On("FindByID", mock.Anything, jobID).
			Return(&domain.ImportJob{ID: jobID, OrganizationID: orgID, Type: domain.ImportTypeUsers}, nil)
		results := &mockResultStore{}
		results.On("Get", mock.Anything, jobID).Return(want, nil)

		svc := newTestService(t, testDeps{job: job, results: results})
		got, err := svc.Result(callerCtx(orgID, string(domain.PermUsersCreate)), jobID)

		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
