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
	"github.com/livekit/protocol/webhook"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
	lk "github.com/4H1R/zoora/internal/platform/livekit"
	"github.com/4H1R/zoora/internal/platform/queue"
)

// LiveKitClient is the LiveKit surface the service depends on, declared
// locally so tests can inject a fake. *lk.Client satisfies it automatically.
type LiveKitClient interface {
	CreateRoom(ctx context.Context, roomName string, maxParticipants uint32) (*lkproto.Room, error)
	DeleteRoom(ctx context.Context, roomName string) error
	GenerateToken(roomName, identity, name, metadata string, sources []lkproto.TrackSource, roomAdmin bool) (string, error)
	StartRecording(ctx context.Context, roomName, s3Path string) (string, error)
	StopRecording(ctx context.Context, egressID string) error
	ListParticipants(ctx context.Context, roomName string) ([]*lkproto.ParticipantInfo, error)
	UpdateParticipant(ctx context.Context, roomName, identity, metadata string, sources []lkproto.TrackSource) error
	MutePublishedTrack(ctx context.Context, roomName, identity, trackSID string, muted bool) error
	RemoveParticipant(ctx context.Context, roomName, identity string) error
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
	pollSvc      domain.PollService
	tx           domain.Transactor
	livekit      LiveKitClient
	queue        *queue.Client
	ent          entitlements.Service
	// hostGracePeriod is how long a room may stay open after its last host
	// leaves before the delayed close task (and the safety-net sweep) closes it.
	hostGracePeriod time.Duration
	logger          *slog.Logger
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
	pollSvc domain.PollService,
	tx domain.Transactor,
	livekit LiveKitClient,
	queueClient *queue.Client,
	ent entitlements.Service,
	hostGracePeriod time.Duration,
	logger *slog.Logger,
) domain.LiveSessionService {
	// Guard against a zero/unset grace period closing rooms the instant their
	// host blips; fall back to the documented 15-minute default.
	if hostGracePeriod <= 0 {
		hostGracePeriod = 15 * time.Minute
	}
	if livekit == nil {
		panic("livesessions.NewService: livekit client is required")
	}
	return &service{
		rooms:           rooms,
		participants:    participants,
		recordings:      recordings,
		whiteboards:     whiteboards,
		sessions:        sessions,
		classes:         classes,
		members:         members,
		chatSvc:         chatSvc,
		pollSvc:         pollSvc,
		tx:              tx,
		livekit:         livekit,
		queue:           queueClient,
		ent:             ent,
		hostGracePeriod: hostGracePeriod,
		logger:          logger,
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

	cfg := normalizeRoomConfig(dto.Config, int(caller.Ent.Limit(domain.LimitMaxParticipants)))

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
			ModelType: domain.ChatModelLiveSession,
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
		return nil, domain.NewValidationError(map[string]string{"status": "room already finished"})
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
		if !isModerator {
			return nil, domain.NewValidationError(map[string]string{"status": "room not started yet"})
		}
		if _, err := s.startRoomInternal(ctx, room, class.OrganizationID, caller.Ent); err != nil {
			if !errors.Is(err, domain.ErrConflict) {
				return nil, err
			}
			// Lost the auto-start race to a concurrent moderator join —
			// reload and continue as long as the room ended up active.
			room, err = s.rooms.FindByID(ctx, roomID)
			if err != nil {
				return nil, err
			}
			if room.Status != domain.LiveRoomStatusActive {
				return nil, domain.NewValidationError(map[string]string{"status": "room not started yet"})
			}
		}
	}

	// Defensive: guarantee the LiveKit room exists before issuing a token
	// (idempotent — returns the existing room if already created).
	if _, err := s.livekit.CreateRoom(ctx, room.LiveKitRoomName, uint32(room.Config.MaxParticipants)); err != nil {
		return nil, fmt.Errorf("livesessions.service.JoinRoom livekit: %w", err)
	}

	identity := caller.UserID.String()
	displayName := caller.Name
	if displayName == "" {
		displayName = identity
	}

	role := domain.ParticipantRoleViewer
	if isModerator {
		role = domain.ParticipantRoleHost
	}

	// A rejoin (refresh, reconnect) reuses the open participation row instead
	// of stacking a duplicate, and keeps any role a host granted mid-session
	// (e.g. presenter) so the new token carries the same publish rights.
	participant, err := s.participants.FindActiveByRoomAndUser(ctx, roomID, caller.UserID)
	switch {
	case err == nil:
		if !isModerator {
			role = participant.Role
		}
	case errors.Is(err, domain.ErrNotFound):
		participant = nil
	default:
		return nil, err
	}

	canPublish := role == domain.ParticipantRoleHost || role == domain.ParticipantRolePresenter
	sources := publishSources(canPublish)

	token, err := s.livekit.GenerateToken(room.LiveKitRoomName, identity, displayName, participantMetadata(role), sources, isModerator)
	if err != nil {
		return nil, fmt.Errorf("livesessions.service.JoinRoom token: %w", err)
	}
	if participant == nil {
		participant = &domain.LiveParticipant{
			LiveRoomID: roomID,
			UserID:     caller.UserID,
			Identity:   identity,
			Role:       role,
			JoinedAt:   time.Now(),
		}
		if err := s.participants.Create(ctx, participant); err != nil {
			return nil, err
		}
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

	if chat, err := s.chatSvc.FindChatByModel(ctx, domain.ChatModelLiveSession, roomID); err == nil {
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

	if _, err := s.startRoomInternal(ctx, room, class.OrganizationID, caller.Ent); err != nil {
		return nil, err
	}

	s.logger.Info("live room started",
		"room_id", room.ID.String(),
		"started_by", caller.UserID.String(),
	)
	return room, nil
}

// startRoomInternal creates the LiveKit room and promotes the DB row to active.
// Callers must have already verified manage rights and that the room is in the
// created state. Shared by StartRoom and JoinRoom (host auto-start) so the two
// paths can never drift. Returns domain.ErrConflict when a concurrent start won
// the created→active transition.
func (s *service) startRoomInternal(ctx context.Context, room *domain.LiveRoom, orgID uuid.UUID, ent domain.Entitlements) (*domain.LiveRoom, error) {
	// Enforce the org's concurrent-active-rooms limit before the room goes live.
	if s.ent != nil {
		if err := s.ent.CheckConcurrentRoomsLimit(ctx, orgID, ent); err != nil {
			return nil, err
		}
	}
	if _, err := s.livekit.CreateRoom(ctx, room.LiveKitRoomName, uint32(room.Config.MaxParticipants)); err != nil {
		return nil, fmt.Errorf("livesessions.service.startRoomInternal livekit: %w", err)
	}

	now := time.Now()
	room.Status = domain.LiveRoomStatusActive
	room.ActualStartTime = &now
	room.HostLastSeenAt = &now
	if err := s.rooms.Transition(ctx, room, domain.LiveRoomStatusCreated); err != nil {
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

// endRoomInternal finalizes an active room. The guarded active→finished
// transition runs first so a concurrent close (webhook vs task vs sweep vs
// manual end) short-circuits with domain.ErrConflict before any side effects.
// Teardown steps after the transition are best-effort: each failure is logged
// but never blocks the rest (the room is already finished in the DB).
func (s *service) endRoomInternal(ctx context.Context, room *domain.LiveRoom) (*domain.LiveRoom, error) {
	now := time.Now()
	room.Status = domain.LiveRoomStatusFinished
	room.ActualEndTime = &now
	if err := s.rooms.Transition(ctx, room, domain.LiveRoomStatusActive); err != nil {
		return nil, err
	}

	if rec, err := s.recordings.FindActiveByRoom(ctx, room.ID); err == nil {
		if err := s.livekit.StopRecording(ctx, rec.EgressID); err != nil {
			s.logger.Error("end room: stop recording", "room_id", room.ID.String(), "egress_id", rec.EgressID, "error", err)
		}
		rec.Status = domain.LiveRecordingStatusCompleted
		rec.EndedAt = &now
		if err := s.recordings.Update(ctx, rec); err != nil {
			s.logger.Error("end room: finalize recording", "room_id", room.ID.String(), "error", err)
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		s.logger.Error("end room: lookup active recording", "room_id", room.ID.String(), "error", err)
	}

	if err := s.participants.MarkAllLeft(ctx, room.ID, now); err != nil {
		s.logger.Error("end room: mark participants left", "room_id", room.ID.String(), "error", err)
	}

	if err := s.livekit.DeleteRoom(ctx, room.LiveKitRoomName); err != nil {
		s.logger.Error("end room: delete livekit room", "room_id", room.ID.String(), "error", err)
	}

	if err := s.chatSvc.ArchiveByModel(ctx, domain.ChatModelLiveSession, room.ID); err != nil {
		s.logger.Error("failed to archive chat for room", "room_id", room.ID.String(), "error", err)
	}

	// Close any open polls so late clicks can't mutate results after the room ends.
	if err := s.pollSvc.CloseByModel(ctx, domain.ChatModelLiveSession, room.ID); err != nil {
		s.logger.Error("failed to close polls for room", "room_id", room.ID.String(), "error", err)
	}

	// The whiteboard snapshot is only readable during an active session
	// (JoinRoom rejects finished rooms), so it is dead data once the room ends.
	if err := s.whiteboards.Delete(ctx, room.ID); err != nil {
		s.logger.Error("end room: delete whiteboard", "room_id", room.ID.String(), "error", err)
	}

	// Clear any pending no-host close task; the room is already finished so the
	// task would otherwise linger in Redis until it fires and self-cancels.
	s.disarmCloseIfNoHost(room)

	s.enqueueAutoMark(ctx, room)
	s.enqueueSlidesCleanup(ctx, room)

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

// enqueueSlidesCleanup schedules deletion of the slide PDFs the host shared into
// this room (media rows + S3 objects). Best-effort: failures are logged and
// never block room teardown.
func (s *service) enqueueSlidesCleanup(ctx context.Context, room *domain.LiveRoom) {
	if s.queue == nil {
		return
	}
	payload, err := json.Marshal(domain.MediaCleanupPayload{
		ModelType:      domain.MediaModelLiveRoom,
		ModelID:        room.ID,
		CollectionName: domain.MediaCollectionSlides,
	})
	if err != nil {
		s.logger.Error("slides cleanup enqueue: marshal payload", "room_id", room.ID.String(), "error", err)
		return
	}
	if _, err := s.queue.Enqueue(asynq.NewTask(domain.TypeMediaCleanup, payload)); err != nil {
		s.logger.Error("slides cleanup enqueue", "room_id", room.ID.String(), "error", err)
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
	room.Config = normalizeRoomConfig(*dto.Config, int(caller.Ent.Limit(domain.LimitMaxParticipants)))
	if err := s.rooms.UpdateConfig(ctx, room.ID, room.Config); err != nil {
		return nil, err
	}
	return room, nil
}

// normalizeRoomConfig sets the participant cap from the org's plan: a
// non-positive request defaults to the plan ceiling, and a request above the
// ceiling is clamped down to it. Caller-supplied flags are preserved. Shared by
// create and update so the two paths can't drift. A non-positive ceiling (no
// plan resolved) falls back to the built-in default.
func normalizeRoomConfig(cfg domain.LiveRoomConfig, ceiling int) domain.LiveRoomConfig {
	if ceiling <= 0 {
		ceiling = domain.DefaultLiveRoomConfig().MaxParticipants
	}
	if cfg.MaxParticipants <= 0 || cfg.MaxParticipants > ceiling {
		cfg.MaxParticipants = ceiling
	}
	return cfg
}

func (s *service) Heartbeat(ctx context.Context, roomID uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	_, _, class, err := s.loadRoomWithClass(ctx, roomID)
	if err != nil {
		return err
	}
	if !s.canManageRoom(caller, class) {
		return domain.ErrForbidden
	}
	// Conditional touch: only bumps host_last_seen_at while the room is still
	// active, so a heartbeat racing a close can never resurrect a finished
	// room (a full save would write the stale status back).
	return s.rooms.TouchHostLastSeen(ctx, roomID, time.Now())
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
	if !caller.HasFeature(domain.FeatureRecording) {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureRecording)
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil, domain.NewValidationError(map[string]string{"status": "room must be active to record"})
	}

	// One active egress per room: a double-start would record (and bill) twice
	// and orphan the loser, since teardown only stops one active recording.
	if _, err := s.recordings.FindActiveByRoom(ctx, room.ID); err == nil {
		return nil, domain.ErrConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	// Namespace recording objects per tenant (orgs/{org_id}/…) so a single
	// bucket isolates each organization's files by key prefix, matching the
	// media object layout.
	s3Path := fmt.Sprintf("orgs/%s/recordings/%s/%s.mp4", class.OrganizationID.String(), room.ID.String(), uuid.New().String())
	// Bound the egress start: if no egress worker is available the LiveKit RPC
	// blocks until the client gives up, which surfaces to the browser as a 502.
	// A deadline turns that into a clean, fast error instead.
	startCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	egressID, err := s.livekit.StartRecording(startCtx, room.LiveKitRoomName, s3Path)
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

	now := time.Now()
	status := domain.LiveRecordingStatusCompleted
	if err := s.livekit.StopRecording(ctx, rec.EgressID); err != nil {
		// The egress already terminated on LiveKit's side (commonly it aborted
		// before the user hit stop — e.g. the composite template never loaded).
		// Reconcile the record as failed and return it instead of 500ing.
		if !errors.Is(err, lk.ErrEgressNotActive) {
			return nil, fmt.Errorf("livesessions.service.StopRecording: %w", err)
		}
		s.logger.Warn("stop recording: egress already terminal, marking failed",
			"recording_id", rec.ID.String(), "egress_id", rec.EgressID)
		status = domain.LiveRecordingStatusFailed
	}

	rec.Status = status
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

// AutoCloseStaleRooms is the periodic safety net for missed LiveKit webhooks.
// It finds active rooms whose host heartbeat has gone stale past the grace
// period, then — critically — cross-checks LiveKit before closing so a stale
// heartbeat alone can never tear down a room that still has a host connected.
func (s *service) AutoCloseStaleRooms(ctx context.Context) error {
	rooms, err := s.rooms.FindActiveRoomsWithStaleHost(ctx, s.hostGracePeriod)
	if err != nil {
		return err
	}
	for _, room := range rooms {
		r := room
		present, err := s.hostPresent(ctx, &r)
		if err != nil {
			s.logger.Error("auto-close: host presence check failed", "room_id", r.ID.String(), "error", err)
			continue
		}
		if present {
			// Host is connected but not heartbeating (e.g. client-side heartbeat
			// stalled). Refresh last-seen so we don't re-scan it every sweep.
			_ = s.rooms.TouchHostLastSeen(ctx, r.ID, time.Now())
			continue
		}
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

// closeTaskID is the room-scoped, deterministic Asynq task ID for the delayed
// no-host close. Deterministic so re-arming is idempotent (ErrTaskIDConflict)
// and so a returning host can cancel the exact pending task.
func closeTaskID(roomID uuid.UUID) string {
	return "livesession-close-" + roomID.String()
}

// hostPresent asks LiveKit (the source of truth for real-time presence) whether
// any connected participant carries the host role in their metadata. A room
// LiveKit no longer knows (already torn down by its empty_timeout — exactly the
// missed-webhook case the sweep exists for) counts as "no host present" so the
// close paths can still finalize it instead of erroring forever.
func (s *service) hostPresent(ctx context.Context, room *domain.LiveRoom) (bool, error) {
	parts, err := s.livekit.ListParticipants(ctx, room.LiveKitRoomName)
	if err != nil {
		if errors.Is(err, lk.ErrRoomNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("livesessions.service.hostPresent: %w", err)
	}
	for _, p := range parts {
		if p.Metadata == "" {
			continue
		}
		var meta struct {
			Role string `json:"role"`
		}
		if json.Unmarshal([]byte(p.Metadata), &meta) == nil &&
			meta.Role == string(domain.ParticipantRoleHost) {
			return true, nil
		}
	}
	return false, nil
}

// armCloseIfNoHost schedules the delayed no-host close for a room. Idempotent:
// if a timer is already armed (same deterministic task ID) it is left running
// so the grace window keeps counting from the first host departure.
func (s *service) armCloseIfNoHost(ctx context.Context, room *domain.LiveRoom) {
	if s.queue == nil {
		return
	}
	payload, err := json.Marshal(domain.LiveSessionCloseIfNoHostPayload{RoomID: room.ID})
	if err != nil {
		s.logger.Error("arm close-if-no-host: marshal", "room_id", room.ID.String(), "error", err)
		return
	}
	task := asynq.NewTask(domain.TypeLiveSessionCloseIfNoHost, payload)
	_, err = s.queue.Enqueue(task,
		asynq.TaskID(closeTaskID(room.ID)),
		asynq.Queue("default"),
		asynq.ProcessIn(s.hostGracePeriod),
	)
	if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
		s.logger.Error("arm close-if-no-host: enqueue", "room_id", room.ID.String(), "error", err)
		return
	}
	s.logger.Info("armed no-host close", "room_id", room.ID.String(), "grace", s.hostGracePeriod.String())
}

// disarmCloseIfNoHost cancels a pending no-host close (best effort) when a host
// reconnects within the grace window.
func (s *service) disarmCloseIfNoHost(room *domain.LiveRoom) {
	if s.queue == nil {
		return
	}
	if err := s.queue.Cancel("default", closeTaskID(room.ID)); err != nil {
		s.logger.Error("disarm close-if-no-host", "room_id", room.ID.String(), "error", err)
		return
	}
	s.logger.Info("disarmed no-host close", "room_id", room.ID.String())
}

// OnLiveKitEvent reacts to a verified LiveKit webhook (see interface doc).
func (s *service) OnLiveKitEvent(ctx context.Context, eventType, livekitRoomName, participantIdentity string) error {
	switch eventType {
	case webhook.EventParticipantLeft:
		return s.onParticipantLeft(ctx, livekitRoomName, participantIdentity)
	case webhook.EventParticipantJoined:
		return s.onParticipantJoined(ctx, livekitRoomName)
	case webhook.EventRoomFinished:
		return s.onRoomFinished(ctx, livekitRoomName)
	default:
		return nil
	}
}

func (s *service) onParticipantLeft(ctx context.Context, lkRoomName, identity string) error {
	room, err := s.rooms.FindByLiveKitRoomName(ctx, lkRoomName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil
	}

	// Close the participation row server-side so attendance stays honest even
	// when the client crashed and never called /leave. Best-effort: a missing
	// row (client already left cleanly) is fine.
	if identity != "" {
		err := s.participants.MarkLeftByIdentity(ctx, room.ID, identity, time.Now())
		if err != nil && !errors.Is(err, domain.ErrParticipantNotFound) {
			s.logger.Error("participant_left: mark left", "room_id", room.ID.String(), "identity", identity, "error", err)
		}
	}

	present, err := s.hostPresent(ctx, room)
	if err != nil {
		return err
	}
	if present {
		return nil
	}
	s.armCloseIfNoHost(ctx, room)
	return nil
}

func (s *service) onParticipantJoined(ctx context.Context, lkRoomName string) error {
	room, err := s.rooms.FindByLiveKitRoomName(ctx, lkRoomName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil
	}
	present, err := s.hostPresent(ctx, room)
	if err != nil {
		return err
	}
	if present {
		s.disarmCloseIfNoHost(room)
	}
	return nil
}

func (s *service) onRoomFinished(ctx context.Context, lkRoomName string) error {
	room, err := s.rooms.FindByLiveKitRoomName(ctx, lkRoomName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil
	}
	// LiveKit tore the room down itself (its own empty_timeout, or all
	// participants incl. host disconnected). Finalize our side.
	if _, err := s.endRoomInternal(ctx, room); err != nil {
		return err
	}
	s.logger.Info("finalized room after livekit room_finished", "room_id", room.ID.String())
	return nil
}

// CloseRoomIfNoHost is the delayed-task target: close the room only if LiveKit
// confirms no host is present. A host who returned within the grace window
// keeps the room alive.
func (s *service) CloseRoomIfNoHost(ctx context.Context, roomID uuid.UUID) error {
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil
	}
	present, err := s.hostPresent(ctx, room)
	if err != nil {
		return err
	}
	if present {
		s.logger.Info("no-host close skipped: host present", "room_id", room.ID.String())
		return nil
	}
	if _, err := s.endRoomInternal(ctx, room); err != nil {
		return err
	}
	s.logger.Info("closed room: no host after grace period", "room_id", room.ID.String())
	return nil
}

// OnEgressEnded finalizes a recording from a LiveKit egress_ended webhook —
// the authoritative source for file size, duration, and failure status
// (StopRecording only knows the egress was asked to stop).
func (s *service) OnEgressEnded(ctx context.Context, result domain.EgressResult) error {
	rec, err := s.recordings.FindByEgressID(ctx, result.EgressID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}

	rec.Status = domain.LiveRecordingStatusCompleted
	if result.Failed {
		rec.Status = domain.LiveRecordingStatusFailed
	}
	if result.SizeBytes > 0 {
		rec.Size = result.SizeBytes
	}
	if result.Duration > 0 {
		rec.Duration = int(result.Duration.Seconds())
	}
	if rec.EndedAt == nil {
		now := time.Now()
		rec.EndedAt = &now
	}
	if err := s.recordings.Update(ctx, rec); err != nil {
		return fmt.Errorf("livesessions.service.OnEgressEnded: %w", err)
	}
	s.logger.Info("finalized recording from egress webhook",
		"egress_id", result.EgressID,
		"status", string(rec.Status),
		"size", rec.Size,
	)
	return nil
}

const (
	roomEventRoleChanged = "role_changed"
	roomEventHand        = "hand"
	roomEventRemoved     = "removed"
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

// broadcastHand emits a `hand` room event. `raisedAt` is included only when the
// hand is raised and a timestamp is known, so the client queue can order hands.
func (s *service) broadcastHand(ctx context.Context, roomName, identity string, raised bool, raisedAt *time.Time) {
	data := map[string]any{"identity": identity, "raised": raised}
	if raised && raisedAt != nil {
		data["raisedAt"] = raisedAt.UnixMilli()
	}
	s.broadcastRoomEvent(ctx, roomName, roomEventHand, data)
}

func (s *service) broadcastRoomEvent(ctx context.Context, roomName, eventType string, data map[string]any) {
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
	if err := s.livekit.UpdateParticipant(ctx, room.LiveKitRoomName, identity, participantMetadata(dto.Role), sources); err != nil {
		return nil, fmt.Errorf("livesessions.service.SetParticipantRole livekit: %w", err)
	}

	if err := s.participants.UpdateParticipantRole(ctx, roomID, identity, dto.Role); err != nil {
		return nil, fmt.Errorf("livesessions.service.SetParticipantRole persist: %w", err)
	}
	target.Role = dto.Role

	// A promotion answers the raised hand — clear it so the queue drops the
	// participant (who, as a publisher, loses the raise-hand button anyway).
	if target.HandRaisedAt != nil {
		if err := s.participants.SetHandRaised(ctx, roomID, identity, false); err != nil {
			return nil, fmt.Errorf("livesessions.service.SetParticipantRole lower hand: %w", err)
		}
		target.HandRaisedAt = nil
		s.broadcastHand(ctx, room.LiveKitRoomName, identity, false, nil)
	}

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
	return s.livekit.MutePublishedTrack(ctx, room.LiveKitRoomName, identity, dto.TrackSID, dto.Muted)
}

func (s *service) RemoveParticipant(ctx context.Context, roomID uuid.UUID, identity string) error {
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

	// A host may not eject a fellow host (hosts are immutable for the session),
	// nor themselves — mirrors the SetParticipantRole/mute guardrails.
	if identity == caller.UserID.String() {
		return domain.ErrCannotRemoveSelf
	}
	target, err := s.participants.GetActiveParticipant(ctx, roomID, identity)
	if err != nil {
		return err
	}
	if target.Role == domain.ParticipantRoleHost {
		return domain.ErrCannotRemoveHost
	}

	// Disconnect at LiveKit first (idempotent — a NotFound is swallowed), then
	// close the participation row. The participant_left webhook is the backstop
	// but we mark left here so the DB is honest even without webhook delivery.
	if err := s.livekit.RemoveParticipant(ctx, room.LiveKitRoomName, identity); err != nil {
		return fmt.Errorf("livesessions.service.RemoveParticipant livekit: %w", err)
	}
	if err := s.participants.MarkLeftByIdentity(ctx, roomID, identity, time.Now()); err != nil {
		return fmt.Errorf("livesessions.service.RemoveParticipant persist: %w", err)
	}

	s.broadcastRoomEvent(ctx, room.LiveKitRoomName, roomEventRemoved, map[string]any{
		"identity": identity,
	})
	s.logger.Info("host removed participant from live room",
		"room_id", roomID.String(),
		"identity", identity,
		"by_user_id", caller.UserID.String(),
	)
	return nil
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

	s.broadcastHand(ctx, room.LiveKitRoomName, identity, dto.Raised, participant.HandRaisedAt)

	return participant, nil
}

// SetParticipantHand lets a host lower (or raise) another participant's hand.
// Authorization mirrors role/mute: only a room manager may call it.
func (s *service) SetParticipantHand(ctx context.Context, roomID uuid.UUID, identity string, dto domain.SetHandDTO) (*domain.LiveParticipant, error) {
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

	if err := s.participants.SetHandRaised(ctx, roomID, identity, dto.Raised); err != nil {
		return nil, fmt.Errorf("livesessions.service.SetParticipantHand: %w", err)
	}

	participant, err := s.participants.GetActiveParticipant(ctx, roomID, identity)
	if err != nil {
		return nil, err
	}

	s.broadcastHand(ctx, room.LiveKitRoomName, identity, dto.Raised, participant.HandRaisedAt)

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
