package practices

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	rooms    domain.PracticeRoomRepository
	subs     domain.PracticeSubmissionRepository
	sessions domain.ClassSessionRepository
	classes  domain.ClassRepository
	members  domain.ClassMemberRepository
	tx       domain.Transactor
	audit    domain.AuditRecorder
	logger   *slog.Logger
}

func NewService(
	rooms domain.PracticeRoomRepository,
	subs domain.PracticeSubmissionRepository,
	sessions domain.ClassSessionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	tx domain.Transactor,
	audit domain.AuditRecorder,
	logger *slog.Logger,
) domain.PracticeService {
	return &service{
		rooms:    rooms,
		subs:     subs,
		sessions: sessions,
		classes:  classes,
		members:  members,
		tx:       tx,
		audit:    audit,
		logger:   logger,
	}
}

func canManageRoom(caller domain.Caller, room *domain.PracticeRoom) bool {
	return caller.CanManage(room.UserID, domain.PermPracticesUpdateAny)
}

func canDeleteRoom(caller domain.Caller, room *domain.PracticeRoom) bool {
	return caller.CanManage(room.UserID, domain.PermPracticesDeleteAny)
}

func (s *service) canViewRoom(ctx context.Context, caller domain.Caller, room *domain.PracticeRoom) (bool, error) {
	if canManageRoom(caller, room) {
		return true, nil
	}
	if caller.HasPermission(domain.PermPracticesViewAny) {
		return true, nil
	}
	return s.members.Exists(ctx, room.ClassID, caller.UserID)
}

func (s *service) CreateRoom(ctx context.Context, dto domain.CreatePracticeRoomDTO) (*domain.PracticeRoom, error) {
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
	if !caller.IsAdmin && !caller.HasPermission(domain.PermPracticesCreateAny) && caller.UserID != class.UserID {
		return nil, domain.ErrForbidden
	}
	room := &domain.PracticeRoom{
		OrganizationID: class.OrganizationID,
		ClassID:        class.ID,
		ClassSessionID: dto.ClassSessionID,
		UserID:         caller.UserID,
		Title:          dto.Title,
		Content:        dto.Content,
		MaxScore:       dto.MaxScore,
		StartTime:      dto.StartTime,
		EndTime:        dto.EndTime,
		Attachments:    dto.Attachments,
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.rooms.Create(ctx, room); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditCreated,
			TargetType:  domain.AuditTargetPractice,
			TargetID:    &room.ID,
			TargetLabel: room.Title,
			OrgID:       &room.OrganizationID,
			Metadata:    map[string]any{"class_id": room.ClassID.String()},
		})
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info("practice room created",
		"room_id", room.ID.String(),
		"class_id", room.ClassID.String(),
		"created_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) GetRoom(ctx context.Context, id uuid.UUID) (*domain.PracticeRoom, error) {
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
	return room, nil
}

func (s *service) UpdateRoom(ctx context.Context, id uuid.UUID, dto domain.UpdatePracticeRoomDTO) (*domain.PracticeRoom, error) {
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
	// Shallow changed-fields diff captured before mutating so the audit entry
	// records exactly what this update altered (attachments summarized, not dumped).
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
	if dto.Content != nil {
		setChanged("content_changed", false, true)
		room.Content = *dto.Content
	}
	if dto.MaxScore != nil {
		setChanged("max_score", room.MaxScore, *dto.MaxScore)
		room.MaxScore = *dto.MaxScore
	}
	if dto.StartTime != nil {
		setChanged("start_time", room.StartTime, *dto.StartTime)
		room.StartTime = *dto.StartTime
	}
	if dto.EndTime != nil {
		setChanged("end_time", room.EndTime, *dto.EndTime)
		room.EndTime = *dto.EndTime
	}
	if dto.Attachments != nil {
		changed["attachments"] = len(*dto.Attachments)
		room.Attachments = *dto.Attachments
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.rooms.Update(ctx, room); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetPractice,
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
			TargetType:  domain.AuditTargetPractice,
			TargetID:    &id,
			TargetLabel: room.Title,
			OrgID:       &room.OrganizationID,
			Metadata:    map[string]any{"class_id": room.ClassID.String()},
		})
	})
	if err != nil {
		return err
	}
	s.logger.Info("practice room deleted",
		"room_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) ListRooms(ctx context.Context, q domain.ListPracticeRoomsQuery) ([]domain.PracticeRoomView, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	q.ViewerID = caller.UserID
	if !canListDeleted(caller) {
		q.IncludeDeleted = false
	}
	rooms, total, err := s.rooms.List(ctx, scope, q)
	if err != nil {
		return nil, 0, err
	}
	if len(rooms) == 0 {
		return []domain.PracticeRoomView{}, total, nil
	}

	roomIDs := make([]uuid.UUID, len(rooms))
	classIDSet := make(map[uuid.UUID]struct{})
	for i, r := range rooms {
		roomIDs[i] = r.ID
		classIDSet[r.ClassID] = struct{}{}
	}
	classIDs := make([]uuid.UUID, 0, len(classIDSet))
	for id := range classIDSet {
		classIDs = append(classIDs, id)
	}

	mySubsList, err := s.subs.ListByRoomsAndUser(ctx, roomIDs, caller.UserID)
	if err != nil {
		return nil, 0, err
	}
	mySubs := make(map[uuid.UUID]*domain.PracticeSubmission, len(mySubsList))
	for i := range mySubsList {
		mySubs[mySubsList[i].PracticeRoomID] = &mySubsList[i]
	}

	isManager := caller.HasAny(domain.PermPracticesViewAny, domain.PermPracticesUpdateAny, domain.PermPracticesGrade)
	var stats map[uuid.UUID]domain.PracticeRoomStats
	var memberCounts map[uuid.UUID]int64
	if isManager {
		if stats, err = s.subs.CountsByRooms(ctx, roomIDs); err != nil {
			return nil, 0, err
		}
		if memberCounts, err = s.rooms.MemberCountsByClasses(ctx, classIDs); err != nil {
			return nil, 0, err
		}
	}

	memberClassIDs, err := s.rooms.ViewerMemberClasses(ctx, caller.UserID, classIDs)
	if err != nil {
		return nil, 0, err
	}
	memberClasses := make(map[uuid.UUID]bool, len(memberClassIDs))
	for _, id := range memberClassIDs {
		memberClasses[id] = true
	}

	now := time.Now()
	views := make([]domain.PracticeRoomView, len(rooms))
	for i := range rooms {
		r := rooms[i]
		sub := mySubs[r.ID]
		v := domain.PracticeRoomView{
			PracticeRoom: r,
			MySubmission: sub,
			Mine:         sub != nil,
			Status:       derivePracticeStatus(now, r, sub),
			CanGrade:     canManageRoom(caller, &r),
		}
		v.CanSubmit = memberClasses[r.ClassID] && !v.CanGrade && sub == nil &&
			!now.Before(r.StartTime) && !now.After(r.EndTime)
		if v.CanGrade && stats != nil {
			st := stats[r.ID]
			st.MemberCount = memberCounts[r.ClassID]
			v.Stats = &st
		}
		views[i] = v
	}
	return views, total, nil
}

func derivePracticeStatus(now time.Time, r domain.PracticeRoom, sub *domain.PracticeSubmission) string {
	if sub != nil {
		if sub.Score != nil {
			return domain.PracticeStatusGraded
		}
		return domain.PracticeStatusSubmitted
	}
	if now.Before(r.StartTime) {
		return domain.PracticeStatusUpcoming
	}
	if now.After(r.EndTime) {
		return domain.PracticeStatusMissed
	}
	return domain.PracticeStatusToSubmit
}

func (s *service) resolveListScope(caller domain.Caller) domain.PracticeRoomListScope {
	if caller.IsAdmin {
		return domain.PracticeRoomListScope{All: true}
	}
	if caller.HasPermission(domain.PermPracticesViewAny) || caller.HasPermission(domain.PermPracticesUpdateAny) {
		return domain.PracticeRoomListScope{All: true, OrganizationID: caller.OrgID}
	}
	userID := caller.UserID
	return domain.PracticeRoomListScope{
		OwnerID:      &userID,
		MemberUserID: &userID,
	}
}

func canListDeleted(caller domain.Caller) bool {
	return caller.HasAny(domain.PermPracticesUpdateAny)
}

func (s *service) Submit(ctx context.Context, roomID uuid.UUID, dto domain.CreatePracticeSubmissionDTO) (*domain.PracticeSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	isMember, err := s.members.Exists(ctx, room.ClassID, caller.UserID)
	if err != nil {
		return nil, err
	}
	if !isMember && !canManageRoom(caller, room) {
		return nil, domain.ErrForbidden
	}
	now := time.Now()
	if now.Before(room.StartTime) || now.After(room.EndTime) {
		return nil, domain.NewValidationError(map[string]string{
			"time": "submissions only accepted between start_time and end_time",
		})
	}
	if existing, err := s.subs.FindByRoomAndUser(ctx, roomID, caller.UserID); err == nil && existing != nil {
		return nil, domain.ErrConflict
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	sub := &domain.PracticeSubmission{
		PracticeRoomID: roomID,
		UserID:         caller.UserID,
		Content:        dto.Content,
		SubmittedAt:    now,
		Attachments:    dto.Attachments,
	}
	if err := s.subs.Create(ctx, sub); err != nil {
		return nil, err
	}
	s.logger.Info("practice submission created",
		"submission_id", sub.ID.String(),
		"room_id", roomID.String(),
		"user_id", caller.UserID.String(),
	)
	return sub, nil
}

func (s *service) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.PracticeSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	sub, err := s.subs.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if caller.UserID == sub.UserID {
		return sub, nil
	}
	room, err := s.rooms.FindByID(ctx, sub.PracticeRoomID)
	if err != nil {
		return nil, err
	}
	if !canManageRoom(caller, room) {
		return nil, domain.ErrForbidden
	}
	return sub, nil
}

func (s *service) ListSubmissions(ctx context.Context, roomID uuid.UUID, q domain.ListPracticeSubmissionsQuery) ([]domain.PracticeSubmission, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, 0, err
	}
	if !canManageRoom(caller, room) {
		return nil, 0, domain.ErrForbidden
	}
	return s.subs.ListByRoom(ctx, roomID, q.ListParams)
}

func (s *service) Grade(ctx context.Context, submissionID uuid.UUID, dto domain.GradePracticeSubmissionDTO) (*domain.PracticeSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	sub, err := s.subs.FindByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	room, err := s.rooms.FindByID(ctx, sub.PracticeRoomID)
	if err != nil {
		return nil, err
	}
	if !canManageRoom(caller, room) {
		return nil, domain.ErrForbidden
	}
	if dto.Score != nil {
		if room.MaxScore > 0 && *dto.Score > room.MaxScore {
			return nil, domain.NewValidationError(map[string]string{
				"score": fmt.Sprintf("score cannot exceed max_score (%g)", room.MaxScore),
			})
		}
		sub.Score = dto.Score
	}
	if dto.TeacherComment != nil {
		sub.TeacherComment = *dto.TeacherComment
	}
	meta := map[string]any{
		"submission_id": submissionID.String(),
		"student_id":    sub.UserID.String(),
	}
	if sub.Score != nil {
		meta["score"] = *sub.Score
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.subs.Update(ctx, sub); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditGraded,
			TargetType:  domain.AuditTargetPractice,
			TargetID:    &room.ID,
			TargetLabel: room.Title,
			OrgID:       &room.OrganizationID,
			Metadata:    meta,
		})
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info("practice submission graded",
		"submission_id", submissionID.String(),
		"graded_by", caller.UserID.String(),
	)
	return sub, nil
}
