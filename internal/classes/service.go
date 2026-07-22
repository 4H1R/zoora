package classes

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/authz"
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
	repo     domain.ClassRepository
	sessions domain.ClassSessionRepository
	members  domain.ClassMemberRepository
	chat     domain.ClassChatProvisioner // may be nil (chat provisioning disabled)
	tx       domain.Transactor
	audit    domain.AuditRecorder
	logger   *slog.Logger
}

func NewService(
	repo domain.ClassRepository,
	sessions domain.ClassSessionRepository,
	members domain.ClassMemberRepository,
	chat domain.ClassChatProvisioner,
	tx domain.Transactor,
	audit domain.AuditRecorder,
	logger *slog.Logger,
) domain.ClassService {
	return &service{repo: repo, sessions: sessions, members: members, chat: chat, tx: tx, audit: audit, logger: logger}
}

// canManageClass returns true if caller can mutate the given class (update,
// add sessions, enroll others). Students never qualify here.
func canManageClass(caller domain.Caller, class *domain.Class) bool {
	return caller.CanManage(class.UserID, domain.PermClassesUpdateAny)
}

func canDeleteClass(caller domain.Caller, class *domain.Class) bool {
	return caller.CanManage(class.UserID, domain.PermClassesDeleteAny)
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
	err := s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, class); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditCreated,
			TargetType:  domain.AuditTargetClass,
			TargetID:    &class.ID,
			TargetLabel: class.Name,
			OrgID:       &class.OrganizationID,
		})
	})
	if err != nil {
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
	// Build a shallow changed-fields diff (from/to) before mutating so the audit
	// entry records exactly what this update altered.
	changed := map[string]any{}
	if dto.Name != nil && *dto.Name != class.Name {
		changed["name"] = map[string]any{"from": class.Name, "to": *dto.Name}
		class.Name = *dto.Name
	}
	if dto.Description != nil && *dto.Description != class.Description {
		changed["description"] = map[string]any{"from": class.Description, "to": *dto.Description}
		class.Description = *dto.Description
	}
	if dto.TotalUsers != nil && *dto.TotalUsers != class.TotalUsers {
		changed["total_users"] = map[string]any{"from": class.TotalUsers, "to": *dto.TotalUsers}
		class.TotalUsers = *dto.TotalUsers
	}
	if dto.UserID != nil && (caller.IsAdmin || caller.HasPermission(domain.PermClassesUpdateAny)) && *dto.UserID != class.UserID {
		changed["user_id"] = map[string]any{"from": class.UserID.String(), "to": dto.UserID.String()}
		class.UserID = *dto.UserID
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, class); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetClass,
			TargetID:    &class.ID,
			TargetLabel: class.Name,
			OrgID:       &class.OrganizationID,
			Metadata:    map[string]any{"changed": changed},
		})
	})
	if err != nil {
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
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		// Cheap cascade count for forensic metadata: how many enrollments the
		// delete takes down with it. Best-effort — a count error must not block
		// the delete, so it degrades to zero.
		memberCount, _ := s.members.CountByClass(ctx, id)
		if err := s.repo.Delete(ctx, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditDeleted,
			TargetType:  domain.AuditTargetClass,
			TargetID:    &id,
			TargetLabel: class.Name,
			OrgID:       &class.OrganizationID,
			Metadata:    map[string]any{"cascaded": map[string]any{"enrollments": memberCount}},
		})
	})
	if err != nil {
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
// repository understands, via the shared authz.ListScope resolver (the four
// RBAC tiers: admin / org-wide _any / teacher / enrolled member).
func (s *service) resolveListScope(caller domain.Caller) domain.ClassListScope {
	return authz.ListScope(caller, domain.PermClassesViewAny, domain.PermClassesUpdateAny)
}

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
		ClassID:     classID,
		Name:        dto.Name,
		Description: dto.Description,
		StartTime:   dto.StartTime,
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

func (s *service) Enroll(ctx context.Context, classID uuid.UUID, dto domain.EnrollClassMemberDTO) (*domain.ClassMember, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.repo.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	// Tenant boundary: a non-admin may only enroll into a class that belongs to
	// their own org. This runs before both the manage and self-enroll branches,
	// so neither can cross tenants. Admins may enroll into any org.
	if !caller.IsAdmin {
		if caller.OrgID == nil || class.OrganizationID != *caller.OrgID {
			return nil, domain.ErrForbidden
		}
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
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.members.Create(ctx, m); err != nil {
			return err
		}
		// No user repo here, so the class name stands in as the human label; the
		// enrolled user is captured in metadata.
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditEnrolled,
			TargetType:  domain.AuditTargetEnrollment,
			TargetID:    &m.ID,
			TargetLabel: class.Name,
			OrgID:       &class.OrganizationID,
			Metadata:    map[string]any{"class_id": classID.String(), "user_id": dto.UserID.String()},
		})
	})
	if err != nil {
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
	return s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.members.Delete(ctx, classID, userID); err != nil {
			return err
		}
		// No membership row is loaded (delete is by class+user), so the class name
		// is the label and the unenrolled user is captured in metadata.
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUnenrolled,
			TargetType:  domain.AuditTargetEnrollment,
			TargetLabel: class.Name,
			OrgID:       &class.OrganizationID,
			Metadata:    map[string]any{"class_id": classID.String(), "user_id": userID.String()},
		})
	})
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

// ProvisionConversation creates or syncs the class's group/channel chat. Only a
// class manager (owning teacher, org-wide _any holder, or super-admin) may do
// this — class ownership stands in for conversations:manage. The org must have
// the chat feature. Members are seeded/synced from the current roster: the
// teacher (Class.UserID) as admin and every enrolled student as member.
func (s *service) ProvisionConversation(ctx context.Context, classID uuid.UUID, dto domain.ProvisionClassConversationDTO) (*domain.Conversation, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if s.chat == nil {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureChat)
	}
	class, err := s.repo.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	if !canManageClass(caller, class) {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasFeature(domain.FeatureChat) {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureChat)
	}

	// Roster = every enrolled student. The teacher is added separately as admin
	// by the provisioner (CreatorID), so it is not part of the member list here.
	roster, err := s.members.ListAllByClass(ctx, classID)
	if err != nil {
		return nil, err
	}
	memberIDs := make([]uuid.UUID, 0, len(roster)+1)
	for _, m := range roster {
		memberIDs = append(memberIDs, m.UserID)
	}
	// A manager who is not the class's teacher would otherwise not belong to the
	// conversation they just provisioned. Add them as a member so they can see and
	// moderate it. The teacher is skipped here because CreatorID already seats them
	// as admin; an enrolled-student caller is deduped by the create/sync paths.
	if caller.UserID != class.UserID {
		memberIDs = append(memberIDs, caller.UserID)
	}

	// Existing link → additive member sync. If the linked conversation no longer
	// exists (there is no FK to null a stale link), fall through to recreate it.
	if class.ConversationID != nil {
		conv, err := s.chat.SyncClassMembers(ctx, *class.ConversationID, memberIDs)
		if err == nil {
			return conv, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		class.ConversationID = nil
	}

	name := dto.Name
	if name == "" {
		name = class.Name
	}
	conv, err := s.chat.CreateForClass(ctx, domain.ProvisionClassChatDTO{
		OrganizationID: class.OrganizationID,
		CreatorID:      class.UserID,
		Type:           dto.Type,
		Name:           name,
		ColorIndex:     dto.ColorIndex,
		MemberIDs:      memberIDs,
	})
	if err != nil {
		return nil, err
	}
	class.ConversationID = &conv.ID
	if err := s.repo.Update(ctx, class); err != nil {
		return nil, err
	}
	s.logger.Info("class conversation provisioned",
		"class_id", classID.String(),
		"conversation_id", conv.ID.String(),
		"by", caller.UserID.String(),
	)
	return conv, nil
}
