package questionbanks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/storage"
)

// copyObjectStorage is the slice of the S3 client the copier needs.
type copyObjectStorage interface {
	CopyObject(ctx context.Context, srcKey, dstKey string) error
}

// BankCopier clones a question bank into another organization when a share
// code is redeemed. It runs only in the worker: it copies every question, its
// uploaded media rows, and the underlying S3 objects (server-side CopyObject),
// then flips the target bank's status. Anti-cheat system images are NOT copied
// — clones start at ImageRenderStatusNone and re-render on demand.
//
// Idempotent: each run first purges the target's existing questions (and their
// media), so a retry after a mid-copy failure starts clean.
type BankCopier struct {
	banks     domain.QuestionBankRepository
	questions domain.QuestionRepository
	media     domain.MediaRepository
	mediaSvc  domain.MediaService
	storage   copyObjectStorage
	logger    *slog.Logger
}

func NewBankCopier(
	banks domain.QuestionBankRepository,
	questions domain.QuestionRepository,
	media domain.MediaRepository,
	mediaSvc domain.MediaService,
	storage *storage.Client,
	logger *slog.Logger,
) *BankCopier {
	return &BankCopier{banks: banks, questions: questions, media: media, mediaSvc: mediaSvc, storage: storage, logger: logger}
}

// NewCopyBankHandler adapts the copier to an Asynq handler for questionbank:copy.
func NewCopyBankHandler(c *BankCopier) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload domain.QuestionBankCopyPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("question bank copy: unmarshal payload: %w: %w", err, asynq.SkipRetry)
		}
		return c.Copy(ctx, payload)
	}
}

func (c *BankCopier) Copy(ctx context.Context, payload domain.QuestionBankCopyPayload) error {
	target, err := c.banks.FindByID(ctx, payload.TargetBankID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil // target deleted between enqueue and processing
		}
		return fmt.Errorf("question bank copy: load target: %w", err)
	}
	if target.Status == domain.QuestionBankStatusReady {
		return nil // duplicate delivery of an already-finished copy
	}

	source, err := c.banks.FindByID(ctx, payload.SourceBankID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Source deleted mid-flight — nothing to copy, no point retrying.
			c.markFailed(ctx, target)
			return nil
		}
		return fmt.Errorf("question bank copy: load source: %w", err)
	}

	// Idempotency: purge whatever a previous (failed) attempt copied in.
	existing, err := c.questions.ListAllByBank(ctx, target.ID)
	if err != nil {
		return fmt.Errorf("question bank copy: list target questions: %w", err)
	}
	for i := range existing {
		if err := c.mediaSvc.CleanupByModel(ctx, domain.QuestionMediaModelType, existing[i].ID, ""); err != nil {
			return fmt.Errorf("question bank copy: purge stale question media: %w", err)
		}
		if err := c.questions.HardDelete(ctx, existing[i].ID); err != nil && !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("question bank copy: purge stale question: %w", err)
		}
	}

	sourceQuestions, err := c.questions.ListAllByBank(ctx, source.ID)
	if err != nil {
		return fmt.Errorf("question bank copy: list source questions: %w", err)
	}
	for i := range sourceQuestions {
		if err := c.copyQuestion(ctx, &sourceQuestions[i], target); err != nil {
			c.markFailed(ctx, target)
			return fmt.Errorf("question bank copy: question %s: %w", sourceQuestions[i].ID.String(), err)
		}
	}

	target.Status = domain.QuestionBankStatusReady
	if err := c.banks.Update(ctx, target); err != nil {
		return fmt.Errorf("question bank copy: mark ready: %w", err)
	}
	c.logger.Info("question bank copied",
		"source_bank_id", source.ID.String(),
		"target_bank_id", target.ID.String(),
		"questions", len(sourceQuestions),
	)
	return nil
}

// markFailed best-effort flags the target so the redeemer sees the copy died.
// A later retry proceeds from 'failed' the same as from 'copying'.
func (c *BankCopier) markFailed(ctx context.Context, target *domain.QuestionBank) {
	target.Status = domain.QuestionBankStatusFailed
	if err := c.banks.Update(ctx, target); err != nil {
		c.logger.Error("question bank copy: mark failed", "bank_id", target.ID.String(), "error", err)
	}
}

func (c *BankCopier) copyQuestion(ctx context.Context, src *domain.Question, target *domain.QuestionBank) error {
	options := make([]domain.QuestionOption, len(src.Options))
	for i, o := range src.Options {
		o.SystemImageMediaID = nil // rendered per-org; the clone re-renders on demand
		o.ImageMediaID = nil       // remapped below once the media rows are cloned
		options[i] = o
	}

	clone := &domain.Question{
		BankID:            target.ID,
		OrganizationID:    target.OrganizationID,
		Text:              src.Text,
		Type:              src.Type,
		Options:           options,
		Metadata:          []domain.QuestionMetadata{},
		ModelAnswer:       src.ModelAnswer,
		NegativeMarkMode:  src.NegativeMarkMode,
		NegativeValue:     src.NegativeValue,
		WrongsPerPoint:    src.WrongsPerPoint,
		MinSeconds:        src.MinSeconds,
		ImageRenderStatus: domain.ImageRenderStatusNone,
	}
	if err := c.questions.Create(ctx, clone); err != nil {
		return fmt.Errorf("create clone: %w", err)
	}

	// Media rows are keyed by question, so the clone's ID had to exist first;
	// now clone each referenced media row + S3 object and remap the references.
	dirty := false
	for _, m := range src.Metadata {
		newID, err := c.copyMedia(ctx, m.MediaID, clone.ID, target.OrganizationID)
		if err != nil {
			return err
		}
		if newID == nil {
			continue // source media row vanished — drop the dangling reference
		}
		clone.Metadata = append(clone.Metadata, domain.QuestionMetadata{Type: m.Type, MediaID: *newID})
		dirty = true
	}
	for i, o := range src.Options {
		if o.ImageMediaID == nil {
			continue
		}
		newID, err := c.copyMedia(ctx, *o.ImageMediaID, clone.ID, target.OrganizationID)
		if err != nil {
			return err
		}
		if newID == nil {
			continue
		}
		clone.Options[i].ImageMediaID = newID
		dirty = true
	}
	if dirty {
		if err := c.questions.Update(ctx, clone); err != nil {
			return fmt.Errorf("save remapped media refs: %w", err)
		}
	}
	return nil
}

// copyMedia clones one media row (and its S3 object) onto the clone question.
// Returns nil (no error) when the source row no longer exists.
func (c *BankCopier) copyMedia(ctx context.Context, mediaID, cloneQuestionID uuid.UUID, targetOrgID uuid.UUID) (*uuid.UUID, error) {
	src, err := c.media.FindByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("load media %s: %w", mediaID.String(), err)
	}

	props := src.CustomProperties
	if len(props) == 0 {
		props = json.RawMessage("{}")
	}
	orgID := targetOrgID
	dst := &domain.Media{
		OrganizationID:   &orgID,
		ModelType:        domain.QuestionMediaModelType,
		ModelID:          cloneQuestionID,
		CollectionName:   src.CollectionName,
		Name:             src.Name,
		FileName:         src.FileName,
		MimeType:         src.MimeType,
		Disk:             src.Disk,
		Size:             src.Size,
		CustomProperties: props,
		OrderColumn:      src.OrderColumn,
	}
	if err := c.media.Create(ctx, dst); err != nil {
		return nil, fmt.Errorf("create media clone: %w", err)
	}
	if err := c.storage.CopyObject(ctx, src.S3Key(), dst.S3Key()); err != nil {
		return nil, err
	}
	return &dst.ID, nil
}
