package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
)

// calendarTTL is short because the events endpoint aggregates four tables
// (live rooms, quizzes, practices, offlines) that mutate across many features;
// per-key invalidation would be cross-cutting and brittle. A brief TTL caps
// staleness while collapsing the repeated month-view fetches that dominate
// load, and the query is the heaviest read in the codebase.
const calendarTTL = 60 * time.Second

func uuidToken(id *uuid.UUID) string {
	if id == nil {
		return "-"
	}
	return id.String()
}

// calendarKey is deterministic in the RBAC scope and the requested window so
// two identical requests share a cache entry. Scope fields fully determine
// which events a caller may see (see authz.ListScope).
func calendarKey(scope domain.ClassListScope, r domain.CalendarRange) string {
	return fmt.Sprintf("calendar:%t:%s:%s:%s:%t:%d:%d",
		scope.All,
		uuidToken(scope.OrganizationID),
		uuidToken(scope.TeacherID),
		uuidToken(scope.MemberUserID),
		scope.IncludeDeleted,
		r.From.Unix(), r.To.Unix(),
	)
}

// GetCalendarEvents returns the cached events for a (scope, range). Miss or
// decode failure returns an error — callers fall back to the repository.
func GetCalendarEvents(ctx context.Context, rdb *redis.Client, scope domain.ClassListScope, r domain.CalendarRange) ([]domain.CalendarEvent, error) {
	data, err := rdb.Get(ctx, calendarKey(scope, r)).Bytes()
	if err != nil {
		return nil, err
	}
	var events []domain.CalendarEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("unmarshaling cached calendar events: %w", err)
	}
	return events, nil
}

// SetCalendarEvents caches the events for a (scope, range).
func SetCalendarEvents(ctx context.Context, rdb *redis.Client, scope domain.ClassListScope, r domain.CalendarRange, events []domain.CalendarEvent) error {
	data, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("marshaling calendar events: %w", err)
	}
	return rdb.Set(ctx, calendarKey(scope, r), data, calendarTTL).Err()
}
