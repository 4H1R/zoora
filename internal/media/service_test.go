package media_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/media"
)

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

func newMediaService(repo *mediaRepoMock) domain.MediaService {
	return media.NewService(repo, nil, slog.Default())
}

func mediaCtx(isAdmin bool, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: isAdmin, Permissions: perms})
}

func TestMediaGetAndListRequireCaller(t *testing.T) {
	repo := &mediaRepoMock{}
	svc := newMediaService(repo)
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
			svc := newMediaService(repo)
			if tt.allowed {
				repo.On("Delete", tt.ctx, mediaID).Return(nil)
			}

			err := svc.Delete(tt.ctx, mediaID)

			if tt.allowed {
				assert.NoError(t, err)
				repo.AssertExpectations(t)
			} else {
				assert.ErrorIs(t, err, domain.ErrForbidden)
				repo.AssertNotCalled(t, "Delete")
			}
		})
	}
}

func TestMediaPresignUploadRequiresCallerBeforeRepositoryWrite(t *testing.T) {
	repo := &mediaRepoMock{}
	svc := newMediaService(repo)

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
