package livesessions

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	lk "github.com/4H1R/zoora/internal/platform/livekit"
)

type service struct {
	rooms        domain.LiveRoomRepository
	participants domain.LiveParticipantRepository
	recordings   domain.LiveRecordingRepository
	sessions     domain.ClassSessionRepository
	classes      domain.ClassRepository
	members      domain.ClassMemberRepository
	chatSvc      domain.ChatService
	tx           domain.Transactor
	livekit      *lk.Client
	logger       *slog.Logger
}

func NewService(
	rooms domain.LiveRoomRepository,
	participants domain.LiveParticipantRepository,
	recordings domain.LiveRecordingRepository,
	sessions domain.ClassSessionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	chatSvc domain.ChatService,
	tx domain.Transactor,
	livekit *lk.Client,
	logger *slog.Logger,
) domain.LiveSessionService {
	return &service{
		rooms:        rooms,
		participants: participants,
		recordings:   recordings,
		sessions:     sessions,
		classes:      classes,
		members:      members,
		chatSvc:      chatSvc,
		tx:           tx,
		livekit:      livekit,
		logger:       logger,
	}
}

func (s *service) canUpdateRoom(caller domain.Caller, class *domain.Class) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission("livesessions:update_any") {
		return true
	}
	if caller.HasPermission("livesessions:update") && caller.UserID == class.UserID {
		return true
	}
	return false
}

func (s *service) canManageRoom(caller domain.Caller, class *domain.Class) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission("livesessions:manage_any") {
		return true
	}
	if caller.HasPermission("livesessions:manage") && caller.UserID == class.UserID {
		return true
	}
	return false
}

func (s *service) canViewRoom(ctx context.Context, caller domain.Caller, class *domain.Class) (bool, error) {
	if caller.IsAdmin || caller.HasPermission("livesessions:view_any") {
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

func (s *service) resolveListScope(caller domain.Caller) domain.LiveRoomListScope {
	if caller.IsAdmin || caller.HasPermission("livesessions:view_any") {
		return domain.LiveRoomListScope{All: true}
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

	cfg := dto.Config
	if cfg.MaxParticipants <= 0 {
		cfg = domain.DefaultLiveRoomConfig()
	}

	roomName := fmt.Sprintf("session-%s", dto.ClassSessionID.String())

	room := &domain.LiveRoom{
		ClassSessionID:  dto.ClassSessionID,
		LiveKitRoomName: roomName,
		Status:          domain.LiveRoomStatusCreated,
		Config:          cfg,
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
	hasJoinAny := caller.HasPermission("livesessions:join_any")
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
	canPublish := isModerator || room.Config.AllowMicDefault || room.Config.AllowCameraDefault
	roomAdmin := isModerator

	identity := caller.UserID.String()
	token, err := s.livekit.GenerateToken(room.LiveKitRoomName, identity, identity, canPublish, roomAdmin)
	if err != nil {
		return nil, fmt.Errorf("livesessions.service.JoinRoom token: %w", err)
	}

	participant := &domain.LiveParticipant{
		LiveRoomID: roomID,
		UserID:     caller.UserID,
		Identity:   identity,
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

	_, err = s.livekit.CreateRoom(ctx, room.LiveKitRoomName, uint32(room.Config.MaxParticipants))
	if err != nil {
		return nil, fmt.Errorf("livesessions.service.StartRoom livekit: %w", err)
	}

	now := time.Now()
	room.Status = domain.LiveRoomStatusActive
	room.ActualStartTime = &now
	room.HostLastSeenAt = &now
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}

	if room.Config.AutoRecord {
		s3Path := fmt.Sprintf("recordings/%s/%s.mp4", room.ID.String(), uuid.New().String())
		egressID, err := s.livekit.StartRecording(ctx, room.LiveKitRoomName, s3Path)
		if err != nil {
			s.logger.Error("auto-record failed", "room_id", room.ID.String(), "error", err)
		} else {
			rec := &domain.LiveRecording{
				LiveRoomID: room.ID,
				EgressID:   egressID,
				Status:     domain.LiveRecordingStatusStarted,
				StartedAt:  now,
			}
			if err := s.recordings.Create(ctx, rec); err != nil {
				s.logger.Error("saving auto-record", "room_id", room.ID.String(), "error", err)
			}
		}
	}

	s.logger.Info("live room started",
		"room_id", room.ID.String(),
		"started_by", caller.UserID.String(),
	)
	return room, nil
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

	return room, nil
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

func (s *service) List(ctx context.Context, p domain.ListParams) ([]domain.LiveRoom, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	return s.rooms.List(ctx, scope, p)
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

func (s *service) ListRecordings(ctx context.Context, roomID uuid.UUID) ([]domain.LiveRecording, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	_, _, class, err := s.loadRoomWithClass(ctx, roomID)
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
	return s.recordings.ListByRoom(ctx, roomID)
}

func (s *service) ListParticipants(ctx context.Context, roomID uuid.UUID, p domain.ListParams) ([]domain.LiveParticipant, int64, error) {
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
	return s.participants.ListByRoom(ctx, roomID, p)
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
