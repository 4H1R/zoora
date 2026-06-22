package calendar

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type calendarRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.CalendarRepository {
	return &calendarRepository{db: db}
}

// applyScope appends the role-resolved class filter. All four sub-queries
// alias the classes table as "c", so this is shared verbatim across them.
func applyScope(q *gorm.DB, scope domain.ClassListScope) *gorm.DB {
	if scope.All {
		if scope.OrganizationID != nil {
			q = q.Where("c.organization_id = ?", *scope.OrganizationID)
		}
		return q
	}
	switch {
	case scope.TeacherID != nil && scope.MemberUserID != nil:
		q = q.Where(
			"c.user_id = ? OR c.id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
			*scope.TeacherID, *scope.MemberUserID,
		)
	case scope.TeacherID != nil:
		q = q.Where("c.user_id = ?", *scope.TeacherID)
	case scope.MemberUserID != nil:
		q = q.Where(
			"c.id IN (SELECT class_id FROM class_members WHERE user_id = ?)",
			*scope.MemberUserID,
		)
	}
	return q
}

func (r *calendarRepository) ListEvents(ctx context.Context, scope domain.ClassListScope, rng domain.CalendarRange) ([]domain.CalendarEvent, error) {
	var events []domain.CalendarEvent

	live, err := r.liveEvents(ctx, scope, rng)
	if err != nil {
		return nil, err
	}
	events = append(events, live...)

	quiz, err := r.quizEvents(ctx, scope, rng)
	if err != nil {
		return nil, err
	}
	events = append(events, quiz...)

	practice, err := r.practiceEvents(ctx, scope, rng)
	if err != nil {
		return nil, err
	}
	events = append(events, practice...)

	offline, err := r.offlineEvents(ctx, scope, rng)
	if err != nil {
		return nil, err
	}
	events = append(events, offline...)

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime.Before(events[j].StartTime)
	})
	return events, nil
}

func (r *calendarRepository) liveEvents(ctx context.Context, scope domain.ClassListScope, rng domain.CalendarRange) ([]domain.CalendarEvent, error) {
	var rows []domain.CalendarEvent
	q := database.DB(ctx, r.db).
		Table("live_rooms AS lr").
		Select(`lr.id AS id, 'live' AS type, lr.name AS title,
			c.id AS class_id, c.name AS class_name, cs.id AS class_session_id,
			lr.id AS entity_id, lr.scheduled_start_time AS start_time, lr.actual_end_time AS end_time`).
		Joins("JOIN class_sessions cs ON cs.id = lr.class_session_id AND cs.deleted_at IS NULL").
		Joins("JOIN classes c ON c.id = cs.class_id AND c.deleted_at IS NULL").
		Where("lr.deleted_at IS NULL").
		Where("lr.scheduled_start_time IS NOT NULL").
		Where("lr.scheduled_start_time >= ? AND lr.scheduled_start_time <= ?", rng.From, rng.To)
	q = applyScope(q, scope)
	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("calendar.repository.liveEvents: %w", err)
	}
	return rows, nil
}

func (r *calendarRepository) quizEvents(ctx context.Context, scope domain.ClassListScope, rng domain.CalendarRange) ([]domain.CalendarEvent, error) {
	var rows []domain.CalendarEvent
	q := database.DB(ctx, r.db).
		Table("quiz_rooms AS qr").
		Select(`qr.id AS id, 'quiz' AS type, qz.title AS title,
			c.id AS class_id, c.name AS class_name, cs.id AS class_session_id,
			qz.id AS entity_id, qr.started_at AS start_time, qr.ended_at AS end_time`).
		Joins("JOIN quizzes qz ON qz.id = qr.quiz_id AND qz.deleted_at IS NULL").
		Joins("JOIN class_sessions cs ON cs.id = qr.class_session_id AND cs.deleted_at IS NULL").
		Joins("JOIN classes c ON c.id = cs.class_id AND c.deleted_at IS NULL").
		Where("qr.started_at IS NOT NULL").
		Where("qr.started_at >= ? AND qr.started_at <= ?", rng.From, rng.To)
	q = applyScope(q, scope)
	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("calendar.repository.quizEvents: %w", err)
	}
	return rows, nil
}

func (r *calendarRepository) practiceEvents(ctx context.Context, scope domain.ClassListScope, rng domain.CalendarRange) ([]domain.CalendarEvent, error) {
	var rows []domain.CalendarEvent
	q := database.DB(ctx, r.db).
		Table("practice_rooms AS pr").
		Select(`pr.id AS id, 'practice' AS type, pr.title AS title,
			c.id AS class_id, c.name AS class_name, cs.id AS class_session_id,
			cs.id AS entity_id, pr.start_time AS start_time, pr.end_time AS end_time`).
		Joins("JOIN class_sessions cs ON cs.id = pr.class_session_id AND cs.deleted_at IS NULL").
		Joins("JOIN classes c ON c.id = cs.class_id AND c.deleted_at IS NULL").
		Where("pr.deleted_at IS NULL").
		Where("pr.start_time >= ? AND pr.start_time <= ?", rng.From, rng.To)
	q = applyScope(q, scope)
	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("calendar.repository.practiceEvents: %w", err)
	}
	return rows, nil
}

func (r *calendarRepository) offlineEvents(ctx context.Context, scope domain.ClassListScope, rng domain.CalendarRange) ([]domain.CalendarEvent, error) {
	var rows []domain.CalendarEvent
	q := database.DB(ctx, r.db).
		Table("offline_rooms AS ofr").
		Select(`ofr.id AS id, 'offline' AS type, ofr.title AS title,
			c.id AS class_id, c.name AS class_name, cs.id AS class_session_id,
			ofr.id AS entity_id, ofr.published_at AS start_time, NULL AS end_time`).
		Joins("JOIN class_sessions cs ON cs.id = ofr.class_session_id AND cs.deleted_at IS NULL").
		Joins("JOIN classes c ON c.id = cs.class_id AND c.deleted_at IS NULL").
		Where("ofr.deleted_at IS NULL").
		Where("ofr.published_at IS NOT NULL").
		Where("ofr.published_at >= ? AND ofr.published_at <= ?", rng.From, rng.To)
	q = applyScope(q, scope)
	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("calendar.repository.offlineEvents: %w", err)
	}
	return rows, nil
}
