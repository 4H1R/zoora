package questionbanks

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
)

// --- fakes (internal test package: define our own, service_test.go's mocks live
// in the external questionbanks_test package and aren't visible here) ---

type fakeStorage struct {
	mu            sync.Mutex
	puts, deletes int
}

func (f *fakeStorage) PutObject(_ context.Context, _ string, _ []byte, _ string) error {
	f.mu.Lock()
	f.puts++
	f.mu.Unlock()
	return nil
}
func (f *fakeStorage) DeleteObject(_ context.Context, _ string) error {
	f.mu.Lock()
	f.deletes++
	f.mu.Unlock()
	return nil
}

type fakeQuestions struct {
	q         *domain.Question
	updateErr error
}

func (f *fakeQuestions) FindByID(_ context.Context, _ uuid.UUID) (*domain.Question, error) {
	if f.q == nil {
		return nil, domain.ErrNotFound
	}
	return f.q, nil
}
func (f *fakeQuestions) Update(_ context.Context, _ *domain.Question) error { return f.updateErr }
func (f *fakeQuestions) Create(_ context.Context, _ *domain.Question) error { return nil }
func (f *fakeQuestions) Delete(_ context.Context, _ uuid.UUID) error        { return nil }
func (f *fakeQuestions) ListByBank(_ context.Context, _ uuid.UUID, _ domain.ListQuestionsQuery) ([]domain.Question, int64, error) {
	return nil, 0, nil
}
func (f *fakeQuestions) ListAllByBank(_ context.Context, _ uuid.UUID) ([]domain.Question, error) {
	return nil, nil
}
func (f *fakeQuestions) FindByIDs(_ context.Context, _ []uuid.UUID) ([]domain.Question, error) {
	return nil, nil
}
func (f *fakeQuestions) CountByBank(_ context.Context, _ uuid.UUID) (int64, error) { return 0, nil }
func (f *fakeQuestions) RandomByBank(_ context.Context, _ uuid.UUID, _ int) ([]domain.Question, error) {
	return nil, nil
}
func (f *fakeQuestions) HardDelete(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeQuestions) AdminList(_ context.Context, _ domain.AdminListQuestionsQuery) ([]domain.Question, int64, error) {
	return nil, 0, nil
}

type fakeMedia struct {
	mu           sync.Mutex
	byCollection map[string][]domain.Media
	deleted      []uuid.UUID
}

func (f *fakeMedia) Create(_ context.Context, m *domain.Media) error {
	m.ID = uuid.New()
	return nil
}
func (f *fakeMedia) FindByID(_ context.Context, _ uuid.UUID) (*domain.Media, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeMedia) Delete(_ context.Context, id uuid.UUID) error {
	f.mu.Lock()
	f.deleted = append(f.deleted, id)
	f.mu.Unlock()
	return nil
}
func (f *fakeMedia) ListByModel(_ context.Context, _ string, _ uuid.UUID, collection string) ([]domain.Media, error) {
	return f.byCollection[collection], nil
}
func (f *fakeMedia) ListFolders(_ context.Context, _ uuid.UUID) ([]domain.MediaFolder, error) {
	return nil, nil
}
func (f *fakeMedia) ListFiles(_ context.Context, _ uuid.UUID, _ string, _ domain.ListParams) ([]domain.Media, int64, error) {
	return nil, 0, nil
}
func (f *fakeMedia) ListOwnerMedia(_ context.Context, _ uuid.UUID) ([]domain.MediaOwner, error) {
	return nil, nil
}
func (f *fakeMedia) ListOwnerRecordings(_ context.Context, _ uuid.UUID) ([]domain.MediaOwner, error) {
	return nil, nil
}
func (f *fakeMedia) ListOwnerFiles(_ context.Context, _ uuid.UUID, _ string, _ *uuid.UUID, _ domain.ListParams) ([]domain.OwnerFile, int64, error) {
	return nil, 0, nil
}

func TestImageRenderer_RendersBodyAndChoiceOptions(t *testing.T) {
	q := &domain.Question{
		ID:                uuid.New(),
		OrganizationID:    uuid.New(),
		Type:              domain.QuestionTypeChoice,
		Text:              "پایتخت ایران کجاست؟",
		ImageRenderStatus: domain.ImageRenderStatusPending,
		Options: []domain.QuestionOption{
			{ID: "a", Value: "تهران", Score: 1},
			{ID: "b", Value: "شیراز", Score: 0},
			{ID: "c", Value: "", Score: 0}, // image-only option: no value to render
		},
	}
	st := &fakeStorage{}
	r := &ImageRenderer{questions: &fakeQuestions{q: q}, media: &fakeMedia{}, storage: st, logger: slog.Default()}

	assert.NoError(t, r.Render(context.Background(), q.ID))

	assert.Equal(t, domain.ImageRenderStatusReady, q.ImageRenderStatus)
	assert.NotNil(t, q.SystemImageMediaID, "body image should be set")
	assert.NotNil(t, q.Options[0].SystemImageMediaID)
	assert.NotNil(t, q.Options[1].SystemImageMediaID)
	assert.Nil(t, q.Options[2].SystemImageMediaID, "empty-value option gets no image")
	assert.Equal(t, 3, st.puts, "body + two non-empty options")
}

// TestImageRenderer_SkipsWhenContentUnchanged verifies a re-enqueue for a
// question that is already 'ready' with matching content hash does no work.
func TestImageRenderer_SkipsWhenContentUnchanged(t *testing.T) {
	existing := uuid.New()
	q := &domain.Question{
		ID:                 uuid.New(),
		OrganizationID:     uuid.New(),
		Type:               domain.QuestionTypeChoice,
		Text:               "پایتخت ایران کجاست؟",
		ImageRenderStatus:  domain.ImageRenderStatusReady,
		SystemImageMediaID: &existing,
		Options: []domain.QuestionOption{
			{ID: "a", Value: "تهران", Score: 1},
		},
	}
	q.SystemImageContentHash = renderContentHash(q)
	st := &fakeStorage{}
	media := &fakeMedia{}
	r := &ImageRenderer{questions: &fakeQuestions{q: q}, media: media, storage: st, logger: slog.Default()}

	assert.NoError(t, r.Render(context.Background(), q.ID))
	assert.Equal(t, 0, st.puts, "unchanged content should not re-render")
	assert.Equal(t, 0, st.deletes, "unchanged content should not purge")
	assert.Empty(t, media.deleted)
	assert.Equal(t, domain.ImageRenderStatusReady, q.ImageRenderStatus)

	// Editing the text changes the hash → a render happens.
	q.Text = "changed"
	assert.NoError(t, r.Render(context.Background(), q.ID))
	assert.Greater(t, st.puts, 0, "changed content must re-render")
}

// TestImageRenderer_PurgesStaleBeforeRerender verifies the renderer drops any
// previously generated media before re-rendering (idempotency on re-save).
func TestImageRenderer_PurgesStaleBeforeRerender(t *testing.T) {
	q := &domain.Question{
		ID:                uuid.New(),
		OrganizationID:    uuid.New(),
		Type:              domain.QuestionTypeChoice,
		Text:              "پایتخت ایران کجاست؟",
		ImageRenderStatus: domain.ImageRenderStatusPending,
		Options: []domain.QuestionOption{
			{ID: "a", Value: "تهران", Score: 1},
		},
	}
	stale := domain.Media{ID: uuid.New(), ModelType: domain.QuestionMediaModelType, ModelID: q.ID, CollectionName: domain.QuestionSystemPhotosCollection, FileName: "body.png"}
	media := &fakeMedia{byCollection: map[string][]domain.Media{
		domain.QuestionSystemPhotosCollection: {stale},
	}}
	st := &fakeStorage{}
	r := &ImageRenderer{questions: &fakeQuestions{q: q}, media: media, storage: st, logger: slog.Default()}

	assert.NoError(t, r.Render(context.Background(), q.ID))
	assert.Equal(t, domain.ImageRenderStatusReady, q.ImageRenderStatus)
	assert.Equal(t, []uuid.UUID{stale.ID}, media.deleted, "stale row purged before re-render")
	assert.Equal(t, 1, st.deletes, "stale object purged from storage")
}
