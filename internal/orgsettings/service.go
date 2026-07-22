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
	tx     domain.Transactor
	audit  domain.AuditRecorder
	logger *slog.Logger
}

func NewService(repo domain.OrganizationSettingsRepository, tx domain.Transactor, audit domain.AuditRecorder, logger *slog.Logger) *service {
	return &service{repo: repo, tx: tx, audit: audit, logger: logger}
}

// Get returns the org's settings, falling back to defaults if no row exists.
// It is the HTTP-facing entrypoint: the caller must be an admin or belong to
// the requested org.
func (s *service) Get(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && (caller.OrgID == nil || *caller.OrgID != orgID) {
		return nil, domain.ErrForbidden
	}
	return s.get(ctx, orgID)
}

// get is the unguarded read used by both the HTTP-facing Get and the internal
// GetByOrgID provider. It performs no caller check.
func (s *service) get(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	settings, err := s.repo.FindByOrgID(ctx, orgID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NewDefaultOrganizationSettings(orgID), nil
		}
		return nil, err
	}
	return settings, nil
}

// GetByOrgID satisfies domain.OrganizationSettingsProvider. It is an internal
// provider called by other services whose ctx carries no caller, so it uses the
// unguarded read.
func (s *service) GetByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	return s.get(ctx, orgID)
}

func (s *service) Update(ctx context.Context, orgID uuid.UUID, dto domain.UpdateOrganizationSettingsDTO) (*domain.OrganizationSettings, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && (caller.OrgID == nil || *caller.OrgID != orgID) {
		return nil, domain.ErrForbidden
	}
	settings, err := s.repo.FindByOrgID(ctx, orgID)
	exists := true
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		settings = domain.NewDefaultOrganizationSettings(orgID)
		exists = false
	}
	// Shallow changed-fields diff so the audit entry records exactly what moved.
	changed := map[string]any{}
	if dto.AttendancePresentThresholdPercent != nil && *dto.AttendancePresentThresholdPercent != settings.AttendancePresentThresholdPercent {
		changed["attendance_present_threshold_percent"] = map[string]any{
			"from": settings.AttendancePresentThresholdPercent,
			"to":   *dto.AttendancePresentThresholdPercent,
		}
		settings.AttendancePresentThresholdPercent = *dto.AttendancePresentThresholdPercent
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if !exists {
			if err := s.repo.Create(ctx, settings); err != nil {
				return err
			}
		}
		if err := s.repo.Update(ctx, settings); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetOrgSettings,
			TargetID:    &settings.ID,
			TargetLabel: "organization settings",
			OrgID:       &settings.OrganizationID,
			Metadata:    map[string]any{"changed": changed},
		})
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info("organization settings updated", "org_id", orgID.String())
	return settings, nil
}

// AdminUpdate mutates superAdmin-only settings (SMS gate).
func (s *service) AdminUpdate(ctx context.Context, orgID uuid.UUID, dto domain.AdminUpdateOrgSettingsDTO) (*domain.OrganizationSettings, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return nil, domain.ErrForbidden
	}
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
	if dto.SMSEnabled != nil {
		settings.SMSEnabled = *dto.SMSEnabled
	}
	if err := s.repo.Update(ctx, settings); err != nil {
		return nil, err
	}
	s.logger.Info("organization settings admin-updated", "org_id", orgID.String())
	return settings, nil
}

var (
	_ domain.OrganizationSettingsService  = (*service)(nil)
	_ domain.OrganizationSettingsProvider = (*service)(nil)
)
