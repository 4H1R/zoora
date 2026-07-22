package attendance

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/authz"
)

// attendanceMatrixMaxSessions caps the column axis. Classes never approach
// this; it exists so the "all sessions" fetch stays a single bounded query.
const attendanceMatrixMaxSessions = 1000

type service struct {
	repo         domain.AttendanceRepository
	classes      domain.ClassRepository
	sessions     domain.ClassSessionRepository
	members      domain.ClassMemberRepository
	liveRooms    domain.LiveRoomRepository
	participants domain.LiveParticipantRepository
	offlineViews domain.OfflineRoomViewRepository
	offlineRooms domain.OfflineRoomRepository
	orgSettings  domain.OrganizationSettingsProvider
	resolver     *authz.Resolver
	tx           domain.Transactor
	audit        domain.AuditRecorder
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
	orgSettings domain.OrganizationSettingsProvider,
	resolver *authz.Resolver,
	tx domain.Transactor,
	audit domain.AuditRecorder,
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
		orgSettings:  orgSettings,
		resolver:     resolver,
		tx:           tx,
		audit:        audit,
		logger:       logger,
	}
}

func canManageAttendance(caller domain.Caller, class *domain.Class) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermAttendanceCreateAny) || caller.HasPermission(domain.PermAttendanceUpdateAny) {
		return true
	}
	return caller.UserID == class.UserID
}

// upsertEntry updates an existing attendance row for (session,user) or creates a
// new one. Mirrors the dedupe used by AutoMark so re-marking never duplicates.
// upsertEntry returns the persisted row plus created=true when a new row was
// inserted (false when an existing row was updated), so callers can record the
// right audit verb.
func (s *service) upsertEntry(ctx context.Context, class *domain.Class, sessionID uuid.UUID, entry domain.CreateAttendanceDTO) (*domain.Attendance, bool, error) {
	existing, err := s.repo.FindBySessionAndUser(ctx, sessionID, entry.UserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, false, err
	}
	if existing != nil {
		existing.Status = entry.Status
		existing.Remarks = entry.Remarks
		existing.IsAutoMarked = entry.IsAutoMarked
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, false, err
		}
		return existing, false, nil
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
		return nil, false, err
	}
	return a, true, nil
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

	var a *domain.Attendance
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		var created bool
		a, created, err = s.upsertEntry(ctx, class, sessionID, dto)
		if err != nil {
			return err
		}
		action := domain.AuditUpdated
		if created {
			action = domain.AuditCreated
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      action,
			TargetType:  domain.AuditTargetAttendance,
			TargetID:    &a.ID,
			TargetLabel: class.Name,
			OrgID:       &class.OrganizationID,
			Metadata: map[string]any{
				"class_id":   class.ID.String(),
				"session_id": sessionID.String(),
				"user_id":    dto.UserID.String(),
				"status":     string(dto.Status),
			},
		})
	})
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
		a, _, err := s.upsertEntry(ctx, class, sessionID, entry)
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

	switch dto.Source {
	case domain.AutoMarkSourceLive:
		if dto.RoomID != uuid.Nil {
			return s.autoMarkLiveRoom(ctx, class, sessionID, dto.RoomID)
		}
		return s.autoMarkLiveSession(ctx, class, sessionID)
	case domain.AutoMarkSourceOffline:
		if dto.RoomID == uuid.Nil {
			return nil, domain.NewValidationError(map[string]string{"room_id": "required for offline source"})
		}
		present, err := s.resolvePresentFromOffline(ctx, classID, dto.RoomID)
		if err != nil {
			return nil, err
		}
		result, err := s.writeAutoMark(ctx, class, sessionID, present, "offline_room")
		if err != nil {
			return nil, err
		}
		s.logger.Info("auto-mark attendance completed",
			"session_id", sessionID.String(),
			"source", string(dto.Source),
			"marked", result.Marked,
			"skipped", result.Skipped,
			"triggered_by", caller.UserID.String(),
		)
		return result, nil
	default:
		return nil, domain.ErrValidation
	}
}

// AutoMarkSessionLive runs live auto-mark for a whole session using the org's
// configured threshold. No caller authz — used by the worker / system path.
func (s *service) AutoMarkSessionLive(ctx context.Context, classID, sessionID uuid.UUID) (*domain.AutoMarkResult, error) {
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	session, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.ClassID != classID {
		return nil, domain.ErrNotFound
	}
	return s.autoMarkLiveSession(ctx, class, sessionID)
}

// autoMarkLiveSession aggregates participation across the session's live rooms,
// resolves present users against the org threshold, and writes attendance.
func (s *service) autoMarkLiveSession(ctx context.Context, class *domain.Class, sessionID uuid.UUID) (*domain.AutoMarkResult, error) {
	settings, err := s.orgSettings.GetByOrgID(ctx, class.OrganizationID)
	if err != nil {
		return nil, err
	}
	present, ok, err := s.resolvePresentLiveSession(ctx, sessionID, settings.AttendancePresentThresholdPercent)
	if err != nil {
		return nil, err
	}
	if !ok {
		s.logger.Warn("auto-mark skipped: no valid room duration",
			"class_id", class.ID.String(), "session_id", sessionID.String())
		return &domain.AutoMarkResult{Marked: 0, Skipped: 0}, nil
	}
	result, err := s.writeAutoMark(ctx, class, sessionID, present, "live_room")
	if err != nil {
		return nil, err
	}
	s.logger.Info("auto-mark attendance completed",
		"session_id", sessionID.String(),
		"source", "live_room",
		"percent", settings.AttendancePresentThresholdPercent,
		"marked", result.Marked,
		"skipped", result.Skipped,
	)
	return result, nil
}

// autoMarkLiveRoom scopes live auto-mark to a single room of the session so a
// side room (e.g. an optional Q&A) never counts toward the roll.
func (s *service) autoMarkLiveRoom(ctx context.Context, class *domain.Class, sessionID, roomID uuid.UUID) (*domain.AutoMarkResult, error) {
	room, err := s.liveRooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.ClassSessionID != sessionID {
		return nil, domain.ErrNotFound
	}
	settings, err := s.orgSettings.GetByOrgID(ctx, class.OrganizationID)
	if err != nil {
		return nil, err
	}
	present, ok, err := s.resolvePresentLiveRooms(ctx, []domain.LiveRoom{*room}, settings.AttendancePresentThresholdPercent)
	if err != nil {
		return nil, err
	}
	if !ok {
		s.logger.Warn("auto-mark skipped: no valid room duration",
			"class_id", class.ID.String(), "session_id", sessionID.String(), "room_id", roomID.String())
		return &domain.AutoMarkResult{Marked: 0, Skipped: 0}, nil
	}
	result, err := s.writeAutoMark(ctx, class, sessionID, present, "live_room")
	if err != nil {
		return nil, err
	}
	s.logger.Info("auto-mark attendance completed",
		"session_id", sessionID.String(),
		"source", "live_room",
		"room_id", roomID.String(),
		"percent", settings.AttendancePresentThresholdPercent,
		"marked", result.Marked,
		"skipped", result.Skipped,
	)
	return result, nil
}

// resolvePresentLiveSession aggregates participant durations across all of the
// session's live rooms that have valid actual start/end times. Returns ok=false
// when no room contributes a positive duration (caller should skip).
func (s *service) resolvePresentLiveSession(ctx context.Context, sessionID uuid.UUID, percent int) ([]uuid.UUID, bool, error) {
	rooms, err := s.liveRooms.ListByClassSession(ctx, sessionID)
	if err != nil {
		return nil, false, err
	}
	return s.resolvePresentLiveRooms(ctx, rooms, percent)
}

// resolvePresentLiveRooms sums participant durations across the given rooms
// (skipping rooms without a valid actual start/end window) and resolves present
// users against the percent threshold.
func (s *service) resolvePresentLiveRooms(ctx context.Context, rooms []domain.LiveRoom, percent int) ([]uuid.UUID, bool, error) {
	totalRoomSeconds := 0
	userSeconds := make(map[uuid.UUID]int)
	for _, room := range rooms {
		if room.ActualStartTime == nil || room.ActualEndTime == nil {
			continue
		}
		dur := int(room.ActualEndTime.Sub(*room.ActualStartTime).Seconds())
		if dur <= 0 {
			continue
		}
		totalRoomSeconds += dur
		participants, err := s.participants.ListAllByRoom(ctx, room.ID)
		if err != nil {
			return nil, false, err
		}
		for _, p := range participants {
			userSeconds[p.UserID] += p.TotalDurationSeconds
		}
	}
	present, ok := computePresentByPercent(totalRoomSeconds, userSeconds, percent)
	return present, ok, nil
}

// writeAutoMark sets present for the given users and absent for all other class
// members. Existing auto-marked rows are overwritten; manual rows are preserved.
func (s *service) writeAutoMark(ctx context.Context, class *domain.Class, sessionID uuid.UUID, presentUserIDs []uuid.UUID, source string) (*domain.AutoMarkResult, error) {
	members, err := s.members.ListAllByClass(ctx, class.ID)
	if err != nil {
		return nil, err
	}
	presentSet := make(map[uuid.UUID]bool, len(presentUserIDs))
	for _, id := range presentUserIDs {
		presentSet[id] = true
	}

	result := &domain.AutoMarkResult{}
	for _, member := range members {
		status := domain.AttendanceStatusAbsent
		if presentSet[member.UserID] {
			status = domain.AttendanceStatusPresent
		}

		existing, err := s.repo.FindBySessionAndUser(ctx, sessionID, member.UserID)
		switch {
		case err == nil && !existing.IsAutoMarked:
			// Manual override — never touch.
			result.Skipped++
		case err == nil && existing.IsAutoMarked:
			existing.Status = status
			existing.Remarks = "auto-marked from " + source
			if err := s.repo.Update(ctx, existing); err != nil {
				return nil, err
			}
			result.Marked++
		case errors.Is(err, domain.ErrNotFound):
			a := &domain.Attendance{
				OrganizationID: class.OrganizationID,
				ClassID:        class.ID,
				ClassSessionID: sessionID,
				UserID:         member.UserID,
				Status:         status,
				IsAutoMarked:   true,
				Remarks:        "auto-marked from " + source,
			}
			if err := s.repo.Create(ctx, a); err != nil {
				return nil, err
			}
			result.Marked++
		default:
			return nil, err
		}
	}
	return result, nil
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
	// Shallow changed-fields diff captured before mutating so the audit entry
	// records exactly what this update altered.
	changed := map[string]any{}
	if dto.Status != nil && *dto.Status != a.Status {
		changed["status"] = map[string]any{"from": string(a.Status), "to": string(*dto.Status)}
	}
	if dto.Status != nil {
		a.Status = *dto.Status
	}
	if dto.Remarks != nil {
		if *dto.Remarks != a.Remarks {
			changed["remarks"] = map[string]any{"from": a.Remarks, "to": *dto.Remarks}
		}
		a.Remarks = *dto.Remarks
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, a); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetAttendance,
			TargetID:    &a.ID,
			TargetLabel: class.Name,
			OrgID:       &class.OrganizationID,
			Metadata: map[string]any{
				"session_id": a.ClassSessionID.String(),
				"user_id":    a.UserID.String(),
				"changed":    changed,
			},
		})
	})
	if err != nil {
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
	return s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Delete(ctx, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditDeleted,
			TargetType:  domain.AuditTargetAttendance,
			TargetID:    &a.ID,
			TargetLabel: class.Name,
			OrgID:       &class.OrganizationID,
			Metadata: map[string]any{
				"session_id": a.ClassSessionID.String(),
				"user_id":    a.UserID.String(),
			},
		})
	})
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
	class, err := s.classes.FindByID(ctx, a.ClassID)
	if err != nil {
		return nil, err
	}
	scope, err := s.resolver.Scope(ctx, caller, class, domain.PermAttendanceViewAny)
	if err != nil {
		return nil, err
	}
	// Org-wide and class-wide viewers see any row; everyone else (enrolled
	// student) may only see their own attendance row.
	if scope == authz.ScopeAll || scope == authz.ScopeClass || caller.UserID == a.UserID {
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
	scope, err := s.resolver.Scope(ctx, caller, class, domain.PermAttendanceViewAny)
	if err != nil {
		return nil, 0, err
	}
	switch scope {
	case authz.ScopeNone:
		return nil, 0, domain.ErrForbidden
	case authz.ScopeOwn:
		// Enrolled student sees only their own attendance.
		userID := caller.UserID
		q.UserID = &userID
	}
	return s.repo.ListBySession(ctx, sessionID, q)
}

func (s *service) Matrix(ctx context.Context, classID uuid.UUID, q domain.ListAttendanceMatrixQuery) (*domain.AttendanceMatrixResult, error) {
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

	// Columns: every session, oldest first.
	sessions, _, err := s.sessions.ListByClass(ctx, classID, domain.ListClassSessionsQuery{
		ListParams: domain.ListParams{
			Page:     1,
			PageSize: attendanceMatrixMaxSessions,
			OrderBy:  "start_time",
			OrderDir: "asc",
		},
	})
	if err != nil {
		return nil, err
	}

	// Rows: paged students (User preloaded by the member repo).
	members, total, err := s.members.ListByClass(ctx, classID, q.ListParams)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	records, err := s.repo.ListByClassAndUsers(ctx, classID, userIDs)
	if err != nil {
		return nil, err
	}
	byUser := make(map[uuid.UUID]map[uuid.UUID]domain.Attendance, len(userIDs))
	for _, rec := range records {
		m, ok := byUser[rec.UserID]
		if !ok {
			m = make(map[uuid.UUID]domain.Attendance)
			byUser[rec.UserID] = m
		}
		m[rec.ClassSessionID] = rec
	}

	now := time.Now()
	startedCount := 0
	resSessions := make([]domain.AttendanceMatrixSession, 0, len(sessions))
	for _, sess := range sessions {
		if !sess.StartTime.After(now) {
			startedCount++
		}
		resSessions = append(resSessions, domain.AttendanceMatrixSession{
			ID:        sess.ID,
			Name:      sess.Name,
			StartTime: sess.StartTime,
		})
	}

	students := make([]domain.AttendanceMatrixStudent, 0, len(members))
	for _, m := range members {
		cells := make(map[uuid.UUID]domain.AttendanceMatrixCell)
		summary := domain.AttendanceMatrixSummary{StartedCount: startedCount}
		for sid, rec := range byUser[m.UserID] {
			cells[sid] = domain.AttendanceMatrixCell{
				ID:           rec.ID,
				Status:       rec.Status,
				IsAutoMarked: rec.IsAutoMarked,
			}
			switch rec.Status {
			case domain.AttendanceStatusPresent:
				summary.Present++
			case domain.AttendanceStatusAbsent:
				summary.Absent++
			case domain.AttendanceStatusLate:
				summary.Late++
			case domain.AttendanceStatusExcused:
				summary.Excused++
			}
		}
		if startedCount > 0 {
			summary.Rate = float64(summary.Present+summary.Late) / float64(startedCount)
		}
		students = append(students, domain.AttendanceMatrixStudent{
			UserID:  m.UserID,
			User:    m.User,
			Cells:   cells,
			Summary: summary,
		})
	}

	return &domain.AttendanceMatrixResult{
		Sessions: resSessions,
		Students: students,
		Total:    total,
		Page:     q.ListParams.Page,
		PageSize: q.ListParams.Limit(),
	}, nil
}

// ListMine returns the caller's own attendance history + a status summary.
// The summary is aggregated over the full filtered set, not the current page.
func (s *service) ListMine(ctx context.Context, q domain.ListMyAttendanceQuery) (*domain.MyAttendance, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	rows, total, err := s.repo.ListByUser(ctx, caller.UserID, q)
	if err != nil {
		return nil, fmt.Errorf("listing my attendance: %w", err)
	}
	summary, err := s.repo.SummarizeByUser(ctx, caller.UserID, q)
	if err != nil {
		return nil, fmt.Errorf("summarizing my attendance: %w", err)
	}
	return &domain.MyAttendance{
		Summary:  summary,
		Items:    rows,
		Total:    total,
		Page:     q.ListParams.Page,
		PageSize: q.ListParams.Limit(),
	}, nil
}
