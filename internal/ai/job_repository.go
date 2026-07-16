package ai

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
)

// JobRepository persists AI grading jobs. Implements domain.AIGradingJobRepository.
type JobRepository struct {
	db *gorm.DB
}

// NewJobRepository builds the grading-job repository.
func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(ctx context.Context, job *domain.AIGradingJob) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		return fmt.Errorf("creating ai grading job: %w", err)
	}
	return nil
}

func (r *JobRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.AIGradingJob, error) {
	var job domain.AIGradingJob
	if err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("finding ai grading job: %w", err)
	}
	return &job, nil
}

func (r *JobRepository) MarkRunning(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.AIGradingJob{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": domain.AIGradingStatusRunning, "updated_at": gorm.Expr("NOW()")}).Error
}

// IncrementProgress atomically bumps counters and completes the job when all
// tasks have reported. A single UPDATE keeps concurrent task completions correct.
func (r *JobRepository) IncrementProgress(ctx context.Context, id uuid.UUID, doneDelta, failedDelta int) error {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE ai_grading_jobs
		SET done = done + ?,
		    failed = failed + ?,
		    status = CASE WHEN (done + ? + failed + ?) >= total THEN 'completed' ELSE 'running' END,
		    updated_at = NOW()
		WHERE id = ?`,
		doneDelta, failedDelta, doneDelta, failedDelta, id)
	if res.Error != nil {
		return fmt.Errorf("incrementing ai grading job progress: %w", res.Error)
	}
	return nil
}
