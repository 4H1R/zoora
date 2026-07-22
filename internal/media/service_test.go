package media_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/media"
)

type storageMock struct{ mock.Mock }

func (m *storageMock) GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	a := m.Called(ctx, key, expiry)
	return a.String(0), a.Error(1)
}

func (m *storageMock) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	a := m.Called(ctx, key, expiry)
	return a.String(0), a.Error(1)
}

func (m *storageMock) DeleteObject(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

type mediaRepoMock struct{ mock.Mock }

func (m *mediaRepoMock) Create(ctx context.Context, item *domain.Media) error {
	return m.Called(ctx, item).Error(0)
}

func (m *mediaRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Media), args.Error(1)
}

func (m *mediaRepoMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mediaRepoMock) ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]domain.Media, error) {
	args := m.Called(ctx, modelType, modelID, collection)
	items, _ := args.Get(0).([]domain.Media)
	return items, args.Error(1)
}

func (m *mediaRepoMock) ListFolders(ctx context.Context, orgID uuid.UUID) ([]domain.MediaFolder, error) {
	args := m.Called(ctx, orgID)
	folders, _ := args.Get(0).([]domain.MediaFolder)
	return folders, args.Error(1)
}

func (m *mediaRepoMock) ListFiles(ctx context.Context, orgID uuid.UUID, modelType string, p domain.ListParams) ([]domain.Media, int64, error) {
	args := m.Called(ctx, orgID, modelType, p)
	items, _ := args.Get(0).([]domain.Media)
	total, _ := args.Get(1).(int64)
	return items, total, args.Error(2)
}

func (m *mediaRepoMock) ListOwnerMedia(ctx context.Context, orgID uuid.UUID) ([]domain.MediaOwner, error) {
	args := m.Called(ctx, orgID)
	owners, _ := args.Get(0).([]domain.MediaOwner)
	return owners, args.Error(1)
}

func (m *mediaRepoMock) ListOwnerRecordings(ctx context.Context, orgID uuid.UUID) ([]domain.MediaOwner, error) {
	args := m.Called(ctx, orgID)
	owners, _ := args.Get(0).([]domain.MediaOwner)
	return owners, args.Error(1)
}

func (m *mediaRepoMock) ListOwnerFiles(ctx context.Context, orgID uuid.UUID, kind string, ownerID *uuid.UUID, p domain.ListParams) ([]domain.OwnerFile, int64, error) {
	args := m.Called(ctx, orgID, kind, ownerID, p)
	files, _ := args.Get(0).([]domain.OwnerFile)
	total, _ := args.Get(1).(int64)
	return files, total, args.Error(2)
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

func newMediaService(repo *mediaRepoMock, store *storageMock) domain.MediaService {
	return media.NewService(repo, store, nil, nil, fakeTransactor{}, &auditSpy{}, slog.Default())
}

func newMediaServiceAudit(repo *mediaRepoMock, store *storageMock) (domain.MediaService, *auditSpy) {
	spy := &auditSpy{}
	return media.NewService(repo, store, nil, nil, fakeTransactor{}, spy, slog.Default()), spy
}

func mediaCtx(isAdmin bool, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: isAdmin, Permissions: perms})
}

// fakeEntService injects a canned CheckStorageLimit result for quota tests.
type fakeEntService struct{ storageErr error }

func (f fakeEntService) CheckUserLimit(context.Context, uuid.UUID, domain.Entitlements) error {
	return nil
}

func (f fakeEntService) CheckUserLimitN(context.Context, uuid.UUID, domain.Entitlements, int64) error {
	return nil
}

func (f fakeEntService) CheckStorageLimit(context.Context, uuid.UUID, domain.Entitlements, int64) error {
	return f.storageErr
}

func (f fakeEntService) CheckConcurrentRoomsLimit(context.Context, uuid.UUID, domain.Entitlements) error {
	return nil
}

func TestMediaGetAndListRequireCaller(t *testing.T) {
	repo := &mediaRepoMock{}
	svc := newMediaService(repo, &storageMock{})
	mediaID := uuid.New()
	modelID := uuid.New()

	_, err := svc.GetByID(context.Background(), mediaID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	_, err = svc.ListByModel(context.Background(), "practice", modelID, "attachments")
	assert.ErrorIs(t, err, domain.ErrForbidden)

	ctx := mediaCtx(false)
	repo.On("FindByID", ctx, mediaID).Return(&domain.Media{ID: mediaID, FileName: "file.pdf"}, nil)
	item, err := svc.GetByID(ctx, mediaID)
	assert.NoError(t, err)
	assert.Equal(t, "file.pdf", item.FileName)

	repo.On("ListByModel", ctx, "practice", modelID, "attachments").
		Return([]domain.Media{{ID: uuid.New(), ModelType: "practice", ModelID: modelID}}, nil)
	items, err := svc.ListByModel(ctx, "practice", modelID, "attachments")
	assert.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestMediaDeleteRequiresAdminOrDeleteAnyPermission(t *testing.T) {
	mediaID := uuid.New()

	for _, tt := range []struct {
		name    string
		ctx     context.Context
		allowed bool
	}{
		{name: "plain caller forbidden", ctx: mediaCtx(false), allowed: false},
		{name: "delete_any permission allowed", ctx: mediaCtx(false, string(domain.PermMediaDeleteAny)), allowed: true},
		{name: "admin allowed", ctx: mediaCtx(true), allowed: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mediaRepoMock{}
			store := &storageMock{}
			svc := newMediaService(repo, store)
			if tt.allowed {
				repo.On("FindByID", tt.ctx, mediaID).
					Return(&domain.Media{ID: mediaID, ModelType: "practice", ModelID: uuid.New(), FileName: "file.pdf"}, nil)
				store.On("DeleteObject", tt.ctx, mock.Anything).Return(nil)
				repo.On("Delete", tt.ctx, mediaID).Return(nil)
			}

			err := svc.Delete(tt.ctx, mediaID)

			if tt.allowed {
				assert.NoError(t, err)
				repo.AssertExpectations(t)
				store.AssertExpectations(t)
			} else {
				assert.ErrorIs(t, err, domain.ErrForbidden)
				repo.AssertNotCalled(t, "Delete")
			}
		})
	}
}

func TestMediaDeleteRecordsAuditUnderTargetOrg(t *testing.T) {
	repo := &mediaRepoMock{}
	store := &storageMock{}
	svc, spy := newMediaServiceAudit(repo, store)

	orgID := uuid.New()
	mediaID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgID, IsAdmin: true})
	repo.On("FindByID", ctx, mediaID).
		Return(&domain.Media{ID: mediaID, OrganizationID: &orgID, ModelType: "practice", ModelID: uuid.New(), Name: "syllabus.pdf", FileName: "syllabus.pdf"}, nil)
	store.On("DeleteObject", ctx, mock.Anything).Return(nil)
	repo.On("Delete", ctx, mediaID).Return(nil)

	require.NoError(t, svc.Delete(ctx, mediaID))
	require.Len(t, spy.records, 1)
	rec := spy.records[0]
	assert.Equal(t, domain.AuditDeleted, rec.Action)
	assert.Equal(t, domain.AuditTargetMedia, rec.TargetType)
	assert.Equal(t, mediaID, *rec.TargetID)
	assert.Equal(t, "syllabus.pdf", rec.TargetLabel)
	require.NotNil(t, rec.OrgID)
	assert.Equal(t, orgID, *rec.OrgID)
}

func TestMediaDeletePlatformGlobalSkipsAudit(t *testing.T) {
	repo := &mediaRepoMock{}
	store := &storageMock{}
	svc, spy := newMediaServiceAudit(repo, store)

	mediaID := uuid.New()
	ctx := mediaCtx(true)
	repo.On("FindByID", ctx, mediaID).
		Return(&domain.Media{ID: mediaID, ModelType: "changelog", ModelID: uuid.New(), Name: "asset.png"}, nil)
	store.On("DeleteObject", ctx, mock.Anything).Return(nil)
	repo.On("Delete", ctx, mediaID).Return(nil)

	require.NoError(t, svc.Delete(ctx, mediaID))
	assert.Empty(t, spy.records)
}

func TestMediaCleanupByModelDeletesObjectsAndRows(t *testing.T) {
	repo := &mediaRepoMock{}
	store := &storageMock{}
	svc := newMediaService(repo, store)
	modelID := uuid.New()

	items := []domain.Media{
		{ID: uuid.New(), ModelType: domain.MediaModelLiveRoom, ModelID: modelID, CollectionName: domain.MediaCollectionSlides, FileName: "a.pdf"},
		{ID: uuid.New(), ModelType: domain.MediaModelLiveRoom, ModelID: modelID, CollectionName: domain.MediaCollectionSlides, FileName: "b.pdf"},
	}
	ctx := context.Background()
	repo.On("ListByModel", ctx, domain.MediaModelLiveRoom, modelID, domain.MediaCollectionSlides).Return(items, nil)
	for _, m := range items {
		store.On("DeleteObject", ctx, m.S3Key()).Return(nil)
		repo.On("Delete", ctx, m.ID).Return(nil)
	}

	err := svc.CleanupByModel(ctx, domain.MediaModelLiveRoom, modelID, domain.MediaCollectionSlides)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestMediaPresignUploadRequiresCallerBeforeRepositoryWrite(t *testing.T) {
	repo := &mediaRepoMock{}
	svc := newMediaService(repo, &storageMock{})

	resp, err := svc.PresignUpload(context.Background(), domain.PresignUploadDTO{
		ModelType:      "practice",
		ModelID:        uuid.NewString(),
		CollectionName: "attachments",
		FileName:       "file.pdf",
		MimeType:       "application/pdf",
		Size:           10,
	})

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Create")
}

func orgCtx(orgID uuid.UUID, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(), OrgID: &orgID, Permissions: perms,
	})
}

func TestMediaOrgScopeHidesOtherTenants(t *testing.T) {
	myOrg := uuid.New()
	otherOrg := uuid.New()
	mediaID := uuid.New()
	foreign := &domain.Media{ID: mediaID, OrganizationID: &otherOrg, FileName: "leak.pdf"}

	ctx := orgCtx(myOrg, string(domain.PermMediaDeleteAny))

	t.Run("GetByID", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		repo.On("FindByID", ctx, mediaID).Return(foreign, nil)
		_, err := svc.GetByID(ctx, mediaID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("PresignDownload", func(t *testing.T) {
		repo := &mediaRepoMock{}
		store := &storageMock{}
		svc := newMediaService(repo, store)
		repo.On("FindByID", ctx, mediaID).Return(foreign, nil)
		_, err := svc.PresignDownload(ctx, mediaID, 0)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		store.AssertNotCalled(t, "GeneratePresignedDownloadURL")
	})

	t.Run("Delete", func(t *testing.T) {
		repo := &mediaRepoMock{}
		store := &storageMock{}
		svc := newMediaService(repo, store)
		repo.On("FindByID", ctx, mediaID).Return(foreign, nil)
		err := svc.Delete(ctx, mediaID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		repo.AssertNotCalled(t, "Delete")
		store.AssertNotCalled(t, "DeleteObject")
	})

	t.Run("ListByModel filters foreign rows", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		modelID := uuid.New()
		repo.On("ListByModel", ctx, "live_room", modelID, "").Return([]domain.Media{
			{ID: uuid.New(), OrganizationID: &myOrg},
			{ID: uuid.New(), OrganizationID: &otherOrg},
			{ID: uuid.New(), OrganizationID: nil}, // platform-global stays visible
		}, nil)
		items, err := svc.ListByModel(ctx, "live_room", modelID, "")
		assert.NoError(t, err)
		assert.Len(t, items, 2)
	})
}

func TestMediaOrgScopeAllowsOwnOrgGlobalAndAdmin(t *testing.T) {
	myOrg := uuid.New()
	mediaID := uuid.New()

	t.Run("own org", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		ctx := orgCtx(myOrg)
		repo.On("FindByID", ctx, mediaID).Return(&domain.Media{ID: mediaID, OrganizationID: &myOrg}, nil)
		_, err := svc.GetByID(ctx, mediaID)
		assert.NoError(t, err)
	})

	t.Run("platform-global row", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		ctx := orgCtx(myOrg)
		repo.On("FindByID", ctx, mediaID).Return(&domain.Media{ID: mediaID}, nil)
		_, err := svc.GetByID(ctx, mediaID)
		assert.NoError(t, err)
	})

	t.Run("admin crosses orgs", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		ctx := mediaCtx(true)
		other := uuid.New()
		repo.On("FindByID", ctx, mediaID).Return(&domain.Media{ID: mediaID, OrganizationID: &other}, nil)
		_, err := svc.GetByID(ctx, mediaID)
		assert.NoError(t, err)
	})
}

func TestMediaPresignUploadSharedFolder(t *testing.T) {
	orgID := uuid.New()

	base := func() domain.PresignUploadDTO {
		return domain.PresignUploadDTO{
			ModelType: domain.MediaModelOrganization,
			ModelID:   orgID.String(),
			FileName:  "report.pdf",
			MimeType:  "application/pdf",
			Size:      100,
		}
	}

	t.Run("rejects foreign org id", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		dto := base()
		dto.ModelID = uuid.NewString()
		_, err := svc.PresignUpload(orgCtx(orgID), dto)
		var verr *domain.ValidationError
		assert.ErrorAs(t, err, &verr)
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("rejects oversized declared size", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		dto := base()
		dto.Size = 201 << 20
		_, err := svc.PresignUpload(orgCtx(orgID), dto)
		var verr *domain.ValidationError
		assert.ErrorAs(t, err, &verr)
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("forces shared collection and uniquifies key filename", func(t *testing.T) {
		repo := &mediaRepoMock{}
		store := &storageMock{}
		svc := newMediaService(repo, store)
		ctx := orgCtx(orgID)
		repo.On("Create", ctx, mock.MatchedBy(func(m *domain.Media) bool {
			return m.CollectionName == domain.MediaCollectionShared &&
				m.Name == "report.pdf" &&
				m.FileName != "report.pdf" &&
				strings.HasSuffix(m.FileName, "-report.pdf")
		})).Return(nil)
		store.On("GeneratePresignedUploadURL", ctx, mock.Anything, mock.Anything).Return("https://signed", nil)

		resp, err := svc.PresignUpload(ctx, base())

		assert.NoError(t, err)
		assert.Equal(t, "https://signed", resp.UploadURL)
		repo.AssertExpectations(t)
	})
}

func TestMediaPresignDownloadClampsExpiry(t *testing.T) {
	mediaID := uuid.New()
	for _, tt := range []struct {
		name string
		in   time.Duration
		want time.Duration
	}{
		{name: "zero falls back to 1h", in: 0, want: time.Hour},
		{name: "24h passes through", in: 24 * time.Hour, want: 24 * time.Hour},
		{name: "over 7d clamps to 7d", in: 30 * 24 * time.Hour, want: 7 * 24 * time.Hour},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mediaRepoMock{}
			store := &storageMock{}
			svc := newMediaService(repo, store)
			ctx := mediaCtx(false)
			m := &domain.Media{ID: mediaID, ModelType: "live_room", ModelID: uuid.New(), FileName: "a.pdf"}
			repo.On("FindByID", ctx, mediaID).Return(m, nil)
			store.On("GeneratePresignedDownloadURL", ctx, m.S3Key(), tt.want).Return("https://signed", nil)

			resp, err := svc.PresignDownload(ctx, mediaID, tt.in)

			assert.NoError(t, err)
			assert.Equal(t, "https://signed", resp.URL)
			store.AssertExpectations(t)
		})
	}
}

func TestMediaListFoldersAuthzAndSharedPin(t *testing.T) {
	orgID := uuid.New()

	t.Run("requires media:view_any", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		_, err := svc.ListFolders(orgCtx(orgID, string(domain.PermMediaView)))
		assert.ErrorIs(t, err, domain.ErrForbidden)
		repo.AssertNotCalled(t, "ListFolders")
	})

	t.Run("requires org-scoped caller", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		_, err := svc.ListFolders(mediaCtx(false, string(domain.PermMediaViewAny)))
		assert.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("appends shared folder when absent", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		ctx := orgCtx(orgID, string(domain.PermMediaViewAny))
		repo.On("ListFolders", ctx, orgID).Return([]domain.MediaFolder{
			{ModelType: "live_room", FileCount: 2, TotalSize: 30},
		}, nil)
		folders, err := svc.ListFolders(ctx)
		assert.NoError(t, err)
		require.Len(t, folders, 2)
		assert.Equal(t, domain.MediaModelOrganization, folders[1].ModelType)
		assert.Zero(t, folders[1].FileCount)
	})

	t.Run("keeps shared folder when present", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		ctx := orgCtx(orgID, string(domain.PermMediaViewAny))
		repo.On("ListFolders", ctx, orgID).Return([]domain.MediaFolder{
			{ModelType: domain.MediaModelOrganization, FileCount: 1, TotalSize: 5},
		}, nil)
		folders, err := svc.ListFolders(ctx)
		assert.NoError(t, err)
		require.Len(t, folders, 1)
		assert.Equal(t, int64(1), folders[0].FileCount)
	})
}

func TestMediaListFilesAuthz(t *testing.T) {
	orgID := uuid.New()

	t.Run("requires media:view_any", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		_, _, err := svc.ListFiles(orgCtx(orgID), "live_room", domain.ListParams{Page: 1, PageSize: 20})
		assert.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("delegates to repo with caller org", func(t *testing.T) {
		repo := &mediaRepoMock{}
		svc := newMediaService(repo, &storageMock{})
		ctx := orgCtx(orgID, string(domain.PermMediaViewAny))
		p := domain.ListParams{Page: 1, PageSize: 20}
		repo.On("ListFiles", ctx, orgID, "live_room", p).
			Return([]domain.Media{{ID: uuid.New()}}, int64(1), nil)
		items, total, err := svc.ListFiles(ctx, "live_room", p)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, items, 1)
	})
}

func TestMediaPresignUpload_StorageQuotaExceeded(t *testing.T) {
	repo := &mediaRepoMock{}
	orgID := uuid.New()
	ent := fakeEntService{storageErr: domain.NewLimitError(domain.PlanFree, domain.LimitStorageGB, 1, 1)}
	svc := media.NewService(repo, &storageMock{}, ent, nil, fakeTransactor{}, &auditSpy{}, slog.Default())

	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(), OrgID: &orgID, Ent: domain.PlanCatalog[domain.PlanFree],
	})
	_, err := svc.PresignUpload(ctx, domain.PresignUploadDTO{
		ModelType: "practice", ModelID: uuid.NewString(), CollectionName: "attachments",
		FileName: "f.pdf", MimeType: "application/pdf", Size: 1,
	})
	assert.ErrorIs(t, err, domain.ErrPlanLimitReached)
	repo.AssertNotCalled(t, "Create")
}

// usageStub satisfies media's storageUsageReader for the quota header.
type usageStub struct {
	used int64
	err  error
}

func (u usageStub) SumStorageBytes(context.Context, uuid.UUID) (int64, error) {
	return u.used, u.err
}

func ownersCtx(orgID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(), OrgID: &orgID, IsAdmin: true,
		Ent: domain.PlanCatalog[domain.PlanFree],
	})
}

func TestListOwners_MergesClassMediaAndRecordingsAndSortsBySize(t *testing.T) {
	repo := &mediaRepoMock{}
	orgID := uuid.New()
	classID := uuid.New()
	bankID := uuid.New()

	// Class has media (slides) AND recordings — they must sum into one row.
	repo.On("ListOwnerMedia", mock.Anything, orgID).Return([]domain.MediaOwner{
		{OwnerKind: domain.MediaOwnerClass, OwnerID: &classID, Name: "Math 101", FileCount: 2, TotalSize: 100},
		{OwnerKind: domain.MediaOwnerQuestionBank, OwnerID: &bankID, Name: "Algebra Bank", FileCount: 5, TotalSize: 500},
		{OwnerKind: domain.MediaOwnerShared, Name: "", FileCount: 1, TotalSize: 10},
	}, nil)
	repo.On("ListOwnerRecordings", mock.Anything, orgID).Return([]domain.MediaOwner{
		{OwnerKind: domain.MediaOwnerClass, OwnerID: &classID, Name: "Math 101", FileCount: 1, TotalSize: 9000},
	}, nil)

	svc := media.NewService(repo, &storageMock{}, nil, usageStub{used: 9610}, fakeTransactor{}, &auditSpy{}, slog.Default())
	resp, err := svc.ListOwners(ownersCtx(orgID), domain.ListParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, resp.Owners, 3)

	// Largest first: merged class (9100) > bank (500) > shared (10).
	assert.Equal(t, domain.MediaOwnerClass, resp.Owners[0].OwnerKind)
	assert.Equal(t, int64(9100), resp.Owners[0].TotalSize)
	assert.Equal(t, int64(3), resp.Owners[0].FileCount)
	assert.Equal(t, "Algebra Bank", resp.Owners[1].Name)
	assert.Equal(t, int64(3), resp.Total)

	// Free plan is a finite storage limit → quota header reconciles.
	assert.False(t, resp.Quota.Unlimited)
	assert.Equal(t, int64(9610), resp.Quota.UsedBytes)
	assert.Greater(t, resp.Quota.LimitBytes, int64(0))
}

// The union/read-only/sort/paginate logic for owner files lives in the repo's
// SQL now (see the integration test TestIntegration_MediaRepo_OwnerResolution).
// At the service layer we only verify org-scoped authz + faithful delegation.
func TestListOwnerFiles_DelegatesToRepoWithCallerOrg(t *testing.T) {
	repo := &mediaRepoMock{}
	orgID := uuid.New()
	classID := uuid.New()
	p := domain.ListParams{Page: 1, PageSize: 20, OrderBy: "size", OrderDir: "desc"}

	want := []domain.OwnerFile{
		{ID: uuid.NewString(), Source: "recording", Deletable: false, Size: 9000},
		{ID: uuid.NewString(), Source: "media", Deletable: true, Size: 100},
	}
	repo.On("ListOwnerFiles", mock.Anything, orgID, domain.MediaOwnerClass, &classID, p).Return(want, int64(2), nil)

	svc := media.NewService(repo, &storageMock{}, nil, usageStub{}, fakeTransactor{}, &auditSpy{}, slog.Default())
	files, total, err := svc.ListOwnerFiles(ownersCtx(orgID), domain.MediaOwnerClass, &classID, p)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, want, files)
	repo.AssertExpectations(t)
}

func TestListOwnerFiles_RequiresOrgViewAny(t *testing.T) {
	repo := &mediaRepoMock{}
	svc := media.NewService(repo, &storageMock{}, nil, usageStub{}, fakeTransactor{}, &auditSpy{}, slog.Default())

	// No caller in context → forbidden, repo never touched.
	_, _, err := svc.ListOwnerFiles(context.Background(), domain.MediaOwnerShared, nil, domain.ListParams{Page: 1, PageSize: 20})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "ListOwnerFiles", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
