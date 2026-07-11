package calendar

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/authz"
	"github.com/4H1R/zoora/internal/platform/cache"
)

type service struct {
	repo   domain.CalendarRepository
	rdb    *redis.Client // nil disables caching (unit tests)
	logger *slog.Logger
}

func NewService(repo domain.CalendarRepository, rdb *redis.Client, logger *slog.Logger) domain.CalendarService {
	return &service{repo: repo, rdb: rdb, logger: logger}
}

// ListEvents caches per (RBAC scope, window) behind a short TTL: the underlying
// four-table aggregation is the heaviest read in the codebase and the frontend
// refetches it on every month change.
func (s *service) ListEvents(ctx context.Context, r domain.CalendarRange) ([]domain.CalendarEvent, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	scope := authz.ListScope(caller, domain.PermClassesViewAny, domain.PermClassesUpdateAny)

	if s.rdb != nil {
		if events, err := cache.GetCalendarEvents(ctx, s.rdb, scope, r); err == nil {
			return events, nil
		}
	}

	events, err := s.repo.ListEvents(ctx, scope, r)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		if err := cache.SetCalendarEvents(ctx, s.rdb, scope, r, events); err != nil {
			s.logger.WarnContext(ctx, "caching calendar events", "error", err)
		}
	}
	return events, nil
}
