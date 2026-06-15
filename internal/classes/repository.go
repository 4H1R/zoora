package classes

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

type classRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.Class{})
}

func (r *classRepository) Create(ctx context.Context, class *domain.Class) error {
	if err := database.DB(ctx, r.db).Create(class).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("classes.repository.Create: %w", err)
	}
	return nil
}

func (r *classRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	var class domain.Class
	if err := r.baseQuery(ctx).Preload("User").First(&class, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("classes.repository.FindByID: %w", err)
	}
	return &class, nil
}

func (r *classRepository) Update(ctx context.Context, class *domain.Class) error {
	result := database.DB(ctx, r.db).Save(class)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("classes.repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *classRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Class{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("classes.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List applies a role-resolved scope produced by the service. All short-
// circuits all scoping. TeacherID and MemberUserID are OR'd when both are
// set, so a teacher who is also enrolled elsewhere sees both sets in one query.
func (r *classRepository) List(ctx context.Context, scope domain.ClassListScope, p domain.ListParams) ([]domain.Class, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Class{}).Preload("User")
	if scope.IncludeDeleted {
		base = base.Unscoped()
	}
	if scope.OrganizationID != nil {
		base = base.Where("organization_id = ?", *scope.OrganizationID)
	}
	if !scope.All {
		switch {
		case scope.TeacherID != nil && scope.MemberUserID != nil:
			base = base.Where(
				"user_id = ? OR id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.TeacherID, *scope.MemberUserID,
			)
		case scope.TeacherID != nil:
			base = base.Where("user_id = ?", *scope.TeacherID)
		case scope.MemberUserID != nil:
			base = base.Where(
				"id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.MemberUserID,
			)
		}
	}
	var classes []domain.Class
	total, err := listparams.Paginate(base, p, &classes)
	if err != nil {
		return nil, 0, fmt.Errorf("classes.repository.List: %w", err)
	}
	return classes, total, nil
}

func (r *classRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.Class{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("classes.repository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *classRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	var class domain.Class
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.Class{}).First(&class, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("classes.repository.FindByIDIncludingDeleted: %w", err)
	}
	return &class, nil
}

func (r *classRepository) AdminList(ctx context.Context, q domain.AdminListClassesQuery) ([]domain.Class, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Class{}).Preload("User")
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	var classes []domain.Class
	total, err := listparams.Paginate(base, q.ListParams, &classes)
	if err != nil {
		return nil, 0, fmt.Errorf("classes.repository.AdminList: %w", err)
	}
	return classes, total, nil
}

// sessionRepository is a sibling persistence type for class_sessions. Kept in
// the same package because sessions are meaningless outside a class context.
type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) domain.ClassSessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.ClassSession{})
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.ClassSession) error {
	if err := database.DB(ctx, r.db).Create(session).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("classes.sessionRepository.Create: %w", err)
	}
	return nil
}

func (r *sessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	var s domain.ClassSession
	if err := r.baseQuery(ctx).First(&s, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("classes.sessionRepository.FindByID: %w", err)
	}
	return &s, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *domain.ClassSession) error {
	result := database.DB(ctx, r.db).Save(session)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("classes.sessionRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.ClassSession{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("classes.sessionRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *sessionRepository) ListByClass(ctx context.Context, classID uuid.UUID, q domain.ListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.ClassSession{})
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	base = base.Where("class_id = ?", classID)
	var sessions []domain.ClassSession
	total, err := listparams.Paginate(base, q.ListParams, &sessions)
	if err != nil {
		return nil, 0, fmt.Errorf("classes.sessionRepository.ListByClass: %w", err)
	}
	return sessions, total, nil
}

func (r *sessionRepository) AdminList(ctx context.Context, q domain.AdminListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.ClassSession{}).Preload("Class")
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.ClassID != nil {
		base = base.Where("class_id = ?", *q.ClassID)
	}
	var sessions []domain.ClassSession
	total, err := listparams.Paginate(base, q.ListParams, &sessions)
	if err != nil {
		return nil, 0, fmt.Errorf("classes.sessionRepository.AdminList: %w", err)
	}
	return sessions, total, nil
}

func (r *sessionRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.ClassSession{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("classes.sessionRepository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *sessionRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	var s domain.ClassSession
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.ClassSession{}).First(&s, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("classes.sessionRepository.FindByIDIncludingDeleted: %w", err)
	}
	return &s, nil
}

// memberRepository persists class_members rows. Keeping it in the classes
// package (rather than its own module) because membership is an aggregate
// child of Class and never queried independently of one.
type memberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) domain.ClassMemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(ctx context.Context, m *domain.ClassMember) error {
	if err := database.DB(ctx, r.db).Create(m).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("classes.memberRepository.Create: %w", err)
	}
	return nil
}

func (r *memberRepository) Delete(ctx context.Context, classID, userID uuid.UUID) error {
	result := database.DB(ctx, r.db).
		Where("class_id = ? AND user_id = ?", classID, userID).
		Delete(&domain.ClassMember{})
	if result.Error != nil {
		return fmt.Errorf("classes.memberRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *memberRepository) Exists(ctx context.Context, classID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := database.DB(ctx, r.db).Model(&domain.ClassMember{}).
		Where("class_id = ? AND user_id = ?", classID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("classes.memberRepository.Exists: %w", err)
	}
	return count > 0, nil
}

func (r *memberRepository) CountByClass(ctx context.Context, classID uuid.UUID) (int64, error) {
	var count int64
	if err := database.DB(ctx, r.db).Model(&domain.ClassMember{}).
		Where("class_id = ?", classID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("classes.memberRepository.CountByClass: %w", err)
	}
	return count, nil
}

func (r *memberRepository) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.ClassMember, error) {
	var members []domain.ClassMember
	if err := database.DB(ctx, r.db).Model(&domain.ClassMember{}).
		Preload("User").
		Where("class_id = ?", classID).
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("classes.memberRepository.ListAllByClass: %w", err)
	}
	return members, nil
}

func (r *memberRepository) ListByClass(ctx context.Context, classID uuid.UUID, p domain.ListParams) ([]domain.ClassMember, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.ClassMember{}).
		Preload("User").
		Where("class_id = ?", classID)
	var members []domain.ClassMember
	total, err := listparams.Paginate(base, p, &members)
	if err != nil {
		return nil, 0, fmt.Errorf("classes.memberRepository.ListByClass: %w", err)
	}
	return members, total, nil
}
