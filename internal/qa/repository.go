package qa

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type questionRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.QARepository {
	return &questionRepository{db: db}
}

func (r *questionRepository) Create(ctx context.Context, q *domain.QAQuestion) error {
	if err := database.DB(ctx, r.db).Create(q).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("qa.repository.Create: %w", err)
	}
	return nil
}

func (r *questionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.QAQuestion, error) {
	var q domain.QAQuestion
	if err := database.DB(ctx, r.db).Model(&domain.QAQuestion{}).First(&q, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("qa.repository.FindByID: %w", err)
	}
	return &q, nil
}

func (r *questionRepository) Update(ctx context.Context, q *domain.QAQuestion) error {
	result := database.DB(ctx, r.db).Save(q)
	if result.Error != nil {
		return fmt.Errorf("qa.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *questionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.QAQuestion{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("qa.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List returns questions for a model ordered open-first, then by vote count
// desc, then oldest-first. vote_count and voted_by_me are computed per row for
// the viewer in scope.ViewerID.
func (r *questionRepository) List(ctx context.Context, scope domain.QAListScope, p domain.ListParams) ([]domain.QAQuestionView, int64, error) {
	// filtered carries only WHERE clauses so Count reflects filters without the
	// computed viewer-dependent SELECT (which would orphan its bind arg on Count).
	filtered := func() *gorm.DB {
		q := database.DB(ctx, r.db).
			Table("qa_questions AS q").
			Where("q.deleted_at IS NULL")
		if scope.ModelType != nil {
			q = q.Where("q.model_type = ?", *scope.ModelType)
		}
		if scope.ModelID != nil {
			q = q.Where("q.model_id = ?", *scope.ModelID)
		}
		if scope.Status != nil {
			q = q.Where("q.status = ?", *scope.Status)
		}
		return q
	}

	var total int64
	if err := filtered().Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("qa.repository.List (count): %w", err)
	}

	var views []domain.QAQuestionView
	err := filtered().
		Joins("JOIN users u ON u.id = q.user_id").
		Select(`q.id, q.user_id, u.name AS author_name, q.model_type, q.model_id,
			q.text, q.status, q.closed_at, q.created_at,
			(SELECT COUNT(*) FROM qa_votes v WHERE v.question_id = q.id) AS vote_count,
			EXISTS(SELECT 1 FROM qa_votes v WHERE v.question_id = q.id AND v.user_id = ?) AS voted_by_me`,
			scope.ViewerID).
		Order("(q.status = 'open') DESC, vote_count DESC, q.created_at ASC").
		Offset(p.Offset()).
		Limit(p.Limit()).
		Find(&views).Error
	if err != nil {
		return nil, 0, fmt.Errorf("qa.repository.List: %w", err)
	}
	return views, total, nil
}

func (r *questionRepository) CountOpenByUser(ctx context.Context, modelType string, modelID, userID uuid.UUID) (int64, error) {
	var count int64
	err := database.DB(ctx, r.db).Model(&domain.QAQuestion{}).
		Where("model_type = ? AND model_id = ? AND user_id = ? AND status = ?",
			modelType, modelID, userID, domain.QAStatusOpen).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("qa.repository.CountOpenByUser: %w", err)
	}
	return count, nil
}

func (r *questionRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.QAQuestion{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("qa.repository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *questionRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.QAQuestion, error) {
	var q domain.QAQuestion
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.QAQuestion{}).First(&q, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("qa.repository.FindByIDIncludingDeleted: %w", err)
	}
	return &q, nil
}

func (r *questionRepository) AdminList(ctx context.Context, q domain.AdminListQAQuestionsQuery) ([]domain.QAQuestion, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.QAQuestion{})
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	if q.ModelType != nil {
		base = base.Where("model_type = ?", *q.ModelType)
	}
	if q.ModelID != nil {
		base = base.Where("model_id = ?", *q.ModelID)
	}
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	var items []domain.QAQuestion
	total, err := listparams.Paginate(base, q.ListParams, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("qa.repository.AdminList: %w", err)
	}
	return items, total, nil
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) domain.QAVoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) Create(ctx context.Context, v *domain.QAVote) error {
	if err := database.DB(ctx, r.db).Create(v).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("qa.voteRepository.Create: %w", err)
	}
	return nil
}

func (r *voteRepository) Delete(ctx context.Context, questionID, userID uuid.UUID) (bool, error) {
	result := database.DB(ctx, r.db).
		Where("question_id = ? AND user_id = ?", questionID, userID).
		Delete(&domain.QAVote{})
	if result.Error != nil {
		return false, fmt.Errorf("qa.voteRepository.Delete: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}

func (r *voteRepository) Exists(ctx context.Context, questionID, userID uuid.UUID) (bool, error) {
	var count int64
	err := database.DB(ctx, r.db).Model(&domain.QAVote{}).
		Where("question_id = ? AND user_id = ?", questionID, userID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("qa.voteRepository.Exists: %w", err)
	}
	return count > 0, nil
}

func (r *voteRepository) CountByQuestion(ctx context.Context, questionID uuid.UUID) (int64, error) {
	var count int64
	err := database.DB(ctx, r.db).Model(&domain.QAVote{}).
		Where("question_id = ?", questionID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("qa.voteRepository.CountByQuestion: %w", err)
	}
	return count, nil
}
