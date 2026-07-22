package offlines

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/queue"
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
	queue    *queue.Client
	tx       domain.Transactor
	audit    domain.AuditRecorder
	logger   *slog.Logger
}

func NewService(
	rooms domain.OfflineRoomRepository,
	views domain.OfflineRoomViewRepository,
	sessions domain.ClassSessionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	queueClient *queue.Client,
	tx domain.Transactor,
	audit domain.AuditRecorder,
	logger *slog.Logger,
) domain.OfflineService {
	return &service{
		rooms:    rooms,
		views:    views,
		sessions: sessions,
		classes:  classes,
		members:  members,
		queue:    queueClient,
		tx:       tx,
		audit:    audit,
		logger:   logger,
	}
}

// enqueueAttachmentCleanup schedules a purge of ALL media (rows + S3 objects)
// owned by an offline room — teacher-uploaded attachments keyed by room id. A
// hard delete removes the room row for good, so without this every object is
// orphaned (media is polymorphic, no FK to the room, so no cascade reaches it).
// An empty collection name matches every collection. Best-effort: a failure to
// enqueue is logged, not surfaced, so the delete still succeeds.
func (s *service) enqueueAttachmentCleanup(ctx context.Context, roomID uuid.UUID) {
	if s.queue == nil {
		return
	}
	payload, err := json.Marshal(domain.MediaCleanupPayload{
		ModelType: domain.MediaModelOfflineRoom,
		ModelID:   roomID,
	})
	if err != nil {
		s.logger.Error("attachment cleanup enqueue: marshal payload", "room_id", roomID.String(), "error", err)
		return
	}
	if _, err := s.queue.Enqueue(asynq.NewTask(domain.TypeMediaCleanup, payload), asynq.Queue(domain.QueueMedia)); err != nil {
		s.logger.Error("attachment cleanup enqueue", "room_id", roomID.String(), "error", err)
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
	isMember, err := s.members.Exists(ctx, room.ClassID, caller.UserID)
	if err != nil {
		return false, err
	}
	if !isMember {
		return false, nil
	}
	// Members see a room only once it is published.
	return room.PublishedAt != nil && !room.PublishedAt.After(time.Now()), nil
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
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.rooms.Create(ctx, room); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditCreated,
			TargetType:  domain.AuditTargetOffline,
			TargetID:    &room.ID,
			TargetLabel: room.Title,
			OrgID:       &room.OrganizationID,
			Metadata:    map[string]any{"class_id": room.ClassID.String()},
		})
	})
	if err != nil {
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
	// Shallow changed-fields diff (from/to) captured before mutating so the
	// audit entry records exactly what this update altered.
	changed := map[string]any{}
	setChanged := func(key string, from, to any) {
		if from != to {
			changed[key] = map[string]any{"from": from, "to": to}
		}
	}
	if dto.Title != nil {
		setChanged("title", room.Title, *dto.Title)
		room.Title = *dto.Title
	}
	if dto.Description != nil {
		setChanged("description", room.Description, *dto.Description)
		room.Description = *dto.Description
	}
	if dto.PublishedAt != nil {
		room.PublishedAt = dto.PublishedAt
		changed["published_at"] = dto.PublishedAt
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.rooms.Update(ctx, room); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetOffline,
			TargetID:    &room.ID,
			TargetLabel: room.Title,
			OrgID:       &room.OrganizationID,
			Metadata:    map[string]any{"changed": changed},
		})
	})
	if err != nil {
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
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.rooms.Delete(ctx, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditDeleted,
			TargetType:  domain.AuditTargetOffline,
			TargetID:    &id,
			TargetLabel: room.Title,
			OrgID:       &room.OrganizationID,
			Metadata:    map[string]any{"class_id": room.ClassID.String()},
		})
	})
	if err != nil {
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
