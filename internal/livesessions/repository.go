package livesessions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) domain.LiveRoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) baseQuery(ctx context.Context) *gorm.DB {
	return database.DB(ctx, r.db).Model(&domain.LiveRoom{})
}

func (r *roomRepository) Create(ctx context.Context, room *domain.LiveRoom) error {
	if err := database.DB(ctx, r.db).Create(room).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("livesessions.roomRepository.Create: %w", err)
	}
	return nil
}

func (r *roomRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	var room domain.LiveRoom
	if err := r.baseQuery(ctx).First(&room, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("livesessions.roomRepository.FindByID: %w", err)
	}
	return &room, nil
}

func (r *roomRepository) ListByClassSession(ctx context.Context, sessionID uuid.UUID) ([]domain.LiveRoom, error) {
	var rooms []domain.LiveRoom
	if err := r.baseQuery(ctx).Where("class_session_id = ?", sessionID).Find(&rooms).Error; err != nil {
		return nil, fmt.Errorf("livesessions.roomRepository.ListByClassSession: %w", err)
	}
	return rooms, nil
}

func (r *roomRepository) Update(ctx context.Context, room *domain.LiveRoom) error {
	result := database.DB(ctx, r.db).Save(room)
	if result.Error != nil {
		if database.IsUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("livesessions.roomRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.LiveRoom{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("livesessions.roomRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) List(ctx context.Context, scope domain.LiveRoomListScope, p domain.ListParams) ([]domain.LiveRoom, int64, error) {
	db := database.DB(ctx, r.db)
	base := db.Model(&domain.LiveRoom{}).
		Preload("ClassSession").
		Preload("ClassSession.Class").
		Preload("ClassSession.Class.User")
	if scope.IncludeDeleted {
		base = base.Unscoped()
	}

	// RBAC + class filters resolve against class_sessions/classes. Use
	// IN-subqueries instead of JOIN+DISTINCT — GORM's Count emits
	// COUNT(DISTINCT live_rooms.*) which PostgreSQL rejects.
	if !scope.All {
		switch {
		case scope.TeacherID != nil && scope.MemberUserID != nil:
			sub := db.Table("class_sessions cs").
				Select("cs.id").
				Joins("JOIN classes c ON c.id = cs.class_id").
				Where(
					"c.user_id = ? OR c.id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
					*scope.TeacherID, *scope.MemberUserID,
				)
			base = base.Where("live_rooms.class_session_id IN (?)", sub)
		case scope.TeacherID != nil:
			sub := db.Table("class_sessions cs").
				Select("cs.id").
				Joins("JOIN classes c ON c.id = cs.class_id").
				Where("c.user_id = ?", *scope.TeacherID)
			base = base.Where("live_rooms.class_session_id IN (?)", sub)
		case scope.MemberUserID != nil:
			sub := db.Table("class_sessions cs").
				Select("cs.id").
				Where("cs.class_id IN (SELECT class_id FROM class_members WHERE user_id = ?)", *scope.MemberUserID)
			base = base.Where("live_rooms.class_session_id IN (?)", sub)
		}
	}

	if scope.OrganizationID != nil {
		sub := db.Table("class_sessions cs").
			Select("cs.id").
			Joins("JOIN classes c ON c.id = cs.class_id").
			Where("c.organization_id = ?", *scope.OrganizationID)
		base = base.Where("live_rooms.class_session_id IN (?)", sub)
	}
	if scope.ClassID != nil {
		sub := db.Table("class_sessions").
			Select("id").
			Where("class_id = ?", *scope.ClassID)
		base = base.Where("live_rooms.class_session_id IN (?)", sub)
	}
	if scope.Status != nil {
		base = base.Where("live_rooms.status = ?", *scope.Status)
	}
	if scope.ClassSessionID != nil {
		base = base.Where("live_rooms.class_session_id = ?", *scope.ClassSessionID)
	}

	var rooms []domain.LiveRoom
	total, err := listparams.Paginate(base, p, &rooms)
	if err != nil {
		return nil, 0, fmt.Errorf("livesessions.roomRepository.List: %w", err)
	}
	return rooms, total, nil
}

func (r *roomRepository) FindActiveRoomsWithStaleHost(ctx context.Context, staleDuration time.Duration) ([]domain.LiveRoom, error) {
	var rooms []domain.LiveRoom
	cutoff := time.Now().Add(-staleDuration)
	err := r.baseQuery(ctx).
		Where("status = ? AND host_last_seen_at < ?", domain.LiveRoomStatusActive, cutoff).
		Find(&rooms).Error
	if err != nil {
		return nil, fmt.Errorf("livesessions.roomRepository.FindActiveRoomsWithStaleHost: %w", err)
	}
	return rooms, nil
}

func (r *roomRepository) AdminList(ctx context.Context, q domain.AdminListLiveRoomsQuery) ([]domain.LiveRoom, int64, error) {
	db := database.DB(ctx, r.db)
	base := db.Model(&domain.LiveRoom{}).
		Preload("ClassSession").
		Preload("ClassSession.Class")
	if q.IncludeDeleted {
		base = base.Unscoped()
	}

	// Class/teacher filters target class_sessions/classes. Use subqueries instead
	// of JOIN+DISTINCT because GORM's Count emits COUNT(DISTINCT live_rooms.*)
	// which PostgreSQL rejects.
	if q.UserID != nil {
		sub := db.Table("class_sessions cs").
			Select("cs.id").
			Joins("JOIN classes c ON c.id = cs.class_id").
			Where("c.user_id = ?", *q.UserID)
		base = base.Where("live_rooms.class_session_id IN (?)", sub)
	}
	if q.ClassID != nil {
		sub := db.Table("class_sessions").
			Select("id").
			Where("class_id = ?", *q.ClassID)
		base = base.Where("live_rooms.class_session_id IN (?)", sub)
	}
	if q.Status != nil {
		base = base.Where("live_rooms.status = ?", *q.Status)
	}
	if q.ClassSessionID != nil {
		base = base.Where("live_rooms.class_session_id = ?", *q.ClassSessionID)
	}

	var rooms []domain.LiveRoom
	total, err := listparams.Paginate(base, q.ListParams, &rooms)
	if err != nil {
		return nil, 0, fmt.Errorf("livesessions.roomRepository.AdminList: %w", err)
	}
	return rooms, total, nil
}

func (r *roomRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Unscoped().Delete(&domain.LiveRoom{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("livesessions.roomRepository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roomRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	var room domain.LiveRoom
	if err := database.DB(ctx, r.db).Unscoped().Model(&domain.LiveRoom{}).First(&room, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("livesessions.roomRepository.FindByIDIncludingDeleted: %w", err)
	}
	return &room, nil
}

type participantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) domain.LiveParticipantRepository {
	return &participantRepository{db: db}
}

func (r *participantRepository) Create(ctx context.Context, p *domain.LiveParticipant) error {
	if err := database.DB(ctx, r.db).Create(p).Error; err != nil {
		return fmt.Errorf("livesessions.participantRepository.Create: %w", err)
	}
	return nil
}

func (r *participantRepository) FindActiveByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*domain.LiveParticipant, error) {
	var p domain.LiveParticipant
	err := database.DB(ctx, r.db).Model(&domain.LiveParticipant{}).
		Where("live_room_id = ? AND user_id = ? AND left_at IS NULL", roomID, userID).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("livesessions.participantRepository.FindActiveByRoomAndUser: %w", err)
	}
	return &p, nil
}

func (r *participantRepository) Update(ctx context.Context, p *domain.LiveParticipant) error {
	result := database.DB(ctx, r.db).Save(p)
	if result.Error != nil {
		return fmt.Errorf("livesessions.participantRepository.Update: %w", result.Error)
	}
	return nil
}

func (r *participantRepository) ListByRoom(ctx context.Context, roomID uuid.UUID, q domain.ListLiveParticipantsQuery) ([]domain.LiveParticipant, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.LiveParticipant{}).
		Where("live_room_id = ?", roomID)
	if q.ActiveOnly != nil {
		if *q.ActiveOnly {
			base = base.Where("left_at IS NULL")
		} else {
			base = base.Where("left_at IS NOT NULL")
		}
	}
	if q.UserID != nil {
		base = base.Where("user_id = ?", *q.UserID)
	}
	var participants []domain.LiveParticipant
	total, err := listparams.Paginate(base, q.ListParams, &participants)
	if err != nil {
		return nil, 0, fmt.Errorf("livesessions.participantRepository.ListByRoom: %w", err)
	}
	return participants, total, nil
}

func (r *participantRepository) ListAllByRoom(ctx context.Context, roomID uuid.UUID) ([]domain.LiveParticipant, error) {
	var participants []domain.LiveParticipant
	if err := database.DB(ctx, r.db).Model(&domain.LiveParticipant{}).
		Where("live_room_id = ?", roomID).
		Find(&participants).Error; err != nil {
		return nil, fmt.Errorf("livesessions.participantRepository.ListAllByRoom: %w", err)
	}
	return participants, nil
}

func (r *participantRepository) MarkAllLeft(ctx context.Context, roomID uuid.UUID, leftAt time.Time) error {
	result := database.DB(ctx, r.db).Model(&domain.LiveParticipant{}).
		Where("live_room_id = ? AND left_at IS NULL", roomID).
		Updates(map[string]interface{}{
			"left_at":                leftAt,
			"total_duration_seconds": gorm.Expr("EXTRACT(EPOCH FROM ? - joined_at)::int", leftAt),
		})
	if result.Error != nil {
		return fmt.Errorf("livesessions.participantRepository.MarkAllLeft: %w", result.Error)
	}
	return nil
}

type recordingRepository struct {
	db *gorm.DB
}

func NewRecordingRepository(db *gorm.DB) domain.LiveRecordingRepository {
	return &recordingRepository{db: db}
}

func (r *recordingRepository) Create(ctx context.Context, rec *domain.LiveRecording) error {
	if err := database.DB(ctx, r.db).Create(rec).Error; err != nil {
		return fmt.Errorf("livesessions.recordingRepository.Create: %w", err)
	}
	return nil
}

func (r *recordingRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRecording, error) {
	var rec domain.LiveRecording
	err := database.DB(ctx, r.db).Model(&domain.LiveRecording{}).First(&rec, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("livesessions.recordingRepository.FindByID: %w", err)
	}
	return &rec, nil
}

func (r *recordingRepository) FindActiveByRoom(ctx context.Context, roomID uuid.UUID) (*domain.LiveRecording, error) {
	var rec domain.LiveRecording
	err := database.DB(ctx, r.db).Model(&domain.LiveRecording{}).
		Where("live_room_id = ? AND status = ?", roomID, domain.LiveRecordingStatusStarted).
		First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("livesessions.recordingRepository.FindActiveByRoom: %w", err)
	}
	return &rec, nil
}

func (r *recordingRepository) Update(ctx context.Context, rec *domain.LiveRecording) error {
	result := database.DB(ctx, r.db).Save(rec)
	if result.Error != nil {
		return fmt.Errorf("livesessions.recordingRepository.Update: %w", result.Error)
	}
	return nil
}

func (r *recordingRepository) ListByRoom(ctx context.Context, roomID uuid.UUID, q domain.ListLiveRecordingsQuery) ([]domain.LiveRecording, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.LiveRecording{}).
		Where("live_room_id = ?", roomID)
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}
	var recs []domain.LiveRecording
	total, err := listparams.Paginate(base, q.ListParams, &recs)
	if err != nil {
		return nil, 0, fmt.Errorf("livesessions.recordingRepository.ListByRoom: %w", err)
	}
	return recs, total, nil
}
