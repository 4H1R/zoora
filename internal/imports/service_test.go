package imports_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"

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

func (m *mockMediaRepo) ListOwnerMedia(context.Context, uuid.UUID) ([]domain.MediaOwner, error) {
	return nil, nil
}

func (m *mockMediaRepo) ListOwnerRecordings(context.Context, uuid.UUID) ([]domain.MediaOwner, error) {
	return nil, nil
}

func (m *mockMediaRepo) ListOwnerFiles(context.Context, uuid.UUID, string, *uuid.UUID, domain.ListParams) ([]domain.OwnerFile, int64, error) {
	return nil, 0, nil
}

type mockObjectStore struct{ mock.Mock }

func (m *mockObjectStore) GetObject(ctx context.Context, key string, maxSize int64) ([]byte, error) {
	args := m.Called(ctx, key, maxSize)
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
func (stubClassRepo) Delete(ctx context.Context, id uuid.UUID) error        { return nil }
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

// mockUserRepo is a real mock (mock.Mock backed) for the two methods
// processUsers exercises — FindByUsernames and Create. The rest of the wide
// domain.UserRepository surface is unused by import processing, so those
// methods return plain zero values like stubUserRepo does.
type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindByIDWithPermissions(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindByUsernameAndOrg(ctx context.Context, username string, orgID uuid.UUID) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindByUsernames(ctx context.Context, orgID uuid.UUID, usernames []string) ([]domain.User, error) {
	args := m.Called(ctx, orgID, usernames)
	items, _ := args.Get(0).([]domain.User)
	return items, args.Error(1)
}

func (m *mockUserRepo) SearchActiveInOrg(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FilterIDsInOrg(ctx context.Context, orgID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (m *mockUserRepo) FindAdminByUsername(ctx context.Context, username string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *mockUserRepo) List(ctx context.Context, scope domain.UserListScope, p domain.ListParams) ([]domain.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) StatusCounts(ctx context.Context, scope domain.UserListScope) (domain.UserStatusCounts, error) {
	return domain.UserStatusCounts{}, nil
}
func (m *mockUserRepo) HardDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) AdminList(ctx context.Context, q domain.AdminListUsersQuery) ([]domain.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) CountAll(ctx context.Context) (int64, error) { return 0, nil }

// mockRoleRepo is a real mock for List — the only method processUsers'
// roleLookup calls. Everything else returns zero values.
type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepo) FindPresetByName(ctx context.Context, name string) (*domain.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) Update(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *mockRoleRepo) List(ctx context.Context, f domain.RoleFilter) ([]domain.Role, error) {
	args := m.Called(ctx, f)
	items, _ := args.Get(0).([]domain.Role)
	return items, args.Error(1)
}

func (m *mockRoleRepo) AdminList(ctx context.Context, f domain.AdminRoleFilter) ([]domain.Role, int64, error) {
	return nil, 0, nil
}

func (m *mockRoleRepo) SetPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return nil
}

func (m *mockRoleRepo) Stats(ctx context.Context, orgID *uuid.UUID) (*domain.RoleStats, error) {
	return nil, nil
}

func (m *mockRoleRepo) GetPermissionNames(ctx context.Context, roleID uuid.UUID) ([]string, error) {
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
		job := &mockJobRepo{}
		// Default: no prior job for the org/type, so Create's
		// one-running-import-per-org guard doesn't block tests that don't
		// care about it.
		job.On("Latest", mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
		deps.job = job
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
	job.On("Latest", mock.Anything, orgID, domain.ImportTypeUsers).Return(nil, domain.ErrNotFound)
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

// TestCreateImport_RunningJobConflicts covers the one-running-import-per-org
// guard: a still-running (pending/processing, not stale) prior job of the
// same type blocks a new Create before media is even looked up.
func TestCreateImport_RunningJobConflicts(t *testing.T) {
	orgID := uuid.New()

	job := &mockJobRepo{}
	job.On("Latest", mock.Anything, orgID, domain.ImportTypeUsers).
		Return(&domain.ImportJob{
			ID: uuid.New(), OrganizationID: orgID, Type: domain.ImportTypeUsers,
			Status: domain.ImportStatusProcessing, UpdatedAt: time.Now(),
		}, nil)

	media := &mockMediaRepo{}
	svc := newTestService(t, testDeps{job: job, media: media})

	_, err := svc.Create(callerCtx(orgID, string(domain.PermUsersCreate)), domain.CreateImportJobDTO{
		Type: domain.ImportTypeUsers, MediaID: uuid.New(),
	})

	assert.ErrorIs(t, err, domain.ErrConflict)
	media.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}

// TestCreateImport_StaleRunningJobDoesNotConflict covers the flip side: a
// pending/processing job whose UpdatedAt is older than staleImportAfter is
// treated as dead (crashed worker), so Create proceeds instead of
// conflicting.
func TestCreateImport_StaleRunningJobDoesNotConflict(t *testing.T) {
	orgID := uuid.New()
	mediaID := uuid.New()

	job := &mockJobRepo{}
	job.On("Latest", mock.Anything, orgID, domain.ImportTypeUsers).
		Return(&domain.ImportJob{
			ID: uuid.New(), OrganizationID: orgID, Type: domain.ImportTypeUsers,
			Status: domain.ImportStatusProcessing, UpdatedAt: time.Now().Add(-20 * time.Minute),
		}, nil)
	job.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		args.Get(1).(*domain.ImportJob).ID = uuid.New()
	}).Return(nil)

	media := &mockMediaRepo{}
	media.On("FindByID", mock.Anything, mediaID).
		Return(&domain.Media{ID: mediaID, OrganizationID: &orgID, FileName: "u.xlsx", Size: 100}, nil)

	queue := &mockEnqueuer{}
	queue.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{}, nil)

	svc := newTestService(t, testDeps{job: job, media: media, queue: queue})

	created, err := svc.Create(callerCtx(orgID, string(domain.PermUsersCreate)), domain.CreateImportJobDTO{
		Type: domain.ImportTypeUsers, MediaID: mediaID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, created)
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

// TestLatest_StaleRunningJobMarkedFailed covers the stuck-job escape hatch:
// a pending/processing job whose UpdatedAt is older than staleImportAfter
// (no clock injection exists, so a fixed 20-minute-old timestamp stands in
// for "worker crashed a while ago") gets marked failed and persisted before
// being returned, so the polling dialog and the one-running-import guard
// don't wedge forever on a dead worker.
func TestLatest_StaleRunningJobMarkedFailed(t *testing.T) {
	orgID := uuid.New()
	jobID := uuid.New()
	staleTime := time.Now().Add(-20 * time.Minute)

	job := &mockJobRepo{}
	job.On("Latest", mock.Anything, orgID, domain.ImportTypeUsers).
		Return(&domain.ImportJob{
			ID: jobID, OrganizationID: orgID, Type: domain.ImportTypeUsers,
			Status: domain.ImportStatusProcessing, UpdatedAt: staleTime,
		}, nil)
	job.On("Update", mock.Anything, mock.MatchedBy(func(j *domain.ImportJob) bool {
		return j.ID == jobID && j.Status == domain.ImportStatusFailed &&
			j.Error != nil && *j.Error == "import timed out"
	})).Return(nil)

	svc := newTestService(t, testDeps{job: job})
	result, err := svc.Latest(callerCtx(orgID, string(domain.PermUsersCreate)), domain.ImportTypeUsers)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domain.ImportStatusFailed, result.Status)
	require.NotNil(t, result.Error)
	assert.Equal(t, "import timed out", *result.Error)
	job.AssertExpectations(t)
}

// TestLatest_FreshRunningJobUnchanged covers the flip side: a recently
// updated pending/processing job is returned exactly as stored, with no
// repo.Update call at all.
func TestLatest_FreshRunningJobUnchanged(t *testing.T) {
	orgID := uuid.New()
	jobID := uuid.New()

	job := &mockJobRepo{}
	job.On("Latest", mock.Anything, orgID, domain.ImportTypeUsers).
		Return(&domain.ImportJob{
			ID: jobID, OrganizationID: orgID, Type: domain.ImportTypeUsers,
			Status: domain.ImportStatusProcessing, UpdatedAt: time.Now(),
		}, nil)

	svc := newTestService(t, testDeps{job: job})
	result, err := svc.Latest(callerCtx(orgID, string(domain.PermUsersCreate)), domain.ImportTypeUsers)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domain.ImportStatusProcessing, result.Status)
	assert.Nil(t, result.Error)
	job.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
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

// --- processUsers test wiring --------------------------------------------

// usersProcessDeps exposes typed mocks for the dependencies processUsers
// exercises, plus a ready-to-use testDeps for newTestService.
type usersProcessDeps struct {
	testDeps testDeps
	repo     *mockJobRepo
	users    *mockUserRepo
	roles    *mockRoleRepo
	media    *mockMediaRepo
	storage  *mockObjectStore
	results  *mockResultStore
	ent      *fakeEnt
}

// newUsersProcessDeps wires the media lookup that ProcessJob always performs
// before dispatching to processUsers, so every test only needs to stub the
// parts relevant to its scenario (storage.GetObject, roles.List, etc).
func newUsersProcessDeps(t *testing.T) usersProcessDeps {
	t.Helper()
	media := &mockMediaRepo{}
	media.On("FindByID", mock.Anything, mock.Anything).
		Return(&domain.Media{ID: uuid.New(), FileName: "users.xlsx"}, nil)

	repo := &mockJobRepo{}
	users := &mockUserRepo{}
	roles := &mockRoleRepo{}
	storage := &mockObjectStore{}
	results := &mockResultStore{}
	ent := &fakeEnt{}

	return usersProcessDeps{
		repo: repo, users: users, roles: roles, media: media,
		storage: storage, results: results, ent: ent,
		testDeps: testDeps{
			job: repo, users: users, roles: roles, media: media,
			ent: ent, storage: storage, results: results,
		},
	}
}

func usersJobAndPayload(orgID uuid.UUID, perms []string) (*domain.ImportJob, domain.ImportProcessPayload) {
	jobID, userID, mediaID := uuid.New(), uuid.New(), uuid.New()
	job := &domain.ImportJob{
		ID: jobID, OrganizationID: orgID, UserID: userID,
		MediaID: mediaID, Type: domain.ImportTypeUsers, Status: domain.ImportStatusPending,
	}
	payload := domain.ImportProcessPayload{
		JobID: jobID, UserID: userID, OrgID: orgID,
		Permissions: perms, Plan: domain.PlanFree,
	}
	return job, payload
}

// --- processClasses test wiring ------------------------------------------

// mockClassRepo is a real mock for the two methods processClasses exercises
// -- ListByNames and Create. The rest of the wide domain.ClassRepository
// surface is unused by import processing, so those methods return plain
// zero values like stubClassRepo does.
type mockClassRepo struct{ mock.Mock }

func (m *mockClassRepo) Create(ctx context.Context, class *domain.Class) error {
	return m.Called(ctx, class).Error(0)
}

func (m *mockClassRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	return nil, nil
}
func (m *mockClassRepo) Update(ctx context.Context, class *domain.Class) error { return nil }
func (m *mockClassRepo) Delete(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *mockClassRepo) List(ctx context.Context, scope domain.ClassListScope, p domain.ListParams) ([]domain.Class, int64, error) {
	return nil, 0, nil
}

func (m *mockClassRepo) ListByNames(ctx context.Context, orgID uuid.UUID, names []string) ([]domain.Class, error) {
	args := m.Called(ctx, orgID, names)
	items, _ := args.Get(0).([]domain.Class)
	return items, args.Error(1)
}
func (m *mockClassRepo) HardDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockClassRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	return nil, nil
}

func (m *mockClassRepo) AdminList(ctx context.Context, q domain.AdminListClassesQuery) ([]domain.Class, int64, error) {
	return nil, 0, nil
}

// mockClassMemberRepo is a real mock for the two methods processClasses
// exercises -- Create and CountByClass. The rest of the wide
// domain.ClassMemberRepository surface is unused by import processing, so
// those methods return plain zero values like stubClassMemberRepo does.
type mockClassMemberRepo struct{ mock.Mock }

func (m *mockClassMemberRepo) Create(ctx context.Context, cm *domain.ClassMember) error {
	return m.Called(ctx, cm).Error(0)
}

func (m *mockClassMemberRepo) Delete(ctx context.Context, classID, userID uuid.UUID) error {
	return nil
}

func (m *mockClassMemberRepo) Exists(ctx context.Context, classID, userID uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockClassMemberRepo) CountByClass(ctx context.Context, classID uuid.UUID) (int64, error) {
	args := m.Called(ctx, classID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockClassMemberRepo) ListByClass(ctx context.Context, classID uuid.UUID, p domain.ListParams) ([]domain.ClassMember, int64, error) {
	return nil, 0, nil
}

func (m *mockClassMemberRepo) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.ClassMember, error) {
	return nil, nil
}

// classesProcessDeps exposes typed mocks for the dependencies processClasses
// exercises, plus a ready-to-use testDeps for newTestService.
type classesProcessDeps struct {
	testDeps testDeps
	repo     *mockJobRepo
	users    *mockUserRepo
	classes  *mockClassRepo
	members  *mockClassMemberRepo
	media    *mockMediaRepo
	storage  *mockObjectStore
	results  *mockResultStore
	ent      *fakeEnt
}

// newClassesProcessDeps wires the media lookup that ProcessJob always
// performs before dispatching to processClasses, so every test only needs
// to stub the parts relevant to its scenario (storage.GetObject,
// classes.ListByNames, etc).
func newClassesProcessDeps(t *testing.T) classesProcessDeps {
	t.Helper()
	media := &mockMediaRepo{}
	media.On("FindByID", mock.Anything, mock.Anything).
		Return(&domain.Media{ID: uuid.New(), FileName: "classes.xlsx"}, nil)

	repo := &mockJobRepo{}
	users := &mockUserRepo{}
	classes := &mockClassRepo{}
	members := &mockClassMemberRepo{}
	storage := &mockObjectStore{}
	results := &mockResultStore{}
	ent := &fakeEnt{}

	return classesProcessDeps{
		repo: repo, users: users, classes: classes, members: members, media: media,
		storage: storage, results: results, ent: ent,
		testDeps: testDeps{
			job: repo, users: users, classes: classes, members: members, media: media,
			ent: ent, storage: storage, results: results,
		},
	}
}

func classesJobAndPayload(orgID uuid.UUID, perms []string) (*domain.ImportJob, domain.ImportProcessPayload) {
	jobID, userID, mediaID := uuid.New(), uuid.New(), uuid.New()
	job := &domain.ImportJob{
		ID: jobID, OrganizationID: orgID, UserID: userID,
		MediaID: mediaID, Type: domain.ImportTypeClasses, Status: domain.ImportStatusPending,
	}
	payload := domain.ImportProcessPayload{
		JobID: jobID, UserID: userID, OrgID: orgID,
		Permissions: perms, Plan: domain.PlanFree,
	}
	return job, payload
}

func TestProcessUsers_MixedOutcomes(t *testing.T) {
	orgID := uuid.New()
	studentRole := domain.Role{ID: uuid.New(), Name: "Student", IsPreset: true}

	file := usersXLSX(t, [][]string{
		{"name", "username", "password", "role"},
		{"New User", "new.user", "", "-"},                  // created, generated password
		{"Exists", "already.there", "", "-"},               // skipped
		{"Short", "short.pass", "1234567", "-"},            // error: password < 8
		{"Role User", "role.user", "password9", "student"}, // created with role
		{"No Role Cell", "no.role", "password9", ""},       // error: empty role
	})

	deps := newUsersProcessDeps(t) // wires: media.FindByID → media row, storage.GetObject → file
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.roles.On("List", mock.Anything, mock.Anything).Return([]domain.Role{studentRole}, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).
		Return([]domain.User{{Username: "already.there"}}, nil)
	deps.users.On("Create", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate), string(domain.PermRolesUpdate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 5, job.TotalRows)
	assert.Equal(t, 2, job.CreatedCount) // new.user + role.user
	assert.Equal(t, 1, job.SkippedCount)
	assert.Equal(t, 2, job.FailedCount)
	deps.users.AssertNumberOfCalls(t, "Create", 2)

	// created users carry bcrypt hashes, role.user got the Student role
	for _, call := range deps.users.Calls {
		if call.Method != "Create" {
			continue
		}
		u := call.Arguments.Get(1).(*domain.User)
		assert.True(t, strings.HasPrefix(u.Password, "$2"), "password must be a bcrypt hash")
		if u.Username == "role.user" {
			require.NotNil(t, u.RoleID)
			assert.Equal(t, studentRole.ID, *u.RoleID)
		}
	}
}

// TestProcessUsers_SeatLimitFailsWholeJob covers behavior spec bullet 6: the
// seat limit is checked against the whole batch before any row is created,
// and a failure fails the entire job rather than creating a partial batch.
func TestProcessUsers_SeatLimitFailsWholeJob(t *testing.T) {
	orgID := uuid.New()

	file := usersXLSX(t, [][]string{
		{"name", "username", "password", "role"},
		{"New User", "new.user", "", "-"},
		{"Other User", "other.user", "", "-"},
	})

	deps := newUsersProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.roles.On("List", mock.Anything, mock.Anything).Return([]domain.Role{}, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).Return([]domain.User{}, nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.ent.userLimitNErr = errors.New("seat limit exceeded")

	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusFailed, job.Status)
	require.NotNil(t, job.Error)
	assert.Contains(t, *job.Error, "seat limit")
	assert.Equal(t, 0, job.CreatedCount)

	deps.users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	deps.results.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
}

// TestProcessUsers_ManagerRoleRejectedForNonAdmin covers behavior spec bullet
// 5's second clause: even a non-admin caller who does hold roles:update
// cannot assign the preset Manager role via import -- only that row errors,
// the rest of the batch still creates.
func TestProcessUsers_ManagerRoleRejectedForNonAdmin(t *testing.T) {
	orgID := uuid.New()
	managerRole := domain.Role{ID: uuid.New(), Name: domain.PresetRoleManager, IsPreset: true}

	file := usersXLSX(t, [][]string{
		{"name", "username", "password", "role"},
		{"Mgr User", "mgr.user", "password9", "manager"},
		{"Plain User", "plain.user", "password9", "-"},
	})

	deps := newUsersProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.roles.On("List", mock.Anything, mock.Anything).Return([]domain.Role{managerRole}, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).Return([]domain.User{}, nil)
	deps.users.On("Create", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// non-admin caller WITH roles:update -- still must not be able to grant Manager.
	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate), string(domain.PermRolesUpdate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 1, job.CreatedCount) // plain.user only
	assert.Equal(t, 1, job.FailedCount)  // mgr.user rejected
	deps.users.AssertNumberOfCalls(t, "Create", 1)

	for _, call := range deps.users.Calls {
		if call.Method != "Create" {
			continue
		}
		u := call.Arguments.Get(1).(*domain.User)
		assert.Equal(t, "plain.user", u.Username)
	}
}

// TestProcessUsers_NamedRoleRejectedWithoutPermission covers behavior spec
// bullet 5's first clause: a non-admin caller lacking roles:update cannot
// assign any named role, but "-" rows and the rest of the batch still go
// through.
func TestProcessUsers_NamedRoleRejectedWithoutPermission(t *testing.T) {
	orgID := uuid.New()
	studentRole := domain.Role{ID: uuid.New(), Name: "Student", IsPreset: true}

	file := usersXLSX(t, [][]string{
		{"name", "username", "password", "role"},
		{"Role User", "role.user", "password9", "student"},
		{"Plain User", "plain.user", "password9", "-"},
	})

	deps := newUsersProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.roles.On("List", mock.Anything, mock.Anything).Return([]domain.Role{studentRole}, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).Return([]domain.User{}, nil)
	deps.users.On("Create", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// no roles:update permission at all
	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 1, job.CreatedCount) // plain.user only
	assert.Equal(t, 1, job.FailedCount)  // role.user rejected
	deps.users.AssertNumberOfCalls(t, "Create", 1)

	for _, call := range deps.users.Calls {
		if call.Method != "Create" {
			continue
		}
		u := call.Arguments.Get(1).(*domain.User)
		assert.Equal(t, "plain.user", u.Username)
	}
}

// TestProcessUsers_ConflictRaceOnCreateSkipsRow covers behavior spec bullet
// 3's race-condition clause: FindByUsernames misses a username that another
// concurrent request creates first, and users.Create returns ErrConflict --
// the row is skipped, not errored, and the job still completes.
func TestProcessUsers_ConflictRaceOnCreateSkipsRow(t *testing.T) {
	orgID := uuid.New()

	file := usersXLSX(t, [][]string{
		{"name", "username", "password", "role"},
		{"Race User", "race.user", "password9", "-"},
	})

	deps := newUsersProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.roles.On("List", mock.Anything, mock.Anything).Return([]domain.Role{}, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).Return([]domain.User{}, nil) // precheck misses it
	deps.users.On("Create", mock.Anything, mock.Anything).Return(domain.ErrConflict)                   // create-time race
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 0, job.CreatedCount)
	assert.Equal(t, 1, job.SkippedCount)
	assert.Equal(t, 0, job.FailedCount)
}

// TestProcessUsers_UnparseableFileFailsJob covers behavior spec bullet 1: a
// file that can't be parsed at all fails the whole job with a reason and
// creates nobody. The >5000-row case shares the exact same
// "ParseUsersFile err != nil -> s.fail" branch (see MaxRows in parser.go),
// so this single assertion exercises that shared code path too.
func TestProcessUsers_UnparseableFileFailsJob(t *testing.T) {
	orgID := uuid.New()

	deps := newUsersProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return([]byte("junk"), nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusFailed, job.Status)
	require.NotNil(t, job.Error)
	assert.Equal(t, 0, job.CreatedCount)
	deps.users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestProcessUsers_ProgressUpdatesEvery25Creates covers behavior spec bullet
// 8's progress clause: UpdateProgress fires every 25 successful creates, and
// the final Update carries the completed status and full counters.
func TestProcessUsers_ProgressUpdatesEvery25Creates(t *testing.T) {
	orgID := uuid.New()

	rows := [][]string{{"name", "username", "password", "role"}}
	for i := 0; i < 30; i++ {
		rows = append(rows, []string{fmt.Sprintf("User %d", i), fmt.Sprintf("user%d", i), "password9", "-"})
	}
	file := usersXLSX(t, rows)

	deps := newUsersProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.roles.On("List", mock.Anything, mock.Anything).Return([]domain.Role{}, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).Return([]domain.User{}, nil)
	deps.users.On("Create", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 30, job.CreatedCount)
	assert.Equal(t, 30, job.ProcessedRows)
	// fires once at the 25th create, not again before job completion at 30
	deps.repo.AssertCalled(t, "UpdateProgress", mock.Anything, job.ID, 25, 25, 0, 0)
	deps.repo.AssertNumberOfCalls(t, "UpdateProgress", 1)
}

// TestProcessUsers_RowValidationErrors pins the row-error branches inside
// processUsers that TestProcessUsers_MixedOutcomes only exercises indirectly
// via lumped failed-count assertions: bad name, bad username format, unknown
// role, ambiguous role name, and duplicate username in file. It also proves
// username normalization (lowercased+trimmed before validation) by feeding a
// mixed-case, padded username that only becomes valid once normalized.
func TestProcessUsers_RowValidationErrors(t *testing.T) {
	orgID := uuid.New()
	// Two roles differing only in case/ID collapse onto the same lookup key
	// ("student"), so resolveRole reports "ambiguous role name" instead of
	// picking one.
	studentA := domain.Role{ID: uuid.New(), Name: "Student", IsPreset: true}
	studentB := domain.Role{ID: uuid.New(), Name: "STUDENT", IsPreset: false}

	file := usersXLSX(t, [][]string{
		{"name", "username", "password", "role"},
		{" A ", "badname.user", "password9", "-"},                     // row 2: name is 1 rune after trim
		{"Valid Name", "ab", "password9", "-"},                        // row 3: username too short (<3)
		{"Role Unknown", "unknown.role", "password9", "doesnotexist"}, // row 4: unknown role
		{"Role Ambig", "ambig.role", "password9", "student"},          // row 5: ambiguous role
		{"Dup One", "dup.user", "password9", "-"},                     // row 6: created
		{"Dup Two", "DUP.USER", "password9", "-"},                     // row 7: duplicate of row 6 after normalization
		{"Ali R", "  Ali.R  ", "password9", "-"},                      // row 8: normalizes to "ali.r", created
	})

	deps := newUsersProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.roles.On("List", mock.Anything, mock.Anything).Return([]domain.Role{studentA, studentB}, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).Return([]domain.User{}, nil)

	var createdUsers []*domain.User
	deps.users.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { createdUsers = append(createdUsers, args.Get(1).(*domain.User)) }).
		Return(nil)

	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	var resultBytes []byte
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { resultBytes = args.Get(2).([]byte) }).
		Return(nil)

	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate), string(domain.PermRolesUpdate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 7, job.TotalRows)
	assert.Equal(t, 2, job.CreatedCount) // dup.user + ali.r
	assert.Equal(t, 0, job.SkippedCount)
	assert.Equal(t, 5, job.FailedCount) // bad name, bad username, unknown role, ambiguous role, duplicate

	// exact created-user set, distinguishing the two survivors from the five
	// row errors, and pinning the normalized username's exact form.
	require.Len(t, createdUsers, 2)
	createdUsernames := make(map[string]bool, len(createdUsers))
	for _, u := range createdUsers {
		createdUsernames[u.Username] = true
	}
	assert.True(t, createdUsernames["dup.user"])
	assert.True(t, createdUsernames["ali.r"], "mixed-case padded username must normalize to exactly \"ali.r\"")

	// per-row messages, read back from the captured result file — the same
	// pattern TestProcessUsers_GeneratedPasswordOnlyInResultFile uses.
	require.NotNil(t, resultBytes)
	f, err := excelize.OpenReader(bytes.NewReader(resultBytes))
	require.NoError(t, err)
	defer f.Close()
	rows2, err := f.GetRows(f.GetSheetList()[0])
	require.NoError(t, err)
	require.Len(t, rows2, 8) // header + 7 data rows

	// columns: name(0) username(1) role(2) status(3) message(4) generated_password(5)
	assert.Equal(t, "error", rows2[1][3])
	assert.Equal(t, "name must be at least 2 characters", rows2[1][4])

	assert.Equal(t, "error", rows2[2][3])
	assert.Equal(t, "username must be 3-30 chars: lowercase letters, digits, dot or underscore", rows2[2][4])

	assert.Equal(t, "error", rows2[3][3])
	assert.Equal(t, "unknown role", rows2[3][4])

	assert.Equal(t, "error", rows2[4][3])
	assert.Equal(t, "ambiguous role name", rows2[4][4])

	assert.Equal(t, "created", rows2[5][3])

	assert.Equal(t, "error", rows2[6][3])
	assert.Equal(t, "duplicate username in file", rows2[6][4])

	assert.Equal(t, "created", rows2[7][3])
	assert.Equal(t, "ali.r", rows2[7][1], "result file must reflect the normalized username")
}

// TestProcessUsers_GeneratedPasswordOnlyInResultFile covers behavior spec
// bullet 7: an empty password cell gets a generated 10-char password that
// appears in the result file but never on the stored user -- only its
// bcrypt hash does.
func TestProcessUsers_GeneratedPasswordOnlyInResultFile(t *testing.T) {
	orgID := uuid.New()

	file := usersXLSX(t, [][]string{
		{"name", "username", "password", "role"},
		{"Gen User", "gen.user", "", "-"},
	})

	deps := newUsersProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.roles.On("List", mock.Anything, mock.Anything).Return([]domain.Role{}, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).Return([]domain.User{}, nil)

	var createdUser *domain.User
	deps.users.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { createdUser = args.Get(1).(*domain.User) }).
		Return(nil)

	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	var resultBytes []byte
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { resultBytes = args.Get(2).([]byte) }).
		Return(nil)

	job, payload := usersJobAndPayload(orgID, []string{string(domain.PermUsersCreate)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	require.NotNil(t, createdUser)
	assert.True(t, strings.HasPrefix(createdUser.Password, "$2"), "stored password must be a bcrypt hash")

	require.NotNil(t, resultBytes)
	f, err := excelize.OpenReader(bytes.NewReader(resultBytes))
	require.NoError(t, err)
	defer f.Close()
	rows2, err := f.GetRows(f.GetSheetList()[0])
	require.NoError(t, err)
	require.Len(t, rows2, 2) // header + 1 data row
	generated := rows2[1][5] // generated_password column
	assert.Len(t, generated, 10)
	// the generated plaintext is only ever in the result file -- the stored
	// user carries its bcrypt hash, never the plaintext itself.
	assert.NotEqual(t, createdUser.Password, generated)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(createdUser.Password), []byte(generated)))
}

// --- processClasses tests -------------------------------------------------

func TestProcessClasses_MixedOutcomes(t *testing.T) {
	orgID := uuid.New()
	ali, sara := uuid.New(), uuid.New()
	existingClass := domain.Class{ID: uuid.New(), OrganizationID: orgID, UserID: ali, Name: "Old-Class", TotalUsers: 0}

	file := classesXLSX(t,
		[][]string{
			{"class_name", "owner_username", "description", "capacity"},
			{"Math-A", "ali.r", "Algebra", "30"},   // created
			{"Old-Class", "ali.r", "ignored", "5"}, // reused: skipped, desc/cap ignored
			{"Ghost", "who.dis", "", ""},           // error: owner not found
			{"Math-A", "ali.r", "", ""},            // error: duplicate in file
		},
		[][]string{
			{"class_name", "member_username"},
			{"Math-A", "sara.k"},    // enrolled
			{"Old-Class", "sara.k"}, // enrolled into reused class
			{"Ghost", "sara.k"},     // error: class row failed
			{"Math-A", "no.body"},   // error: member not found
		})

	deps := newClassesProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).
		Return([]domain.User{{ID: ali, Username: "ali.r"}, {ID: sara, Username: "sara.k"}}, nil)
	deps.classes.On("ListByNames", mock.Anything, orgID, mock.Anything).
		Return([]domain.Class{existingClass}, nil)
	deps.classes.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		args.Get(1).(*domain.Class).ID = uuid.New() // simulate DB id assignment
	}).Return(nil)
	deps.members.On("Create", mock.Anything, mock.Anything).Return(nil)
	// Math-A is created with capacity 30 (from the sheet), so its single
	// enrollment (sara.k) lazily loads the current member count once.
	// Old-Class is reused with TotalUsers 0 (unlimited), so its enrollment
	// never triggers a capacity check.
	deps.members.On("CountByClass", mock.Anything, mock.Anything).Return(int64(0), nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	job, payload := classesJobAndPayload(orgID, []string{string(domain.PermClassesCreateAny)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 8, job.TotalRows)
	assert.Equal(t, 3, job.CreatedCount) // Math-A class + 2 enrollments
	assert.Equal(t, 1, job.SkippedCount) // Old-Class reuse
	assert.Equal(t, 4, job.FailedCount)
	deps.classes.AssertNumberOfCalls(t, "Create", 1)

	// reused class untouched: only ONE class create, and the created one is Math-A with sheet values
	c := deps.classes.Calls[len(deps.classes.Calls)-1].Arguments.Get(1).(*domain.Class)
	assert.Equal(t, "Math-A", c.Name)
	assert.Equal(t, ali, c.UserID)
	assert.Equal(t, 30, c.TotalUsers)
}

// TestProcessClasses_AmbiguousExistingName covers behavior spec bullet 3's
// ">1 match" clause: when a class name already resolves to more than one
// existing class, the class row errors as ambiguous and its member rows
// error too since the class never resolved to a usable ID.
func TestProcessClasses_AmbiguousExistingName(t *testing.T) {
	orgID := uuid.New()
	ali, sara := uuid.New(), uuid.New()
	dupA := domain.Class{ID: uuid.New(), OrganizationID: orgID, UserID: ali, Name: "Math-A", TotalUsers: 0}
	dupB := domain.Class{ID: uuid.New(), OrganizationID: orgID, UserID: ali, Name: "Math-A", TotalUsers: 0}

	file := classesXLSX(t,
		[][]string{
			{"class_name", "owner_username", "description", "capacity"},
			{"Math-A", "ali.r", "Algebra", "30"}, // error: ambiguous
		},
		[][]string{
			{"class_name", "member_username"},
			{"Math-A", "sara.k"}, // error: class row failed (ambiguous)
		})

	deps := newClassesProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).
		Return([]domain.User{{ID: ali, Username: "ali.r"}, {ID: sara, Username: "sara.k"}}, nil)
	deps.classes.On("ListByNames", mock.Anything, orgID, mock.Anything).
		Return([]domain.Class{dupA, dupB}, nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	job, payload := classesJobAndPayload(orgID, []string{string(domain.PermClassesCreateAny)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 2, job.TotalRows)
	assert.Equal(t, 0, job.CreatedCount)
	assert.Equal(t, 0, job.SkippedCount)
	assert.Equal(t, 2, job.FailedCount)
	deps.classes.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	deps.members.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

	rows := classesResultRows(t, resultBytesFromSet(t, deps.results))
	// Classes sheet: header + 1 data row
	assert.Equal(t, "error", rows.classes[1][4])
	assert.Equal(t, "multiple classes with this name already exist", rows.classes[1][5])
	// Members sheet: header + 1 data row
	assert.Equal(t, "error", rows.members[1][2])
	assert.Equal(t, "class not found in Classes sheet (or its row failed)", rows.members[1][3])
}

// TestProcessClasses_CapacityFull covers behavior spec bullet 4's capacity
// clause: a reused class with a positive TotalUsers that's already at
// capacity (per CountByClass) rejects further member rows with "class is
// full", without ever calling members.Create.
func TestProcessClasses_CapacityFull(t *testing.T) {
	orgID := uuid.New()
	ali, sara := uuid.New(), uuid.New()
	existingClass := domain.Class{ID: uuid.New(), OrganizationID: orgID, UserID: ali, Name: "Full-Class", TotalUsers: 1}

	file := classesXLSX(t,
		[][]string{
			{"class_name", "owner_username", "description", "capacity"},
			{"Full-Class", "ali.r", "ignored", "5"}, // reused: skipped, sheet capacity ignored
		},
		[][]string{
			{"class_name", "member_username"},
			{"Full-Class", "sara.k"}, // error: class is full
		})

	deps := newClassesProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).
		Return([]domain.User{{ID: ali, Username: "ali.r"}, {ID: sara, Username: "sara.k"}}, nil)
	deps.classes.On("ListByNames", mock.Anything, orgID, mock.Anything).
		Return([]domain.Class{existingClass}, nil)
	deps.members.On("CountByClass", mock.Anything, existingClass.ID).Return(int64(1), nil)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	job, payload := classesJobAndPayload(orgID, []string{string(domain.PermClassesCreateAny)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 2, job.TotalRows)
	assert.Equal(t, 0, job.CreatedCount)
	assert.Equal(t, 1, job.SkippedCount) // reused class row
	assert.Equal(t, 1, job.FailedCount)  // capacity-full member row
	deps.members.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	deps.members.AssertNumberOfCalls(t, "CountByClass", 1)

	rows := classesResultRows(t, resultBytesFromSet(t, deps.results))
	assert.Equal(t, "error", rows.members[1][2])
	assert.Equal(t, "class is full", rows.members[1][3])
}

// TestProcessClasses_DuplicateMemberConflictSkips covers behavior spec
// bullet 4's conflict clause: members.Create returning domain.ErrConflict
// (an already-existing membership) is treated as a skip, not a failure.
func TestProcessClasses_DuplicateMemberConflictSkips(t *testing.T) {
	orgID := uuid.New()
	ali, sara := uuid.New(), uuid.New()

	file := classesXLSX(t,
		[][]string{
			{"class_name", "owner_username", "description", "capacity"},
			{"Math-A", "ali.r", "Algebra", ""}, // created
		},
		[][]string{
			{"class_name", "member_username"},
			{"Math-A", "sara.k"}, // already a member -> conflict -> skipped
		})

	deps := newClassesProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).
		Return([]domain.User{{ID: ali, Username: "ali.r"}, {ID: sara, Username: "sara.k"}}, nil)
	deps.classes.On("ListByNames", mock.Anything, orgID, mock.Anything).Return([]domain.Class{}, nil)
	deps.classes.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		args.Get(1).(*domain.Class).ID = uuid.New()
	}).Return(nil)
	deps.members.On("Create", mock.Anything, mock.Anything).Return(domain.ErrConflict)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	job, payload := classesJobAndPayload(orgID, []string{string(domain.PermClassesCreateAny)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 2, job.TotalRows)
	assert.Equal(t, 1, job.CreatedCount) // Math-A class created
	assert.Equal(t, 1, job.SkippedCount) // duplicate member conflict
	assert.Equal(t, 0, job.FailedCount)

	rows := classesResultRows(t, resultBytesFromSet(t, deps.results))
	assert.Equal(t, "skipped", rows.members[1][2])
	assert.Equal(t, "already a member", rows.members[1][3])
}

// --- processClassMembers tests --------------------------------------------

func TestCreateImport_ClassMembersTypeNeedsUpdateAny(t *testing.T) {
	orgID := uuid.New()
	svc := newTestService(t, testDeps{})
	_, err := svc.Create(callerCtx(orgID, string(domain.PermClassesCreateAny), string(domain.PermClassesUpdate)), domain.CreateImportJobDTO{
		Type: domain.ImportTypeClassMembers, MediaID: uuid.New(),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func classMembersJobAndPayload(orgID uuid.UUID, perms []string) (*domain.ImportJob, domain.ImportProcessPayload) {
	jobID, userID, mediaID := uuid.New(), uuid.New(), uuid.New()
	job := &domain.ImportJob{
		ID: jobID, OrganizationID: orgID, UserID: userID,
		MediaID: mediaID, Type: domain.ImportTypeClassMembers, Status: domain.ImportStatusPending,
	}
	payload := domain.ImportProcessPayload{
		JobID: jobID, UserID: userID, OrgID: orgID,
		Permissions: perms, Plan: domain.PlanFree,
	}
	return job, payload
}

// TestProcessClassMembers_MixedOutcomes covers the members-only import: rows
// enroll into existing classes resolved by name; unknown class, ambiguous
// class name, unknown member, capacity-full and duplicate-membership rows
// each get their own outcome without stopping the batch. Classes are never
// created by this import type.
func TestProcessClassMembers_MixedOutcomes(t *testing.T) {
	orgID := uuid.New()
	ali, sara := uuid.New(), uuid.New()
	mathA := domain.Class{ID: uuid.New(), OrganizationID: orgID, UserID: ali, Name: "Math-A", TotalUsers: 0}
	fullClass := domain.Class{ID: uuid.New(), OrganizationID: orgID, UserID: ali, Name: "Full-Class", TotalUsers: 1}
	dupA := domain.Class{ID: uuid.New(), OrganizationID: orgID, UserID: ali, Name: "Dup", TotalUsers: 0}
	dupB := domain.Class{ID: uuid.New(), OrganizationID: orgID, UserID: ali, Name: "Dup", TotalUsers: 0}

	file := membersXLSX(t, [][]string{
		{"class_name", "member_username"},
		{"Math-A", "SARA.K "},    // enrolled (username normalized)
		{"Math-A", "ali.r"},      // already a member -> conflict -> skipped
		{"Ghost", "sara.k"},      // error: class not found
		{"Dup", "sara.k"},        // error: ambiguous class name
		{"Math-A", "no.body"},    // error: member not found
		{"Full-Class", "sara.k"}, // error: class is full
		{"", "sara.k"},           // error: class_name required
	})

	deps := newClassesProcessDeps(t)
	deps.storage.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(file, nil)
	deps.users.On("FindByUsernames", mock.Anything, orgID, mock.Anything).
		Return([]domain.User{{ID: ali, Username: "ali.r"}, {ID: sara, Username: "sara.k"}}, nil)
	deps.classes.On("ListByNames", mock.Anything, orgID, mock.Anything).
		Return([]domain.Class{mathA, fullClass, dupA, dupB}, nil)
	deps.members.On("CountByClass", mock.Anything, fullClass.ID).Return(int64(1), nil)
	deps.members.On("Create", mock.Anything, mock.MatchedBy(func(cm *domain.ClassMember) bool {
		return cm.ClassID == mathA.ID && cm.UserID == sara
	})).Return(nil)
	deps.members.On("Create", mock.Anything, mock.MatchedBy(func(cm *domain.ClassMember) bool {
		return cm.ClassID == mathA.ID && cm.UserID == ali
	})).Return(domain.ErrConflict)
	deps.repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.repo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	deps.results.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	job, payload := classMembersJobAndPayload(orgID, []string{string(domain.PermClassesUpdateAny)})
	deps.repo.On("FindByID", mock.Anything, job.ID).Return(job, nil)

	svc := newTestService(t, deps.testDeps)
	require.NoError(t, svc.ProcessJob(context.Background(), payload))

	assert.Equal(t, domain.ImportStatusCompleted, job.Status)
	assert.Equal(t, 7, job.TotalRows)
	assert.Equal(t, 1, job.CreatedCount) // sara.k into Math-A
	assert.Equal(t, 1, job.SkippedCount) // ali.r already a member
	assert.Equal(t, 5, job.FailedCount)
	deps.classes.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	deps.members.AssertNumberOfCalls(t, "Create", 2)

	// per-row messages from the single-sheet result file
	data := resultBytesFromSet(t, deps.results)
	require.NotNil(t, data)
	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer f.Close()
	rows, err := f.GetRows("Members")
	require.NoError(t, err)
	require.Len(t, rows, 8) // header + 7 data rows

	// columns: class_name(0) member_username(1) status(2) message(3)
	assert.Equal(t, "created", rows[1][2])
	assert.Equal(t, "sara.k", rows[1][1], "result file must reflect the normalized username")
	assert.Equal(t, "skipped", rows[2][2])
	assert.Equal(t, "already a member", rows[2][3])
	assert.Equal(t, "error", rows[3][2])
	assert.Equal(t, "class not found", rows[3][3])
	assert.Equal(t, "error", rows[4][2])
	assert.Equal(t, "multiple classes with this name exist", rows[4][3])
	assert.Equal(t, "error", rows[5][2])
	assert.Equal(t, "member username not found", rows[5][3])
	assert.Equal(t, "error", rows[6][2])
	assert.Equal(t, "class is full", rows[6][3])
	assert.Equal(t, "error", rows[7][2])
	assert.Equal(t, "class_name and member_username are required", rows[7][3])
}

// --- shared result-file readback helpers for processClasses tests --------

// resultBytesFromSet extracts the []byte payload from the mockResultStore's
// captured Set call so tests can assert on per-row messages in the result
// xlsx, the same way the processUsers tests do.
func resultBytesFromSet(t *testing.T, results *mockResultStore) []byte {
	t.Helper()
	for _, call := range results.Calls {
		if call.Method == "Set" {
			return call.Arguments.Get(2).([]byte)
		}
	}
	return nil
}

type classesResultSheets struct {
	classes []([]string)
	members []([]string)
}

func classesResultRows(t *testing.T, data []byte) classesResultSheets {
	t.Helper()
	require.NotNil(t, data)
	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer f.Close()
	classRows, err := f.GetRows("Classes")
	require.NoError(t, err)
	memberRows, err := f.GetRows("Members")
	require.NoError(t, err)
	return classesResultSheets{classes: classRows, members: memberRows}
}
