package quizzes

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

type quizRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.QuizRepository {
	return &quizRepository{db: db}
}

func (r *quizRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.Quiz{})
}

func (r *quizRepository) Create(ctx context.Context, quiz *domain.Quiz) error {
	if err := database.DB(ctx, r.db).Create(quiz).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("quizzes.repository.Create: %w", err)
	}
	return nil
}

func (r *quizRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	var quiz domain.Quiz
	if err := r.baseQuery(ctx).Preload("User").Preload("Class").First(&quiz, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("quizzes.repository.FindByID: %w", err)
	}
	return &quiz, nil
}

func (r *quizRepository) Update(ctx context.Context, quiz *domain.Quiz) error {
	result := database.DB(ctx, r.db).Save(quiz)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("quizzes.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *quizRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Quiz{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("quizzes.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *quizRepository) List(ctx context.Context, scope domain.QuizListScope, p domain.ListParams) ([]domain.Quiz, int64, error) {
	db := database.DB(ctx, r.db)
	base := db.Model(&domain.Quiz{}).Preload("User").Preload("Class")
	if scope.IncludeDeleted {
		base = base.Unscoped()
	}
	if scope.OrganizationID != nil {
		base = base.Where("organization_id = ?", *scope.OrganizationID)
	}
	if scope.ClassID != nil {
		base = base.Where("class_id = ?", *scope.ClassID)
	}
	if scope.ClassSessionID != nil {
		sub := db.Table("quiz_rooms").
			Select("quiz_id").
			Where("class_session_id = ?", *scope.ClassSessionID)
		base = base.Where("id IN (?)", sub)
	}
	if !scope.All {
		switch {
		case scope.OwnerID != nil && scope.MemberUserID != nil:
			base = base.Where(
				"user_id = ? OR class_id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.OwnerID, *scope.MemberUserID,
			)
		case scope.OwnerID != nil:
			base = base.Where("user_id = ?", *scope.OwnerID)
		case scope.MemberUserID != nil:
			base = base.Where(
				"class_id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.MemberUserID,
			)
		}
	}
	var quizzes []domain.Quiz
	total, err := listparams.Paginate(base, p, &quizzes)
	if err != nil {
		return nil, 0, fmt.Errorf("quizzes.repository.List: %w", err)
	}
	return quizzes, total, nil
}

func (r *quizRepository) ListByMemberWithRooms(ctx context.Context, userID uuid.UUID, classID *uuid.UUID, p domain.ListParams) ([]domain.Quiz, error) {
	db := database.DB(ctx, r.db)
	base := db.Model(&domain.Quiz{}).
		Preload("Class").
		Where(
			"class_id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
			userID,
		)
	if classID != nil {
		base = base.Where("class_id = ?", *classID)
	}
	var quizzes []domain.Quiz
	if err := listparams.Apply(base, p).Find(&quizzes).Error; err != nil {
		return nil, fmt.Errorf("quizzes.repository.ListByMemberWithRooms: %w", err)
	}
	return quizzes, nil
}

func (r *quizRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.Quiz{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("quizzes.repository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *quizRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	var quiz domain.Quiz
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.Quiz{}).First(&quiz, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("quizzes.repository.FindByIDIncludingDeleted: %w", err)
	}
	return &quiz, nil
}

func (r *quizRepository) AdminList(ctx context.Context, q domain.AdminListQuizzesQuery) ([]domain.Quiz, int64, error) {
	db := database.DB(ctx, r.db)
	base := db.Model(&domain.Quiz{}).Preload("User").Preload("Class")
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.ClassID != nil {
		base = base.Where("class_id = ?", *q.ClassID)
	}
	if q.ClassSessionID != nil {
		sub := db.Table("quiz_rooms").
			Select("quiz_id").
			Where("class_session_id = ?", *q.ClassSessionID)
		base = base.Where("id IN (?)", sub)
	}
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	var quizzes []domain.Quiz
	total, err := listparams.Paginate(base, q.ListParams, &quizzes)
	if err != nil {
		return nil, 0, fmt.Errorf("quizzes.repository.AdminList: %w", err)
	}
	return quizzes, total, nil
}

// ruleRepository handles quiz_rules persistence.
type ruleRepository struct {
	db *gorm.DB
}

func NewRuleRepository(db *gorm.DB) domain.QuizRuleRepository {
	return &ruleRepository{db: db}
}

func (r *ruleRepository) Create(ctx context.Context, rule *domain.QuizRule) error {
	if err := database.DB(ctx, r.db).Create(rule).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("quizzes.ruleRepository.Create: %w", err)
	}
	return nil
}

func (r *ruleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuizRule, error) {
	var rule domain.QuizRule
	if err := database.DB(ctx, r.db).Model(&domain.QuizRule{}).First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("quizzes.ruleRepository.FindByID: %w", err)
	}
	return &rule, nil
}

func (r *ruleRepository) Update(ctx context.Context, rule *domain.QuizRule) error {
	result := database.DB(ctx, r.db).Save(rule)
	if result.Error != nil {
		return fmt.Errorf("quizzes.ruleRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ruleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.QuizRule{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("quizzes.ruleRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ruleRepository) ListByQuiz(ctx context.Context, quizID uuid.UUID, p domain.ListParams) ([]domain.QuizRule, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.QuizRule{}).Where("quiz_id = ?", quizID)
	var rules []domain.QuizRule
	total, err := listparams.Paginate(base, p, &rules)
	if err != nil {
		return nil, 0, fmt.Errorf("quizzes.ruleRepository.ListByQuiz: %w", err)
	}
	return rules, total, nil
}

// roomRepository handles quiz_rooms persistence.
type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) domain.QuizRoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(ctx context.Context, room *domain.QuizRoom) error {
	if err := database.DB(ctx, r.db).Create(room).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("quizzes.roomRepository.Create: %w", err)
	}
	return nil
}

func (r *roomRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	var room domain.QuizRoom
	if err := database.DB(ctx, r.db).Model(&domain.QuizRoom{}).First(&room, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("quizzes.roomRepository.FindByID: %w", err)
	}
	return &room, nil
}

func (r *roomRepository) Update(ctx context.Context, room *domain.QuizRoom) error {
	result := database.DB(ctx, r.db).Save(room)
	if result.Error != nil {
		return fmt.Errorf("quizzes.roomRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.QuizRoom{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("quizzes.roomRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) ListByQuiz(ctx context.Context, quizID uuid.UUID, p domain.ListParams) ([]domain.QuizRoom, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.QuizRoom{}).Where("quiz_id = ?", quizID)
	var rooms []domain.QuizRoom
	total, err := listparams.Paginate(base, p, &rooms)
	if err != nil {
		return nil, 0, fmt.Errorf("quizzes.roomRepository.ListByQuiz: %w", err)
	}
	return rooms, total, nil
}

func (r *roomRepository) ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]domain.QuizRoom, error) {
	var rooms []domain.QuizRoom
	if err := database.DB(ctx, r.db).Model(&domain.QuizRoom{}).
		Where("class_session_id = ?", sessionID).Find(&rooms).Error; err != nil {
		return nil, fmt.Errorf("quizzes.roomRepository.ListBySessionID: %w", err)
	}
	return rooms, nil
}

func (r *roomRepository) FindOpenByQuizID(ctx context.Context, quizID uuid.UUID) (*domain.QuizRoom, error) {
	var room domain.QuizRoom
	if err := database.DB(ctx, r.db).Model(&domain.QuizRoom{}).
		Where("quiz_id = ? AND started_at IS NOT NULL AND started_at <= NOW() AND (ended_at IS NULL OR ended_at > NOW())", quizID).
		First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("quizzes.roomRepository.FindOpenByQuizID: %w", err)
	}
	return &room, nil
}

// submissionRepository handles quiz_submissions persistence.
type submissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) domain.QuizSubmissionRepository {
	return &submissionRepository{db: db}
}

func (r *submissionRepository) Create(ctx context.Context, sub *domain.QuizSubmission) error {
	if err := database.DB(ctx, r.db).Create(sub).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("quizzes.submissionRepository.Create: %w", err)
	}
	return nil
}

func (r *submissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuizSubmission, error) {
	var sub domain.QuizSubmission
	if err := database.DB(ctx, r.db).Model(&domain.QuizSubmission{}).
		Preload("User").
		First(&sub, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("quizzes.submissionRepository.FindByID: %w", err)
	}
	return &sub, nil
}

func (r *submissionRepository) Update(ctx context.Context, sub *domain.QuizSubmission) error {
	result := database.DB(ctx, r.db).Save(sub)
	if result.Error != nil {
		return fmt.Errorf("quizzes.submissionRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *submissionRepository) FindByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*domain.QuizSubmission, error) {
	var sub domain.QuizSubmission
	if err := database.DB(ctx, r.db).Model(&domain.QuizSubmission{}).
		Where("quiz_id = ? AND user_id = ?", quizID, userID).
		First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("quizzes.submissionRepository.FindByQuizAndUser: %w", err)
	}
	return &sub, nil
}

func (r *submissionRepository) ListByQuiz(ctx context.Context, quizID uuid.UUID, q domain.ListSubmissionsQuery) ([]domain.QuizSubmission, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.QuizSubmission{}).
		Preload("User").
		Where("quiz_id = ?", quizID)
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	var subs []domain.QuizSubmission
	total, err := listparams.Paginate(base, q.ListParams, &subs)
	if err != nil {
		return nil, 0, fmt.Errorf("quizzes.submissionRepository.ListByQuiz: %w", err)
	}
	return subs, total, nil
}
