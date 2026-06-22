package calendar

import (
	"context"
	"log/slog"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/authz"
)

type service struct {
	repo   domain.CalendarRepository
	logger *slog.Logger
}

func NewService(repo domain.CalendarRepository, logger *slog.Logger) domain.CalendarService {
	return &service{repo: repo, logger: logger}
}

func (s *service) ListEvents(ctx context.Context, r domain.CalendarRange) ([]domain.CalendarEvent, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	scope := authz.ListScope(caller, domain.PermClassesViewAny, domain.PermClassesUpdateAny)
	return s.repo.ListEvents(ctx, scope, r)
}
