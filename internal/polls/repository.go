package polls

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

type pollRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.PollRepository {
	return &pollRepository{db: db}
}

func (r *pollRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.Poll{})
}

func (r *pollRepository) Create(ctx context.Context, poll *domain.Poll) error {
	if err := database.DB(ctx, r.db).Create(poll).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("polls.repository.Create: %w", err)
	}
	return nil
}

func (r *pollRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Poll, error) {
	var poll domain.Poll
	if err := r.baseQuery(ctx).First(&poll, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("polls.repository.FindByID: %w", err)
	}
	return &poll, nil
}

func (r *pollRepository) Update(ctx context.Context, poll *domain.Poll) error {
	result := database.DB(ctx, r.db).Save(poll)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("polls.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *pollRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Poll{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("polls.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *pollRepository) List(ctx context.Context, scope domain.PollListScope, p domain.ListParams) ([]domain.Poll, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Poll{})
	if scope.IncludeDeleted {
		base = base.Unscoped()
	}
	if scope.ModelType != nil {
		base = base.Where("model_type = ?", *scope.ModelType)
	}
	if scope.ModelID != nil {
		base = base.Where("model_id = ?", *scope.ModelID)
	}
	if !scope.AllOrgs {
		if scope.OwnerID != nil {
			base = base.Where("user_id = ?", *scope.OwnerID)
		}
	}
	var polls []domain.Poll
	total, err := listparams.Paginate(base, p, &polls)
	if err != nil {
		return nil, 0, fmt.Errorf("polls.repository.List: %w", err)
	}
	return polls, total, nil
}

func (r *pollRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.Poll{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("polls.repository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *pollRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Poll, error) {
	var poll domain.Poll
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.Poll{}).First(&poll, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("polls.repository.FindByIDIncludingDeleted: %w", err)
	}
	return &poll, nil
}

func (r *pollRepository) AdminList(ctx context.Context, q domain.AdminListPollsQuery) ([]domain.Poll, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Poll{})
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
	var polls []domain.Poll
	total, err := listparams.Paginate(base, q.ListParams, &polls)
	if err != nil {
		return nil, 0, fmt.Errorf("polls.repository.AdminList: %w", err)
	}
	return polls, total, nil
}

// answerRepository handles poll_answers persistence.
type answerRepository struct {
	db *gorm.DB
}

func NewAnswerRepository(db *gorm.DB) domain.PollAnswerRepository {
	return &answerRepository{db: db}
}

func (r *answerRepository) Create(ctx context.Context, answer *domain.PollAnswer) error {
	if err := database.DB(ctx, r.db).Create(answer).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("polls.answerRepository.Create: %w", err)
	}
	return nil
}

func (r *answerRepository) FindByPollAndUser(ctx context.Context, pollID, userID uuid.UUID) ([]domain.PollAnswer, error) {
	var answers []domain.PollAnswer
	if err := database.DB(ctx, r.db).Model(&domain.PollAnswer{}).
		Where("poll_id = ? AND user_id = ?", pollID, userID).
		Find(&answers).Error; err != nil {
		return nil, fmt.Errorf("polls.answerRepository.FindByPollAndUser: %w", err)
	}
	return answers, nil
}

func (r *answerRepository) ListByPoll(ctx context.Context, pollID uuid.UUID, q domain.ListPollAnswersQuery) ([]domain.PollAnswer, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.PollAnswer{}).Where("poll_id = ?", pollID)
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	var answers []domain.PollAnswer
	total, err := listparams.Paginate(base, q.ListParams, &answers)
	if err != nil {
		return nil, 0, fmt.Errorf("polls.answerRepository.ListByPoll: %w", err)
	}
	return answers, total, nil
}

func (r *answerRepository) DeleteByPollAndUser(ctx context.Context, pollID, userID uuid.UUID) error {
	result := database.DB(ctx, r.db).
		Where("poll_id = ? AND user_id = ?", pollID, userID).
		Delete(&domain.PollAnswer{})
	if result.Error != nil {
		return fmt.Errorf("polls.answerRepository.DeleteByPollAndUser: %w", result.Error)
	}
	return nil
}
