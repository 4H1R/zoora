package questionbanks

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"golang.org/x/sync/errgroup"

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

// Render (re)generates a question's anti-cheat images and updates
// ImageRenderStatus. The task is enqueued (by the quiz service, or by a content
// edit to an already-participating question) only when a render is wanted, so
// the handler always renders. Purging on deletion is a separate cleanup task.
func (r *ImageRenderer) Render(ctx context.Context, questionID uuid.UUID) error {
	q, err := r.questions.FindByID(ctx, questionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil // question deleted between enqueue and processing — nothing to do
		}
		return fmt.Errorf("question render-images: load question: %w", err)
	}

	// Skip when the exact content was already rendered and the images are still
	// ready — the render is byte-identical (seed derives from stable ids), so a
	// re-enqueue for unchanged content is pure waste. Avoids purge + re-render +
	// S3 writes on the common "quiz re-saved" path.
	hash := renderContentHash(q)
	if q.ImageRenderStatus == domain.ImageRenderStatusReady &&
		q.SystemImageContentHash == hash && q.SystemImageMediaID != nil {
		return nil
	}

	// Idempotency: drop any previously generated media before doing anything else.
	if err := r.purgeSystemMedia(ctx, q.ID); err != nil {
		return err
	}
	clearSystemImages(q)

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
	q.SystemImageContentHash = hash
	if err := r.questions.Update(ctx, q); err != nil {
		return fmt.Errorf("question render-images: mark ready: %w", err)
	}
	r.logger.Info("question images rendered", "question_id", q.ID.String())
	return nil
}

// renderImageConcurrency bounds how many images a single question renders and
// uploads at once. Body + options fan out; the cap keeps one big question from
// saturating the worker's CPU and S3 connections.
const renderImageConcurrency = 4

// renderInto renders the body and (for choice questions) each non-empty option
// value concurrently, storing the media ids on q. It does not persist q. Each
// job writes a distinct field (body → q, option i → q.Options[i]), so no
// locking is needed around the assignments.
func (r *ImageRenderer) renderInto(ctx context.Context, q *domain.Question) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(renderImageConcurrency)

	g.Go(func() error {
		bodyPNG, err := imaging.RenderText(q.Text, imaging.Options{Noise: imaging.NoiseHigh, Seed: seedFromUUID(q.ID)})
		if err != nil {
			return fmt.Errorf("question render-images: render body: %w", err)
		}
		bodyID, err := r.storeImage(ctx, q, domain.QuestionSystemPhotosCollection, "body.png", bodyPNG)
		if err != nil {
			return err
		}
		q.SystemImageMediaID = &bodyID
		return nil
	})

	// Only choice options are shown to students; short_answer/descriptive options
	// hold answer keys and are stripped by the take endpoint, so don't render them.
	if q.Type == domain.QuestionTypeChoice {
		for i := range q.Options {
			o := &q.Options[i]
			if strings.TrimSpace(o.Value) == "" {
				continue // image-only option (teacher-uploaded ImageMediaID), nothing to render
			}
			g.Go(func() error {
				png, err := imaging.RenderText(o.Value, imaging.Options{Noise: imaging.NoiseHigh, Seed: seedFromString(o.ID)})
				if err != nil {
					return fmt.Errorf("question render-images: render option %s: %w", o.ID, err)
				}
				mid, err := r.storeImage(ctx, q, domain.QuestionOptionSystemPhotosCollection, "option-"+o.ID+".png", png)
				if err != nil {
					return err
				}
				o.SystemImageMediaID = &mid
				return nil
			})
		}
	}
	return g.Wait()
}

// renderContentHash fingerprints everything that affects the rendered images:
// the body text and, for choice questions, each option's id and value. A change
// to any of these changes the hash and forces a re-render.
func renderContentHash(q *domain.Question) string {
	h := sha256.New()
	// Writes to a hash.Hash never error.
	_, _ = fmt.Fprintf(h, "%s\x00%s\x00", q.Type, q.Text)
	if q.Type == domain.QuestionTypeChoice {
		for i := range q.Options {
			_, _ = fmt.Fprintf(h, "%s\x00%s\x00", q.Options[i].ID, q.Options[i].Value)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
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
