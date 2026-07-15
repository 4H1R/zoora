package questionbanks

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/imaging"
	"github.com/4H1R/zoora/internal/platform/storage"
)

// objectStorage is the slice of the S3 client the renderer needs. Kept narrow so
// the worker-side dependency is explicit and the renderer is easy to fake.
type objectStorage interface {
	PutObject(ctx context.Context, key string, body []byte, contentType string) error
	DeleteObject(ctx context.Context, key string) error
}

// ImageRenderer generates (and purges) a question's anti-cheat images. It runs
// only in the worker: it renders the body text and each choice option value to
// distorted PNGs, stores them as question-owned media, and records the media ids
// back on the question. All work is idempotent — every run first purges the
// previously generated media — so retries and re-enqueues are safe.
type ImageRenderer struct {
	questions domain.QuestionRepository
	media     domain.MediaRepository
	storage   objectStorage
	logger    *slog.Logger
}

func NewImageRenderer(
	questions domain.QuestionRepository,
	media domain.MediaRepository,
	storage *storage.Client,
	logger *slog.Logger,
) *ImageRenderer {
	return &ImageRenderer{questions: questions, media: media, storage: storage, logger: logger}
}

// NewRenderImagesHandler adapts the renderer to an Asynq handler for
// question:render-images.
func NewRenderImagesHandler(r *ImageRenderer) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload domain.QuestionRenderImagesPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			// Malformed payload is unrecoverable — don't retry.
			return fmt.Errorf("question render-images: unmarshal payload: %w: %w", err, asynq.SkipRetry)
		}
		return r.Render(ctx, payload.QuestionID)
	}
}

// Render (re)generates or purges a question's anti-cheat images to match its
// current RenderAsImage flag, then updates ImageRenderStatus.
func (r *ImageRenderer) Render(ctx context.Context, questionID uuid.UUID) error {
	q, err := r.questions.FindByID(ctx, questionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil // question deleted between enqueue and processing — nothing to do
		}
		return fmt.Errorf("question render-images: load question: %w", err)
	}

	// Idempotency: drop any previously generated media before doing anything else.
	if err := r.purgeSystemMedia(ctx, q.ID); err != nil {
		return err
	}
	clearSystemImages(q)

	if !q.RenderAsImage {
		q.ImageRenderStatus = domain.ImageRenderStatusNone
		if err := r.questions.Update(ctx, q); err != nil {
			return fmt.Errorf("question render-images: clear status: %w", err)
		}
		return nil
	}

	if err := r.renderInto(ctx, q); err != nil {
		// Leave the images purged and flag failure so the take gate keeps blocking
		// and the teacher can re-save to retry.
		q.ImageRenderStatus = domain.ImageRenderStatusFailed
		if uerr := r.questions.Update(ctx, q); uerr != nil {
			r.logger.Error("question render-images: mark failed", "question_id", q.ID.String(), "error", uerr)
		}
		return err
	}

	q.ImageRenderStatus = domain.ImageRenderStatusReady
	if err := r.questions.Update(ctx, q); err != nil {
		return fmt.Errorf("question render-images: mark ready: %w", err)
	}
	r.logger.Info("question images rendered", "question_id", q.ID.String())
	return nil
}

// renderInto renders the body and (for choice questions) each non-empty option
// value, storing the media ids on q. It does not persist q.
func (r *ImageRenderer) renderInto(ctx context.Context, q *domain.Question) error {
	bodyPNG, err := imaging.RenderText(q.Text, imaging.Options{Noise: imaging.NoiseMedium, Seed: seedFromUUID(q.ID)})
	if err != nil {
		return fmt.Errorf("question render-images: render body: %w", err)
	}
	bodyID, err := r.storeImage(ctx, q, domain.QuestionSystemPhotosCollection, "body.png", bodyPNG)
	if err != nil {
		return err
	}
	q.SystemImageMediaID = &bodyID

	// Only choice options are shown to students; short_answer/descriptive options
	// hold answer keys and are stripped by the take endpoint, so don't render them.
	if q.Type != domain.QuestionTypeChoice {
		return nil
	}
	for i := range q.Options {
		o := &q.Options[i]
		if strings.TrimSpace(o.Value) == "" {
			continue // image-only option (teacher-uploaded ImageMediaID), nothing to render
		}
		png, err := imaging.RenderText(o.Value, imaging.Options{Noise: imaging.NoiseMedium, Seed: seedFromString(o.ID)})
		if err != nil {
			return fmt.Errorf("question render-images: render option %s: %w", o.ID, err)
		}
		mid, err := r.storeImage(ctx, q, domain.QuestionOptionSystemPhotosCollection, "option-"+o.ID+".png", png)
		if err != nil {
			return err
		}
		o.SystemImageMediaID = &mid
	}
	return nil
}

// storeImage creates a question-owned media row and uploads the PNG to S3 under
// its key, returning the new media id.
func (r *ImageRenderer) storeImage(ctx context.Context, q *domain.Question, collection, fileName string, data []byte) (uuid.UUID, error) {
	orgID := q.OrganizationID
	m := &domain.Media{
		OrganizationID: &orgID,
		ModelType:      domain.QuestionMediaModelType,
		ModelID:        q.ID,
		CollectionName: collection,
		Name:           fileName,
		FileName:       fileName,
		MimeType:       "image/png",
		Disk:           "s3",
		Size:           int64(len(data)),
	}
	if err := r.media.Create(ctx, m); err != nil {
		return uuid.Nil, fmt.Errorf("question render-images: create media: %w", err)
	}
	if err := r.storage.PutObject(ctx, m.S3Key(), data, "image/png"); err != nil {
		return uuid.Nil, fmt.Errorf("question render-images: put object: %w", err)
	}
	return m.ID, nil
}

// purgeSystemMedia deletes every generated image (rows + S3 objects) for a
// question across both system collections. Idempotent: an empty set is a no-op.
func (r *ImageRenderer) purgeSystemMedia(ctx context.Context, questionID uuid.UUID) error {
	for _, coll := range []string{domain.QuestionSystemPhotosCollection, domain.QuestionOptionSystemPhotosCollection} {
		items, err := r.media.ListByModel(ctx, domain.QuestionMediaModelType, questionID, coll)
		if err != nil {
			return fmt.Errorf("question render-images: list system media: %w", err)
		}
		for i := range items {
			if err := r.storage.DeleteObject(ctx, items[i].S3Key()); err != nil {
				return fmt.Errorf("question render-images: delete object: %w", err)
			}
			if err := r.media.Delete(ctx, items[i].ID); err != nil {
				return fmt.Errorf("question render-images: delete media row: %w", err)
			}
		}
	}
	return nil
}

// seedFromUUID / seedFromString derive a stable render seed so the same content
// always produces byte-identical distortion (re-render idempotency).
func seedFromUUID(id uuid.UUID) int64 {
	return int64(binary.LittleEndian.Uint64(id[:8]))
}

func seedFromString(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return int64(h.Sum64())
}
