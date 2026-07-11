package tutorials

import (
	"context"
	"log/slog"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo   domain.TutorialRepository
	logger *slog.Logger
}

func NewService(repo domain.TutorialRepository, logger *slog.Logger) domain.TutorialService {
	return &service{repo: repo, logger: logger}
}

func (s *service) ListPublished(ctx context.Context) ([]domain.Tutorial, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListPublished(ctx)
}
