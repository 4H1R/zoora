package organizations

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo     domain.OrganizationRepository
	userRepo domain.UserRepository
	logger   *slog.Logger
}

func NewService(repo domain.OrganizationRepository, userRepo domain.UserRepository, logger *slog.Logger) domain.OrganizationService {
	return &service{repo: repo, userRepo: userRepo, logger: logger}
}

func (s *service) Create(ctx context.Context, dto domain.CreateOrganizationDTO) (*domain.Organization, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	status := domain.OrganizationStatusActive
	if dto.Status != "" {
		status = dto.Status
	}
	org := &domain.Organization{
		Name:        dto.Name,
		Description: dto.Description,
		Status:      status,
	}
	if err := s.repo.Create(ctx, org); err != nil {
		return nil, err
	}
	s.logger.Info("organization created", "org_id", org.ID.String())
	return org, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && (caller.OrgID == nil || *caller.OrgID != id) {
		return nil, domain.ErrForbidden
	}
	return s.repo.FindByID(ctx, id)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateOrganizationDTO) (*domain.Organization, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	org, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Name != nil {
		org.Name = *dto.Name
	}
	if dto.Description != nil {
		org.Description = *dto.Description
	}
	if err := s.repo.Update(ctx, org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("organization deleted", "org_id", id.String())
	return nil
}

func (s *service) List(ctx context.Context, f domain.OrganizationFilter) ([]domain.Organization, int64, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, 0, domain.ErrForbidden
	}
	orgs, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("organizations.service.List: %w", err)
	}
	return orgs, total, nil
}

func (s *service) GetStats(ctx context.Context) (*domain.OrganizationStats, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("organizations.service.GetStats: %w", err)
	}
	return stats, nil
}
