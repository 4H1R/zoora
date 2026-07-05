package offlines

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// service implements domain.OfflineService. RBAC hierarchy:
//
//	super-admin (caller.IsAdmin): full access
//	offlines:update_any permission: full access within org
//	creator     (room.CreatorID == caller.UserID): manage own room
//	member      (enrolled via class_members of room.ClassID): view only
//
// Authorization always happens in the service layer so handlers stay thin.
type service struct {
	rooms    domain.OfflineRoomRepository
	views    domain.OfflineRoomViewRepository
	sessions domain.ClassSessionRepository
	classes  domain.ClassRepository
	members  domain.ClassMemberRepository
	logger   *slog.Logger
}

func NewService(
	rooms domain.OfflineRoomRepository,
	views domain.OfflineRoomViewRepository,
	sessions domain.ClassSessionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	logger *slog.Logger,
) domain.OfflineService {
	return &service{
		rooms:    rooms,
		views:    views,
		sessions: sessions,
		classes:  classes,
		members:  members,
		logger:   logger,
	}
}

func canManageRoom(caller domain.Caller, room *domain.OfflineRoom) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermOfflinesUpdateAny) {
		return true
	}
	return caller.UserID == room.CreatorID
}

func canDeleteRoom(caller domain.Caller, room *domain.OfflineRoom) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermOfflinesDeleteAny) {
		return true
	}
	return caller.UserID == room.CreatorID
}

func (s *service) canViewRoom(ctx context.Context, caller domain.Caller, room *domain.OfflineRoom) (bool, error) {
	if canManageRoom(caller, room) {
		return true, nil
	}
	if caller.HasPermission(domain.PermOfflinesViewAny) {
		return true, nil
	}
	return s.members.Exists(ctx, room.ClassID, caller.UserID)
}

func (s *service) CreateRoom(ctx context.Context, dto domain.CreateOfflineRoomDTO) (*domain.OfflineRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	session, err := s.sessions.FindByID(ctx, dto.ClassSessionID)
	if err != nil {
		return nil, err
	}
	class, err := s.classes.FindByID(ctx, session.ClassID)
	if err != nil {
		return nil, err
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermOfflinesCreateAny) && caller.UserID != class.UserID {
		return nil, domain.ErrForbidden
	}
	if !caller.HasFeature(domain.FeatureOfflineRooms) {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureOfflineRooms)
	}
	room := &domain.OfflineRoom{
		OrganizationID: class.OrganizationID,
		ClassID:        class.ID,
		ClassSessionID: dto.ClassSessionID,
		CreatorID:      caller.UserID,
		Title:          dto.Title,
		Description:    dto.Description,
		PublishedAt:    dto.PublishedAt,
	}
	if err := s.rooms.Create(ctx, room); err != nil {
		return nil, err
	}
	s.logger.Info("offline room created",
		"room_id", room.ID.String(),
		"class_id", room.ClassID.String(),
		"created_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) GetRoom(ctx context.Context, id uuid.UUID) (*domain.OfflineRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	visible, err := s.canViewRoom(ctx, caller, room)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, domain.ErrForbidden
	}
	_ = s.rooms.IncrementViewCount(ctx, id)
	_ = s.views.Create(ctx, &domain.OfflineRoomView{
		OfflineRoomID: id,
		UserID:        caller.UserID,
	})
	return room, nil
}

func (s *service) UpdateRoom(ctx context.Context, id uuid.UUID, dto domain.UpdateOfflineRoomDTO) (*domain.OfflineRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageRoom(caller, room) {
		return nil, domain.ErrForbidden
	}
	if dto.Title != nil {
		room.Title = *dto.Title
	}
	if dto.Description != nil {
		room.Description = *dto.Description
	}
	if dto.PublishedAt != nil {
		room.PublishedAt = dto.PublishedAt
	}
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *service) DeleteRoom(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteRoom(caller, room) {
		return domain.ErrForbidden
	}
	if err := s.rooms.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("offline room deleted",
		"room_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) ListRooms(ctx context.Context, q domain.ListOfflineRoomsQuery) ([]domain.OfflineRoom, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	if !canListDeleted(caller) {
		q.IncludeDeleted = false
	}
	return s.rooms.List(ctx, scope, q)
}

// resolveListScope maps a Caller into a role-resolved OfflineRoomListScope.
// super-admin: All across orgs. ViewAny / UpdateAny in caller's org: All but
// constrained to caller.OrgID. Anyone else: their own + classes they're a
// member of.
func (s *service) resolveListScope(caller domain.Caller) domain.OfflineRoomListScope {
	if caller.IsAdmin {
		return domain.OfflineRoomListScope{All: true}
	}
	if caller.HasPermission(domain.PermOfflinesViewAny) || caller.HasPermission(domain.PermOfflinesUpdateAny) {
		return domain.OfflineRoomListScope{All: true, OrganizationID: caller.OrgID}
	}
	userID := caller.UserID
	return domain.OfflineRoomListScope{
		OwnerID:      &userID,
		MemberUserID: &userID,
	}
}

func canListDeleted(caller domain.Caller) bool {
	return caller.IsAdmin || caller.HasPermission(domain.PermOfflinesUpdateAny)
}
