package classes

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// service implements domain.ClassService. RBAC hierarchy:
//
//	super-admin (caller.IsAdmin): full access
//	classes:update_any permission: full access within org
//	teacher     (class.UserID == caller.UserID): manage own class + sessions
//	student     (enrolled via class_members): view + self-leave only
//
// Authorization always happens in the service layer so handlers stay thin.
type service struct {
	repo       domain.ClassRepository
	sessions   domain.ClassSessionRepository
	members    domain.ClassMemberRepository
	logger     *slog.Logger
}

func NewService(
	repo domain.ClassRepository,
	sessions domain.ClassSessionRepository,
	members domain.ClassMemberRepository,
	logger *slog.Logger,
) domain.ClassService {
	return &service{repo: repo, sessions: sessions, members: members, logger: logger}
}

// canManageClass returns true if caller can mutate the given class (update,
// add sessions, enroll others). Students never qualify here.
func canManageClass(caller domain.Caller, class *domain.Class) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermClassesUpdateAny) {
		return true
	}
	return caller.UserID == class.UserID
}

func canDeleteClass(caller domain.Caller, class *domain.Class) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermClassesDeleteAny) {
		return true
	}
	return caller.UserID == class.UserID
}

// canViewClass returns true if caller can read the class. Admins/staff
// bypass; teachers view own; students view classes they're enrolled in.
func (s *service) canViewClass(ctx context.Context, caller domain.Caller, class *domain.Class) (bool, error) {
	if canManageClass(caller, class) {
		return true, nil
	}
	if caller.HasPermission(domain.PermClassesViewAny) {
		return true, nil
	}
	ok, err := s.members.Exists(ctx, class.ID, caller.UserID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *service) Create(ctx context.Context, dto domain.CreateClassDTO) (*domain.Class, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermClassesCreate) && !caller.HasPermission(domain.PermClassesCreateAny) {
		return nil, domain.ErrForbidden
	}
	if caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	userID := caller.UserID
	if dto.UserID != nil && (caller.IsAdmin || caller.HasPermission(domain.PermClassesCreateAny)) {
		userID = *dto.UserID
	}
	class := &domain.Class{
		OrganizationID: *caller.OrgID,
		UserID:         userID,
		Name:           dto.Name,
		Description:    dto.Description,
		TotalUsers:     dto.TotalUsers,
	}
	if err := s.repo.Create(ctx, class); err != nil {
		return nil, err
	}
	s.logger.Info("class created",
		"class_id", class.ID.String(),
		"created_by", caller.UserID.String(),
	)
	return class, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ok, err = s.canViewClass(ctx, caller, class)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}
	return class, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateClassDTO) (*domain.Class, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageClass(caller, class) {
		return nil, domain.ErrForbidden
	}
	if dto.Name != nil {
		class.Name = *dto.Name
	}
	if dto.Description != nil {
		class.Description = *dto.Description
	}
	if dto.TotalUsers != nil {
		class.TotalUsers = *dto.TotalUsers
	}
	if dto.UserID != nil && (caller.IsAdmin || caller.HasPermission(domain.PermClassesUpdateAny)) {
		class.UserID = *dto.UserID
	}
	if err := s.repo.Update(ctx, class); err != nil {
		return nil, err
	}
	return class, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteClass(caller, class) {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("class deleted",
		"class_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) List(ctx context.Context, p domain.ListParams) ([]domain.Class, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	return s.repo.List(ctx, scope, p)
}

// resolveListScope maps a Caller into the role-resolved ClassListScope the
// repository understands. The four modes from the domain doc correspond to
// the four RBAC tiers exactly.
func (s *service) resolveListScope(caller domain.Caller) domain.ClassListScope {
	if caller.IsAdmin {
		return domain.ClassListScope{All: true}
	}
	if caller.HasPermission(domain.PermClassesViewAny) || caller.HasPermission(domain.PermClassesUpdateAny) {
		return domain.ClassListScope{All: true, OrganizationID: caller.OrgID}
	}
	userID := caller.UserID
	return domain.ClassListScope{
		TeacherID:    &userID,
		MemberUserID: &userID,
	}
}

// --- sessions ---

func (s *service) CreateSession(ctx context.Context, classID uuid.UUID, dto domain.CreateClassSessionDTO) (*domain.ClassSession, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	if !canManageClass(caller, class) {
		return nil, domain.ErrForbidden
	}
	session := &domain.ClassSession{
		ClassID:      classID,
		Name:         dto.Name,
		Description:  dto.Description,
		StartTime:    dto.StartTime,
		Type:         dto.Type,
		IsRecordable: dto.IsRecordable,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *service) GetSession(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	session, err := s.sessions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	class, err := s.repo.FindByID(ctx, session.ClassID)
	if err != nil {
		return nil, err
	}
	visible, err := s.canViewClass(ctx, caller, class)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, domain.ErrForbidden
	}
	return session, nil
}

func (s *service) UpdateSession(ctx context.Context, id uuid.UUID, dto domain.UpdateClassSessionDTO) (*domain.ClassSession, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	session, err := s.sessions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	class, err := s.repo.FindByID(ctx, session.ClassID)
	if err != nil {
		return nil, err
	}
	if !canManageClass(caller, class) {
		return nil, domain.ErrForbidden
	}
	if dto.Name != nil {
		session.Name = *dto.Name
	}
	if dto.Description != nil {
		session.Description = *dto.Description
	}
	if dto.StartTime != nil {
		session.StartTime = *dto.StartTime
	}
	if dto.Type != nil {
		session.Type = *dto.Type
	}
	if dto.IsRecordable != nil {
		session.IsRecordable = *dto.IsRecordable
	}
	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *service) DeleteSession(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	session, err := s.sessions.FindByID(ctx, id)
	if err != nil {
		return err
	}
	class, err := s.repo.FindByID(ctx, session.ClassID)
	if err != nil {
		return err
	}
	if !canDeleteClass(caller, class) {
		return domain.ErrForbidden
	}
	return s.sessions.Delete(ctx, id)
}

func (s *service) ListSessions(ctx context.Context, classID uuid.UUID, q domain.ListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, classID)
	if err != nil {
		return nil, 0, err
	}
	visible, err := s.canViewClass(ctx, caller, class)
	if err != nil {
		return nil, 0, err
	}
	if !visible {
		return nil, 0, domain.ErrForbidden
	}
	// Students may not request soft-deleted rows.
	if !canManageClass(caller, class) {
		q.IncludeDeleted = false
	}
	return s.sessions.ListByClass(ctx, classID, q)
}

// --- membership ---

func (s *service) Enroll(ctx context.Context, classID uuid.UUID, dto domain.EnrollClassMemberDTO) (*domain.ClassMember, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	// Authorization: teacher/staff/admin may enroll any user. A student may
	// only self-enroll.
	if !canManageClass(caller, class) && dto.UserID != caller.UserID {
		return nil, domain.ErrForbidden
	}
	// Capacity check: TotalUsers == 0 means unlimited. Otherwise current
	// member count must be strictly less than capacity.
	if class.TotalUsers > 0 {
		count, err := s.members.CountByClass(ctx, classID)
		if err != nil {
			return nil, err
		}
		if count >= int64(class.TotalUsers) {
			return nil, domain.ErrConflict
		}
	}
	m := &domain.ClassMember{ClassID: classID, UserID: dto.UserID}
	if err := s.members.Create(ctx, m); err != nil {
		return nil, err
	}
	s.logger.Info("class enrollment",
		"class_id", classID.String(),
		"user_id", dto.UserID.String(),
		"enrolled_by", caller.UserID.String(),
	)
	return m, nil
}

func (s *service) Leave(ctx context.Context, classID, userID uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, classID)
	if err != nil {
		return err
	}
	// Self-leave always allowed; teachers/staff/admins may unenroll others.
	if userID != caller.UserID && !canManageClass(caller, class) {
		return domain.ErrForbidden
	}
	return s.members.Delete(ctx, classID, userID)
}

func (s *service) ListMembers(ctx context.Context, classID uuid.UUID, q domain.ListClassMembersQuery) ([]domain.ClassMember, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, classID)
	if err != nil {
		return nil, 0, err
	}
	// Roster is considered sensitive: only managers see it. Students can see
	// their own enrollment status via GetByID / List but not the full roster.
	if !canManageClass(caller, class) {
		return nil, 0, domain.ErrForbidden
	}
	return s.members.ListByClass(ctx, classID, q.ListParams)
}
