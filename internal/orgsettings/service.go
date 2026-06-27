package orgsettings

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo   domain.OrganizationSettingsRepository
	logger *slog.Logger
}

func NewService(repo domain.OrganizationSettingsRepository, logger *slog.Logger) *service {
	return &service{repo: repo, logger: logger}
}

// Get returns the org's settings, falling back to defaults if no row exists.
func (s *service) Get(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	settings, err := s.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NewDefaultOrganizationSettings(orgID), nil
		}
		return nil, err
	}
	return settings, nil
}

// GetByOrgID satisfies domain.OrganizationSettingsProvider.
func (s *service) GetByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	return s.Get(ctx, orgID)
}

func (s *service) Update(ctx context.Context, orgID uuid.UUID, dto domain.UpdateOrganizationSettingsDTO) (*domain.OrganizationSettings, error) {
	settings, err := s.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		settings = domain.NewDefaultOrganizationSettings(orgID)
		if err := s.repo.Create(ctx, settings); err != nil {
			return nil, err
		}
	}
	if dto.AttendancePresentThresholdPercent != nil {
		settings.AttendancePresentThresholdPercent = *dto.AttendancePresentThresholdPercent
	}
	if err := s.repo.Update(ctx, settings); err != nil {
		return nil, err
	}
	s.logger.Info("organization settings updated", "org_id", orgID.String())
	return settings, nil
}

var (
	_ domain.OrganizationSettingsService  = (*service)(nil)
	_ domain.OrganizationSettingsProvider = (*service)(nil)
)
