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
	job := &domain.ImportJob{ID: jobID, OrganizationID: orgID, UserID: userID,
		MediaID: mediaID, Type: domain.ImportTypeUsers, Status: domain.ImportStatusPending}
	payload := domain.ImportProcessPayload{JobID: jobID, UserID: userID, OrgID: orgID,
		Permissions: perms, Plan: domain.PlanFree}
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
	deps.storage.On("GetObject", mock.Anything, mock.Anything).Return(file, nil)
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
	deps.storage.On("GetObject", mock.Anything, mock.Anything).Return(file, nil)
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
	deps.storage.On("GetObject", mock.Anything, mock.Anything).Return(file, nil)
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
	deps.storage.On("GetObject", mock.Anything, mock.Anything).Return(file, nil)
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
	deps.storage.On("GetObject", mock.Anything, mock.Anything).Return(file, nil)
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
	deps.storage.On("GetObject", mock.Anything, mock.Anything).Return([]byte("junk"), nil)
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
	deps.storage.On("GetObject", mock.Anything, mock.Anything).Return(file, nil)
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
	deps.storage.On("GetObject", mock.Anything, mock.Anything).Return(file, nil)
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
