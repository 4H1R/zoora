package livesessions

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
)

// recordingObjectStore is the subset of the storage client the retention sweep
// needs (delete the recording's S3 object).
type recordingObjectStore interface {
	DeleteObject(ctx context.Context, key string) error
}

// RecordingWithPlan is one recording joined with its owning org's plan, so the
// sweep can resolve per-org retention without an N+1 query.
type RecordingWithPlan struct {
	ID            uuid.UUID
	FileURL       string
	StartedAt     time.Time
	Plan          domain.Plan
	PlanExpiresAt *time.Time
}

// RetentionRepository lists recordings alongside their org plan and hard-deletes
// expired rows. live_recordings has no soft delete, so Delete is a hard delete.
type RetentionRepository interface {
	ListRecordingsWithPlan(ctx context.Context) ([]RecordingWithPlan, error)
	DeleteRecording(ctx context.Context, id uuid.UUID) error
}

type retentionRepository struct{ db *gorm.DB }

func NewRetentionRepository(db *gorm.DB) RetentionRepository {
	return &retentionRepository{db: db}
}

func (r *retentionRepository) ListRecordingsWithPlan(ctx context.Context) ([]RecordingWithPlan, error) {
	var rows []RecordingWithPlan
	// live_recordings -> live_rooms -> class_sessions -> classes -> organizations
	err := r.db.WithContext(ctx).
		Table("live_recordings AS lr").
		Select("lr.id AS id, lr.file_url AS file_url, lr.started_at AS started_at, o.plan AS plan, o.plan_expires_at AS plan_expires_at").
		Joins("JOIN live_rooms rooms ON rooms.id = lr.live_room_id").
		Joins("JOIN class_sessions cs ON cs.id = rooms.class_session_id").
		Joins("JOIN classes c ON c.id = cs.class_id").
		Joins("JOIN organizations o ON o.id = c.organization_id AND o.deleted_at IS NULL").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("livesessions.retentionRepository.ListRecordingsWithPlan: %w", err)
	}
	return rows, nil
}

func (r *retentionRepository) DeleteRecording(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.LiveRecording{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("livesessions.retentionRepository.DeleteRecording: %w", err)
	}
	return nil
}

// RetentionSweeper deletes recordings older than their org plan's retention
// window. retention == 0 means keep forever (Free/unset), consistent with the
// grandfather policy — a downgraded org never has its old recordings destroyed.
type RetentionSweeper struct {
	repo    RetentionRepository
	storage recordingObjectStore
	logger  *slog.Logger
}

func NewRetentionSweeper(repo RetentionRepository, storage recordingObjectStore, logger *slog.Logger) *RetentionSweeper {
	return &RetentionSweeper{repo: repo, storage: storage, logger: logger}
}

// Sweep runs one retention pass. now is injected so callers/tests control time.
func (s *RetentionSweeper) Sweep(ctx context.Context, now time.Time) error {
	rows, err := s.repo.ListRecordingsWithPlan(ctx)
	if err != nil {
		return err
	}
	var deleted int
	for _, row := range rows {
		ent := domain.EffectiveEntitlements(row.Plan, row.PlanExpiresAt, now)
		retentionDays := ent.Limit(domain.LimitRecordingRetentionDays)
		if retentionDays <= 0 {
			continue // keep forever
		}
		cutoff := now.Add(-time.Duration(retentionDays) * 24 * time.Hour)
		if !row.StartedAt.Before(cutoff) {
			continue // still within the retention window
		}
		if row.FileURL != "" {
			if err := s.storage.DeleteObject(ctx, row.FileURL); err != nil {
				// Log and skip the row delete so a retry can reclaim the object.
				s.logger.Warn("retention: deleting recording object", "recording_id", row.ID.String(), "key", row.FileURL, "error", err)
				continue
			}
		}
		if err := s.repo.DeleteRecording(ctx, row.ID); err != nil {
			s.logger.Warn("retention: deleting recording row", "recording_id", row.ID.String(), "error", err)
			continue
		}
		deleted++
	}
	s.logger.Info("recording retention sweep complete", "scanned", len(rows), "deleted", deleted)
	return nil
}

// NewRetentionSweepHandler adapts the sweeper to an Asynq handler.
func NewRetentionSweepHandler(sweeper *RetentionSweeper) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		return sweeper.Sweep(ctx, time.Now())
	}
}
