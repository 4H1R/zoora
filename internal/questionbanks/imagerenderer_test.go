package questionbanks

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
)

// --- fakes (internal test package: define our own, service_test.go's mocks live
// in the external questionbanks_test package and aren't visible here) ---

type fakeStorage struct{ puts, deletes int }

func (f *fakeStorage) PutObject(_ context.Context, _ string, _ []byte, _ string) error {
	f.puts++
	return nil
}
func (f *fakeStorage) DeleteObject(_ context.Context, _ string) error {
	f.deletes++
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
	f.deleted = append(f.deleted, id)
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

func TestImageRenderer_RendersBodyAndChoiceOptions(t *testing.T) {
	q := &domain.Question{
		ID:                uuid.New(),
		OrganizationID:    uuid.New(),
		Type:              domain.QuestionTypeChoice,
		Text:              "پایتخت ایران کجاست؟",
		RenderAsImage:     true,
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

func TestImageRenderer_DisabledPurgesAndClears(t *testing.T) {
	q := &domain.Question{
		ID:                uuid.New(),
		OrganizationID:    uuid.New(),
		Type:              domain.QuestionTypeChoice,
		RenderAsImage:     false, // toggled off
		ImageRenderStatus: domain.ImageRenderStatusReady,
	}
	stale := domain.Media{ID: uuid.New(), ModelType: domain.QuestionMediaModelType, ModelID: q.ID, CollectionName: domain.QuestionSystemPhotosCollection, FileName: "body.png"}
	media := &fakeMedia{byCollection: map[string][]domain.Media{
		domain.QuestionSystemPhotosCollection: {stale},
	}}
	st := &fakeStorage{}
	r := &ImageRenderer{questions: &fakeQuestions{q: q}, media: media, storage: st, logger: slog.Default()}

	assert.NoError(t, r.Render(context.Background(), q.ID))
	assert.Equal(t, domain.ImageRenderStatusNone, q.ImageRenderStatus)
	assert.Nil(t, q.SystemImageMediaID)
	assert.Equal(t, []uuid.UUID{stale.ID}, media.deleted, "stale row purged")
	assert.Equal(t, 1, st.deletes, "stale object purged from storage")
	assert.Equal(t, 0, st.puts, "nothing rendered when disabled")
}
