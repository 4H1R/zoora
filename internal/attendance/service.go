package attendance

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo         domain.AttendanceRepository
	classes      domain.ClassRepository
	sessions     domain.ClassSessionRepository
	members      domain.ClassMemberRepository
	liveRooms    domain.LiveRoomRepository
	participants domain.LiveParticipantRepository
	offlineViews domain.OfflineRoomViewRepository
	offlineRooms domain.OfflineRoomRepository
	logger       *slog.Logger
}

func NewService(
	repo domain.AttendanceRepository,
	classes domain.ClassRepository,
	sessions domain.ClassSessionRepository,
	members domain.ClassMemberRepository,
	liveRooms domain.LiveRoomRepository,
	participants domain.LiveParticipantRepository,
	offlineViews domain.OfflineRoomViewRepository,
	offlineRooms domain.OfflineRoomRepository,
	logger *slog.Logger,
) domain.AttendanceService {
	return &service{
		repo:         repo,
		classes:      classes,
		sessions:     sessions,
		members:      members,
		liveRooms:    liveRooms,
		participants: participants,
		offlineViews: offlineViews,
		offlineRooms: offlineRooms,
		logger:       logger,
	}
}

func canManageAttendance(caller domain.Caller, class *domain.Class) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission("attendance:create_any") || caller.HasPermission("attendance:update_any") {
		return true
	}
	return caller.UserID == class.UserID
}

// upsertEntry updates an existing attendance row for (session,user) or creates a
// new one. Mirrors the dedupe used by AutoMark so re-marking never duplicates.
func (s *service) upsertEntry(ctx context.Context, class *domain.Class, sessionID uuid.UUID, entry domain.CreateAttendanceDTO) (*domain.Attendance, error) {
	existing, err := s.repo.FindBySessionAndUser(ctx, sessionID, entry.UserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		existing.Status = entry.Status
		existing.Remarks = entry.Remarks
		existing.IsAutoMarked = entry.IsAutoMarked
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	a := &domain.Attendance{
		OrganizationID: class.OrganizationID,
		ClassID:        class.ID,
		ClassSessionID: sessionID,
		UserID:         entry.UserID,
		Status:         entry.Status,
		IsAutoMarked:   entry.IsAutoMarked,
		Remarks:        entry.Remarks,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) Mark(ctx context.Context, classID, sessionID uuid.UUID, dto domain.CreateAttendanceDTO) (*domain.Attendance, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	if !canManageAttendance(caller, class) {
		return nil, domain.ErrForbidden
	}
	session, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.ClassID != classID {
		return nil, domain.ErrNotFound
	}

	a, err := s.upsertEntry(ctx, class, sessionID, dto)
	if err != nil {
		return nil, err
	}
	s.logger.Info("attendance marked",
		"attendance_id", a.ID.String(),
		"session_id", sessionID.String(),
		"user_id", dto.UserID.String(),
		"status", string(dto.Status),
		"marked_by", caller.UserID.String(),
	)
	return a, nil
}

func (s *service) BulkMark(ctx context.Context, classID, sessionID uuid.UUID, dto domain.BulkCreateAttendanceDTO) ([]domain.Attendance, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	if !canManageAttendance(caller, class) {
		return nil, domain.ErrForbidden
	}
	session, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.ClassID != classID {
		return nil, domain.ErrNotFound
	}

	var results []domain.Attendance
	for _, entry := range dto.Entries {
		a, err := s.upsertEntry(ctx, class, sessionID, entry)
		if err != nil {
			return nil, err
		}
		results = append(results, *a)
	}
	s.logger.Info("bulk attendance marked",
		"session_id", sessionID.String(),
		"count", len(results),
		"marked_by", caller.UserID.String(),
	)
	return results, nil
}

func (s *service) AutoMark(ctx context.Context, classID, sessionID uuid.UUID, dto domain.AutoMarkAttendanceDTO) (*domain.AutoMarkResult, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	if !canManageAttendance(caller, class) {
		return nil, domain.ErrForbidden
	}
	session, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.ClassID != classID {
		return nil, domain.ErrNotFound
	}

	allMembers, err := s.members.ListAllByClass(ctx, classID)
	if err != nil {
		return nil, err
	}

	presentUserIDs, err := s.resolvePresent(ctx, classID, dto)
	if err != nil {
		return nil, err
	}
	presentSet := make(map[uuid.UUID]bool, len(presentUserIDs))
	for _, uid := range presentUserIDs {
		presentSet[uid] = true
	}

	result := &domain.AutoMarkResult{}
	for _, member := range allMembers {
		_, err := s.repo.FindBySessionAndUser(ctx, sessionID, member.UserID)
		if err == nil {
			result.Skipped++
			continue
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}

		status := domain.AttendanceStatusAbsent
		if presentSet[member.UserID] {
			status = domain.AttendanceStatusPresent
		}

		a := &domain.Attendance{
			OrganizationID: class.OrganizationID,
			ClassID:        classID,
			ClassSessionID: sessionID,
			UserID:         member.UserID,
			Status:         status,
			IsAutoMarked:   true,
			Remarks:        "auto-marked from " + string(dto.Source),
		}
		if err := s.repo.Create(ctx, a); err != nil {
			return nil, err
		}
		result.Marked++
	}

	s.logger.Info("auto-mark attendance completed",
		"session_id", sessionID.String(),
		"source", string(dto.Source),
		"room_id", dto.RoomID.String(),
		"marked", result.Marked,
		"skipped", result.Skipped,
		"triggered_by", caller.UserID.String(),
	)
	return result, nil
}

func (s *service) resolvePresent(ctx context.Context, classID uuid.UUID, dto domain.AutoMarkAttendanceDTO) ([]uuid.UUID, error) {
	switch dto.Source {
	case domain.AutoMarkSourceLive:
		return s.resolvePresentFromLive(ctx, classID, dto.RoomID, dto.MinDurationSeconds)
	case domain.AutoMarkSourceOffline:
		return s.resolvePresentFromOffline(ctx, classID, dto.RoomID)
	default:
		return nil, domain.ErrValidation
	}
}

func (s *service) resolvePresentFromLive(ctx context.Context, classID uuid.UUID, roomID uuid.UUID, minDuration int) ([]uuid.UUID, error) {
	room, err := s.liveRooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	session, err := s.sessions.FindByID(ctx, room.ClassSessionID)
	if err != nil {
		return nil, err
	}
	if session.ClassID != classID {
		return nil, domain.ErrNotFound
	}

	participants, err := s.participants.ListAllByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	seen := make(map[uuid.UUID]int)
	for _, p := range participants {
		seen[p.UserID] += p.TotalDurationSeconds
	}

	var present []uuid.UUID
	for userID, dur := range seen {
		if dur >= minDuration {
			present = append(present, userID)
		}
	}
	return present, nil
}

func (s *service) resolvePresentFromOffline(ctx context.Context, classID uuid.UUID, roomID uuid.UUID) ([]uuid.UUID, error) {
	room, err := s.offlineRooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.ClassID != classID {
		return nil, domain.ErrNotFound
	}
	return s.offlineViews.ListDistinctUsersByRoom(ctx, roomID)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateAttendanceDTO) (*domain.Attendance, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	class, err := s.classes.FindByID(ctx, a.ClassID)
	if err != nil {
		return nil, err
	}
	if !canManageAttendance(caller, class) {
		return nil, domain.ErrForbidden
	}
	if dto.Status != nil {
		a.Status = *dto.Status
	}
	if dto.Remarks != nil {
		a.Remarks = *dto.Remarks
	}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	class, err := s.classes.FindByID(ctx, a.ClassID)
	if err != nil {
		return err
	}
	if !canManageAttendance(caller, class) {
		return domain.ErrForbidden
	}
	return s.repo.Delete(ctx, id)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Attendance, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if caller.IsAdmin || caller.HasPermission("attendance:view_any") {
		return a, nil
	}
	class, err := s.classes.FindByID(ctx, a.ClassID)
	if err != nil {
		return nil, err
	}
	if caller.UserID == class.UserID || caller.UserID == a.UserID {
		return a, nil
	}
	return nil, domain.ErrForbidden
}

func (s *service) ListBySession(ctx context.Context, classID, sessionID uuid.UUID, q domain.ListAttendanceQuery) ([]domain.Attendance, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, 0, err
	}
	session, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, 0, err
	}
	if session.ClassID != classID {
		return nil, 0, domain.ErrNotFound
	}
	if !caller.IsAdmin && !caller.HasPermission("attendance:view_any") && caller.UserID != class.UserID {
		enrolled, err := s.members.Exists(ctx, classID, caller.UserID)
		if err != nil {
			return nil, 0, err
		}
		if !enrolled {
			return nil, 0, domain.ErrForbidden
		}
		userID := caller.UserID
		q.UserID = &userID
	}
	return s.repo.ListBySession(ctx, sessionID, q)
}

// ListMine returns the caller's own attendance history + a status summary.
func (s *service) ListMine(ctx context.Context, p domain.ListParams) (*domain.MyAttendance, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	rows, _, err := s.repo.ListByUser(ctx, caller.UserID, p)
	if err != nil {
		return nil, fmt.Errorf("listing my attendance: %w", err)
	}
	res := &domain.MyAttendance{Items: rows}
	for i := range rows {
		switch rows[i].Status {
		case domain.AttendanceStatusPresent:
			res.Summary.Present++
		case domain.AttendanceStatusAbsent:
			res.Summary.Absent++
		case domain.AttendanceStatusLate:
			res.Summary.Late++
		case domain.AttendanceStatusExcused:
			res.Summary.Excused++
		}
	}
	return res, nil
}
