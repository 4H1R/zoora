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

// --- LiveRoom repository ---

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

func (r *roomRepository) FindByClassSessionID(ctx context.Context, sessionID uuid.UUID) (*domain.LiveRoom, error) {
	var room domain.LiveRoom
	if err := r.baseQuery(ctx).First(&room, "class_session_id = ?", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("livesessions.roomRepository.FindByClassSessionID: %w", err)
	}
	return &room, nil
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
	base := database.DB(ctx, r.db).Model(&domain.LiveRoom{})
	if !scope.All {
		if scope.TeacherID != nil || scope.MemberUserID != nil {
			base = base.Joins("JOIN class_sessions cs ON cs.id = live_rooms.class_session_id").
				Joins("JOIN classes c ON c.id = cs.class_id")
		}
		switch {
		case scope.TeacherID != nil && scope.MemberUserID != nil:
			base = base.Where(
				"c.user_id = ? OR c.id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.TeacherID, *scope.MemberUserID,
			)
		case scope.TeacherID != nil:
			base = base.Where("c.user_id = ?", *scope.TeacherID)
		case scope.MemberUserID != nil:
			base = base.Where(
				"c.id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
				*scope.MemberUserID,
			)
		}
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
	base := database.DB(ctx, r.db).Model(&domain.LiveRoom{})
	if q.IncludeDeleted {
		base = base.Unscoped()
	}
	if q.Status != nil {
		base = base.Where("live_rooms.status = ?", *q.Status)
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

// --- LiveParticipant repository ---

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

func (r *participantRepository) ListByRoom(ctx context.Context, roomID uuid.UUID, p domain.ListParams) ([]domain.LiveParticipant, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.LiveParticipant{}).
		Where("live_room_id = ?", roomID)
	var participants []domain.LiveParticipant
	total, err := listparams.Paginate(base, p, &participants)
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

// --- LiveRecording repository ---

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

func (r *recordingRepository) ListByRoom(ctx context.Context, roomID uuid.UUID) ([]domain.LiveRecording, error) {
	var recs []domain.LiveRecording
	err := database.DB(ctx, r.db).Model(&domain.LiveRecording{}).
		Where("live_room_id = ?", roomID).
		Order("started_at DESC").
		Find(&recs).Error
	if err != nil {
		return nil, fmt.Errorf("livesessions.recordingRepository.ListByRoom: %w", err)
	}
	return recs, nil
}
