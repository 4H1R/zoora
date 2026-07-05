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

func newMediaService(repo *mediaRepoMock, store *storageMock) domain.MediaService {
	return media.NewService(repo, store, nil, slog.Default())
}

func mediaCtx(isAdmin bool, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: isAdmin, Permissions: perms})
}

// fakeEntService injects a canned CheckStorageLimit result for quota tests.
type fakeEntService struct{ storageErr error }

func (f fakeEntService) CheckUserLimit(context.Context, uuid.UUID, domain.Entitlements) error {
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

func TestMediaPresignUpload_StorageQuotaExceeded(t *testing.T) {
	repo := &mediaRepoMock{}
	orgID := uuid.New()
	ent := fakeEntService{storageErr: domain.NewLimitError(domain.PlanFree, domain.LimitStorageGB, 1, 1)}
	svc := media.NewService(repo, &storageMock{}, ent, slog.Default())

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
