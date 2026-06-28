package livesessions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	lkproto "github.com/livekit/protocol/livekit"

	"github.com/4H1R/zoora/internal/domain"
	lk "github.com/4H1R/zoora/internal/platform/livekit"
	"github.com/4H1R/zoora/internal/platform/queue"
)

// liveKitClient is a local interface so the service can be tested with a nil or
// mock client. *lk.Client satisfies it automatically.
type liveKitClient interface {
	CreateRoom(ctx context.Context, roomName string, maxParticipants uint32) (*lkproto.Room, error)
	DeleteRoom(ctx context.Context, roomName string) error
	GenerateToken(roomName, identity, name, metadata string, sources []lkproto.TrackSource, roomAdmin bool) (string, error)
	StartRecording(ctx context.Context, roomName, s3Path string) (string, error)
	StopRecording(ctx context.Context, egressID string) error
	ListParticipants(ctx context.Context, roomName string) ([]*lkproto.ParticipantInfo, error)
	UpdateParticipant(ctx context.Context, roomName, identity, metadata string, sources []lkproto.TrackSource) error
	MutePublishedTrack(ctx context.Context, roomName, identity, trackSID string, muted bool) error
	SendData(ctx context.Context, roomName string, payload []byte, destinationIdentities []string) error
	PublicURL() string
}

type service struct {
	rooms        domain.LiveRoomRepository
	participants domain.LiveParticipantRepository
	recordings   domain.LiveRecordingRepository
	whiteboards  domain.LiveWhiteboardRepository
	sessions     domain.ClassSessionRepository
	classes      domain.ClassRepository
	members      domain.ClassMemberRepository
	chatSvc      domain.ChatService
	tx           domain.Transactor
	livekit      liveKitClient
	queue        *queue.Client
	logger       *slog.Logger
}

func NewService(
	rooms domain.LiveRoomRepository,
	participants domain.LiveParticipantRepository,
	recordings domain.LiveRecordingRepository,
	whiteboards domain.LiveWhiteboardRepository,
	sessions domain.ClassSessionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	chatSvc domain.ChatService,
	tx domain.Transactor,
	livekit *lk.Client,
	queueClient *queue.Client,
	logger *slog.Logger,
) domain.LiveSessionService {
	// Avoid storing a typed nil in the liveKitClient interface — a typed nil
	// satisfies the interface (non-nil interface) and bypasses the s.livekit != nil
	// guards used throughout the service.
	var lkClient liveKitClient
	if livekit != nil {
		lkClient = livekit
	}
	return &service{
		rooms:        rooms,
		participants: participants,
		recordings:   recordings,
		whiteboards:  whiteboards,
		sessions:     sessions,
		classes:      classes,
		members:      members,
		chatSvc:      chatSvc,
		tx:           tx,
		livekit:      lkClient,
		queue:        queueClient,
		logger:       logger,
	}
}

func (s *service) canUpdateRoom(caller domain.Caller, class *domain.Class) bool {
	return caller.CanManageOwned(class.UserID, domain.PermLiveSessionsUpdate, domain.PermLiveSessionsUpdateAny)
}

func (s *service) canManageRoom(caller domain.Caller, class *domain.Class) bool {
	return caller.CanManageOwned(class.UserID, domain.PermLiveSessionsManage, domain.PermLiveSessionsManageAny)
}

func (s *service) canViewRoom(ctx context.Context, caller domain.Caller, class *domain.Class) (bool, error) {
	if caller.IsAdmin || caller.HasPermission(domain.PermLiveSessionsViewAny) {
		return true, nil
	}
	if s.canManageRoom(caller, class) {
		return true, nil
	}
	ok, err := s.members.Exists(ctx, class.ID, caller.UserID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// loadRoomWithClass fetches room, its class session, and owning class.
func (s *service) loadRoomWithClass(ctx context.Context, roomID uuid.UUID) (*domain.LiveRoom, *domain.ClassSession, *domain.Class, error) {
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, nil, nil, err
	}
	session, err := s.sessions.FindByID(ctx, room.ClassSessionID)
	if err != nil {
		return nil, nil, nil, err
	}
	class, err := s.classes.FindByID(ctx, session.ClassID)
	if err != nil {
		return nil, nil, nil, err
	}
	return room, session, class, nil
}

// resolveListScope maps a Caller into the role-resolved LiveRoomListScope the
// repository understands. Typed filters from the request query are layered on
// top by the caller, not here.
func (s *service) resolveListScope(caller domain.Caller) domain.LiveRoomListScope {
	if caller.IsAdmin {
		return domain.LiveRoomListScope{All: true}
	}
	if caller.HasPermission(domain.PermLiveSessionsViewAny) {
		return domain.LiveRoomListScope{All: true, OrganizationID: caller.OrgID}
	}
	userID := caller.UserID
	return domain.LiveRoomListScope{
		TeacherID:    &userID,
		MemberUserID: &userID,
	}
}

func (s *service) CreateRoom(ctx context.Context, dto domain.CreateLiveRoomDTO) (*domain.LiveRoom, error) {
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
	if !s.canManageRoom(caller, class) {
		return nil, domain.ErrForbidden
	}

	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return nil, domain.NewValidationError(map[string]string{"name": "required"})
	}

	cfg := dto.Config
	if cfg.MaxParticipants <= 0 {
		// Backfill only the participant cap; preserve caller-supplied flags.
		cfg.MaxParticipants = domain.DefaultLiveRoomConfig().MaxParticipants
	}

	roomID := uuid.New()
	room := &domain.LiveRoom{
		ID:                 roomID,
		ClassSessionID:     dto.ClassSessionID,
		Name:               name,
		LiveKitRoomName:    fmt.Sprintf("session-%s-%s", dto.ClassSessionID.String(), roomID.String()),
		Status:             domain.LiveRoomStatusCreated,
		ScheduledStartTime: dto.ScheduledStartTime,
		Config:             cfg,
	}

	err = s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.rooms.Create(txCtx, room); err != nil {
			return err
		}

		chatName := fmt.Sprintf("Chat – %s", session.Name)
		_, cErr := s.chatSvc.CreateChat(txCtx, domain.CreateChatDTO{
			Name:      chatName,
			ModelType: "live_session",
			ModelID:   room.ID.String(),
		})
		return cErr
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("live room created",
		"room_id", room.ID.String(),
		"class_session_id", dto.ClassSessionID.String(),
		"created_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) GetRoom(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, session, class, err := s.loadRoomWithClass(ctx, id)
	if err != nil {
		return nil, err
	}
	ok, err = s.canViewRoom(ctx, caller, class)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	session.Class = class
	room.ClassSession = session

	return room, nil
}

func (s *service) JoinRoom(ctx context.Context, roomID uuid.UUID) (*domain.JoinLiveRoomResponse, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.Status == domain.LiveRoomStatusFinished {
		return nil, domain.ErrForbidden
	}

	isManager := s.canManageRoom(caller, class)
	hasJoinAny := caller.HasPermission(domain.PermLiveSessionsJoinAny)
	if !isManager && !hasJoinAny {
		enrolled, err := s.members.Exists(ctx, class.ID, caller.UserID)
		if err != nil {
			return nil, err
		}
		if !enrolled {
			return nil, domain.ErrForbidden
		}
	}

	isModerator := isManager || hasJoinAny

	// A scheduled room sits in "created" until someone starts it (the prejoin
	// lobby calls join, never start). The host joining auto-starts it so the
	// LiveKit room actually exists; everyone else must wait for the host.
	if room.Status == domain.LiveRoomStatusCreated {
		if isModerator {
			if _, err := s.startRoomInternal(ctx, room); err != nil {
				return nil, err
			}
		} else {
			return nil, domain.NewValidationError(map[string]string{"status": "room not started yet"})
		}
	}

	// Defensive: guarantee the LiveKit room exists before issuing a token
	// (idempotent — returns the existing room if already created).
	if _, err := s.livekit.CreateRoom(ctx, room.LiveKitRoomName, uint32(room.Config.MaxParticipants)); err != nil {
		return nil, fmt.Errorf("livesessions.service.JoinRoom livekit: %w", err)
	}

	sources := publishSources(isModerator)
	roomAdmin := isModerator

	identity := caller.UserID.String()
	displayName := caller.Name
	if displayName == "" {
		displayName = identity
	}

	role := domain.ParticipantRoleViewer
	if isModerator {
		role = domain.ParticipantRoleHost
	}

	token, err := s.livekit.GenerateToken(room.LiveKitRoomName, identity, displayName, participantMetadata(role), sources, roomAdmin)
	if err != nil {
		return nil, fmt.Errorf("livesessions.service.JoinRoom token: %w", err)
	}
	participant := &domain.LiveParticipant{
		LiveRoomID: roomID,
		UserID:     caller.UserID,
		Identity:   identity,
		Role:       role,
		JoinedAt:   time.Now(),
	}
	if err := s.participants.Create(ctx, participant); err != nil {
		return nil, err
	}

	s.logger.Info("user joined live room",
		"room_id", roomID.String(),
		"user_id", caller.UserID.String(),
		"is_manager", isManager,
	)

	resp := &domain.JoinLiveRoomResponse{
		Token:      token,
		LiveKitURL: s.livekit.PublicURL(),
		Room:       room,
	}

	if chat, err := s.chatSvc.FindChatByModel(ctx, "live_session", roomID); err == nil {
		resp.ChatID = &chat.ID
	}

	return resp, nil
}

func (s *service) LeaveRoom(ctx context.Context, roomID uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	p, err := s.participants.FindActiveByRoomAndUser(ctx, roomID, caller.UserID)
	if err != nil {
		return err
	}
	now := time.Now()
	p.LeftAt = &now
	p.TotalDurationSeconds = int(now.Sub(p.JoinedAt).Seconds())
	if err := s.participants.Update(ctx, p); err != nil {
		return err
	}
	s.logger.Info("user left live room",
		"room_id", roomID.String(),
		"user_id", caller.UserID.String(),
	)
	return nil
}

func (s *service) StartRoom(ctx context.Context, roomID uuid.UUID) (*domain.LiveRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !s.canManageRoom(caller, class) {
		return nil, domain.ErrForbidden
	}
	if room.Status != domain.LiveRoomStatusCreated {
		return nil, domain.NewValidationError(map[string]string{"status": "room must be in created state to start"})
	}

	if _, err := s.startRoomInternal(ctx, room); err != nil {
		return nil, err
	}

	s.logger.Info("live room started",
		"room_id", room.ID.String(),
		"started_by", caller.UserID.String(),
	)
	return room, nil
}

// startRoomInternal creates the LiveKit room, promotes the DB row to active and
// kicks off auto-recording. Callers must have already verified manage rights and
// that the room is in the created state. Shared by StartRoom and JoinRoom (host
// auto-start) so the two paths can never drift.
func (s *service) startRoomInternal(ctx context.Context, room *domain.LiveRoom) (*domain.LiveRoom, error) {
	if _, err := s.livekit.CreateRoom(ctx, room.LiveKitRoomName, uint32(room.Config.MaxParticipants)); err != nil {
		return nil, fmt.Errorf("livesessions.service.startRoomInternal livekit: %w", err)
	}

	now := time.Now()
	room.Status = domain.LiveRoomStatusActive
	room.ActualStartTime = &now
	room.HostLastSeenAt = &now
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// publishSources maps a caller's role into the LiveKit track sources they may
// publish. Moderators get everything; everyone else is subscribe-only (view).
// The room is a conference room: students don't publish until granted access.
func publishSources(isModerator bool) []lkproto.TrackSource {
	if isModerator {
		return []lkproto.TrackSource{
			lkproto.TrackSource_CAMERA,
			lkproto.TrackSource_MICROPHONE,
			lkproto.TrackSource_SCREEN_SHARE,
			lkproto.TrackSource_SCREEN_SHARE_AUDIO,
		}
	}
	return nil
}

func (s *service) endRoomInternal(ctx context.Context, room *domain.LiveRoom) (*domain.LiveRoom, error) {
	rec, err := s.recordings.FindActiveByRoom(ctx, room.ID)
	if err == nil && s.livekit != nil {
		_ = s.livekit.StopRecording(ctx, rec.EgressID)
		now := time.Now()
		rec.Status = domain.LiveRecordingStatusCompleted
		rec.EndedAt = &now
		_ = s.recordings.Update(ctx, rec)
	}

	now := time.Now()
	room.Status = domain.LiveRoomStatusFinished
	room.ActualEndTime = &now
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}

	_ = s.participants.MarkAllLeft(ctx, room.ID, now)

	if s.livekit != nil {
		_ = s.livekit.DeleteRoom(ctx, room.LiveKitRoomName)
	}

	if err := s.chatSvc.ArchiveByModel(ctx, "live_session", room.ID); err != nil {
		s.logger.Error("failed to archive chat for room", "room_id", room.ID.String(), "error", err)
	}

	s.enqueueAutoMark(ctx, room)

	return room, nil
}

// enqueueAutoMark schedules a session-scoped attendance auto-mark for the room's
// session. Best-effort: failures are logged and never block room teardown.
func (s *service) enqueueAutoMark(ctx context.Context, room *domain.LiveRoom) {
	if s.queue == nil {
		return
	}
	session, err := s.sessions.FindByID(ctx, room.ClassSessionID)
	if err != nil {
		s.logger.Error("auto-mark enqueue: load session", "room_id", room.ID.String(), "error", err)
		return
	}
	payload, err := json.Marshal(domain.AttendanceAutoMarkPayload{
		ClassID:   session.ClassID,
		SessionID: session.ID,
	})
	if err != nil {
		s.logger.Error("auto-mark enqueue: marshal payload", "room_id", room.ID.String(), "error", err)
		return
	}
	if _, err := s.queue.Enqueue(asynq.NewTask(domain.TypeAttendanceAutoMark, payload)); err != nil {
		s.logger.Error("auto-mark enqueue", "room_id", room.ID.String(), "error", err)
	}
}

func (s *service) EndRoom(ctx context.Context, roomID uuid.UUID) (*domain.LiveRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !s.canManageRoom(caller, class) {
		return nil, domain.ErrForbidden
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil, domain.NewValidationError(map[string]string{"status": "room must be active to end"})
	}

	room, err = s.endRoomInternal(ctx, room)
	if err != nil {
		return nil, err
	}

	s.logger.Info("live room ended",
		"room_id", room.ID.String(),
		"ended_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) UpdateRoomConfig(ctx context.Context, roomID uuid.UUID, dto domain.UpdateLiveRoomConfigDTO) (*domain.LiveRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !s.canUpdateRoom(caller, class) {
		return nil, domain.ErrForbidden
	}
	if room.Status == domain.LiveRoomStatusFinished {
		return nil, domain.NewValidationError(map[string]string{"status": "cannot update finished room"})
	}
	room.Config = *dto.Config
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *service) Heartbeat(ctx context.Context, roomID uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	room, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return err
	}
	if !s.canManageRoom(caller, class) {
		return domain.ErrForbidden
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil
	}
	now := time.Now()
	room.HostLastSeenAt = &now
	return s.rooms.Update(ctx, room)
}

func (s *service) List(ctx context.Context, q domain.ListLiveRoomsQuery) ([]domain.LiveRoom, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	scope.Status = q.Status
	scope.ClassID = q.ClassID
	scope.ClassSessionID = q.ClassSessionID
	// Soft-deleted rooms are an admin/manager-only view.
	if q.IncludeDeleted && (caller.IsAdmin || caller.HasPermission(domain.PermLiveSessionsViewAny)) {
		scope.IncludeDeleted = true
	}
	return s.rooms.List(ctx, scope, q.ListParams)
}

func (s *service) StartRecording(ctx context.Context, roomID uuid.UUID) (*domain.LiveRecording, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !s.canManageRoom(caller, class) {
		return nil, domain.ErrForbidden
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil, domain.NewValidationError(map[string]string{"status": "room must be active to record"})
	}

	s3Path := fmt.Sprintf("recordings/%s/%s.mp4", room.ID.String(), uuid.New().String())
	egressID, err := s.livekit.StartRecording(ctx, room.LiveKitRoomName, s3Path)
	if err != nil {
		return nil, fmt.Errorf("livesessions.service.StartRecording: %w", err)
	}

	rec := &domain.LiveRecording{
		LiveRoomID: room.ID,
		EgressID:   egressID,
		Status:     domain.LiveRecordingStatusStarted,
		FileURL:    s3Path,
		StartedAt:  time.Now(),
	}
	if err := s.recordings.Create(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *service) StopRecording(ctx context.Context, recordingID uuid.UUID) (*domain.LiveRecording, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	rec, err := s.recordings.FindByID(ctx, recordingID)
	if err != nil {
		return nil, err
	}
	_, _, class, err := s.loadRoomWithClass(ctx, rec.LiveRoomID)
	if err != nil {
		return nil, err
	}
	if !s.canManageRoom(caller, class) {
		return nil, domain.ErrForbidden
	}
	if rec.Status != domain.LiveRecordingStatusStarted {
		return nil, domain.NewValidationError(map[string]string{"status": "recording not active"})
	}

	if err := s.livekit.StopRecording(ctx, rec.EgressID); err != nil {
		return nil, fmt.Errorf("livesessions.service.StopRecording: %w", err)
	}

	now := time.Now()
	rec.Status = domain.LiveRecordingStatusCompleted
	rec.EndedAt = &now
	rec.Duration = int(now.Sub(rec.StartedAt).Seconds())
	if err := s.recordings.Update(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *service) ListRecordings(ctx context.Context, roomID uuid.UUID, q domain.ListLiveRecordingsQuery) ([]domain.LiveRecording, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	_, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, 0, err
	}
	ok, err = s.canViewRoom(ctx, caller, class)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	return s.recordings.ListByRoom(ctx, roomID, q)
}

func (s *service) ListParticipants(ctx context.Context, roomID uuid.UUID, q domain.ListLiveParticipantsQuery) ([]domain.LiveParticipant, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	_, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, 0, err
	}
	ok, err = s.canViewRoom(ctx, caller, class)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	return s.participants.ListByRoom(ctx, roomID, q)
}

func (s *service) AutoCloseStaleRooms(ctx context.Context) error {
	rooms, err := s.rooms.FindActiveRoomsWithStaleHost(ctx, 1*time.Hour)
	if err != nil {
		return err
	}
	for _, room := range rooms {
		r := room
		if _, err := s.endRoomInternal(ctx, &r); err != nil {
			s.logger.Error("auto-close failed",
				"room_id", r.ID.String(),
				"error", err,
			)
			continue
		}
		s.logger.Info("auto-closed stale room", "room_id", r.ID.String())
	}
	return nil
}

const (
	roomEventRoleChanged = "role_changed"
	roomEventHand        = "hand"
)

// participantMetadata encodes a participant's room role into the LiveKit
// participant metadata so every connected client can read it off the
// participant object — no separate snapshot fetch or seeding required.
func participantMetadata(role domain.ParticipantRole) string {
	b, err := json.Marshal(map[string]string{"role": string(role)})
	if err != nil {
		return ""
	}
	return string(b)
}

func (s *service) broadcastRoomEvent(ctx context.Context, roomName, eventType string, data map[string]any) {
	if s.livekit == nil {
		return
	}
	payload, err := json.Marshal(map[string]any{"type": eventType, "data": data})
	if err != nil {
		s.logger.Error("broadcastRoomEvent: marshal", "event", eventType, "error", err)
		return
	}
	if err := s.livekit.SendData(ctx, roomName, payload, nil); err != nil {
		s.logger.Error("broadcastRoomEvent: send", "event", eventType, "error", err)
	}
}

func (s *service) SetParticipantRole(ctx context.Context, roomID uuid.UUID, identity string, dto domain.SetParticipantRoleDTO) (*domain.LiveParticipant, error) {
	if !dto.Role.Valid() {
		return nil, domain.ErrInvalidParticipantRole
	}
	if dto.Role == domain.ParticipantRoleHost {
		return nil, domain.ErrCannotChangeHostRole
	}

	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !s.canManageRoom(caller, class) {
		return nil, domain.ErrForbidden
	}

	target, err := s.participants.GetActiveParticipant(ctx, roomID, identity)
	if err != nil {
		return nil, err
	}
	if target.Role == domain.ParticipantRoleHost {
		return nil, domain.ErrCannotChangeHostRole
	}

	isPresenter := dto.Role == domain.ParticipantRolePresenter
	sources := publishSources(isPresenter)
	if s.livekit != nil {
		if err := s.livekit.UpdateParticipant(ctx, room.LiveKitRoomName, identity, participantMetadata(dto.Role), sources); err != nil {
			return nil, fmt.Errorf("livesessions.service.SetParticipantRole livekit: %w", err)
		}
	}

	if err := s.participants.UpdateParticipantRole(ctx, roomID, identity, dto.Role); err != nil {
		return nil, fmt.Errorf("livesessions.service.SetParticipantRole persist: %w", err)
	}
	target.Role = dto.Role

	s.broadcastRoomEvent(ctx, room.LiveKitRoomName, roomEventRoleChanged, map[string]any{
		"identity": identity,
		"role":     string(dto.Role),
	})

	return target, nil
}

func (s *service) MuteParticipant(ctx context.Context, roomID uuid.UUID, identity string, dto domain.MuteParticipantDTO) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	room, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return err
	}
	if !s.canManageRoom(caller, class) {
		return domain.ErrForbidden
	}
	if s.livekit == nil {
		return nil
	}
	return s.livekit.MutePublishedTrack(ctx, room.LiveKitRoomName, identity, dto.TrackSID, dto.Muted)
}

func (s *service) SetHand(ctx context.Context, roomID uuid.UUID, dto domain.SetHandDTO) (*domain.LiveParticipant, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, _, _, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}

	identity := caller.UserID.String()
	if err := s.participants.SetHandRaised(ctx, roomID, identity, dto.Raised); err != nil {
		return nil, fmt.Errorf("livesessions.service.SetHand: %w", err)
	}

	participant, err := s.participants.GetActiveParticipant(ctx, roomID, identity)
	if err != nil {
		return nil, err
	}

	s.broadcastRoomEvent(ctx, room.LiveKitRoomName, roomEventHand, map[string]any{
		"identity": identity,
		"raised":   dto.Raised,
	})

	return participant, nil
}

// GetWhiteboard returns the current tldraw snapshot for the room.
// Any participant (viewer or above) may read it; returns an empty board if none has been saved yet.
func (s *service) GetWhiteboard(ctx context.Context, roomID uuid.UUID) (*domain.LiveWhiteboard, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	_, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}
	canView, err := s.canViewRoom(ctx, caller, class)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, domain.ErrForbidden
	}

	wb, err := s.whiteboards.Get(ctx, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrWhiteboardNotFound) {
			return &domain.LiveWhiteboard{
				LiveRoomID: roomID,
				Snapshot:   json.RawMessage("{}"),
			}, nil
		}
		return nil, fmt.Errorf("livesessions.service.GetWhiteboard: %w", err)
	}
	return wb, nil
}

// SaveWhiteboard persists a tldraw snapshot. Only hosts and presenters may write.
func (s *service) SaveWhiteboard(ctx context.Context, roomID uuid.UUID, dto domain.SaveWhiteboardDTO) (*domain.LiveWhiteboard, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	_, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Hosts (canManageRoom) may always draw. Presenters may also draw.
	// Pure viewers are forbidden.
	if !s.canManageRoom(caller, class) {
		participant, err := s.participants.GetActiveParticipant(ctx, roomID, caller.UserID.String())
		if err != nil {
			// Not an active participant → forbidden.
			return nil, domain.ErrForbidden
		}
		if participant.Role != domain.ParticipantRolePresenter {
			return nil, domain.ErrForbidden
		}
	}

	wb, err := s.whiteboards.Upsert(ctx, roomID, dto.Snapshot)
	if err != nil {
		return nil, fmt.Errorf("livesessions.service.SaveWhiteboard: %w", err)
	}
	return wb, nil
}
