package organizations

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

// enqueuer is the async task port used to schedule background cleanup after an
// admin hard-deletes an org. Nil = skip enqueue (e.g. unit tests).
type enqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type service struct {
	repo         domain.OrganizationRepository
	userRepo     domain.UserRepository
	settingsRepo domain.OrganizationSettingsRepository
	redis        *redis.Client
	queue        enqueuer // may be nil
	tx           domain.Transactor
	audit        domain.AuditRecorder
	logger       *slog.Logger
}

func NewService(repo domain.OrganizationRepository, userRepo domain.UserRepository, settingsRepo domain.OrganizationSettingsRepository, rdb *redis.Client, queue enqueuer, tx domain.Transactor, audit domain.AuditRecorder, logger *slog.Logger) domain.OrganizationService {
	return &service{repo: repo, userRepo: userRepo, settingsRepo: settingsRepo, redis: rdb, queue: queue, tx: tx, audit: audit, logger: logger}
}

// bustTenant removes cached slug->org entries after a slug or status change so
// the tenant middleware re-resolves from the DB. No-op when redis is unset.
func (s *service) bustTenant(ctx context.Context, slugs ...string) {
	if s.redis == nil {
		return
	}
	for _, slug := range slugs {
		if slug == "" {
			continue
		}
		if err := cache.BustTenant(ctx, s.redis, slug); err != nil {
			s.logger.Warn("busting tenant cache", "slug", slug, "error", err)
		}
	}
}

func (s *service) Create(ctx context.Context, dto domain.CreateOrganizationDTO) (*domain.Organization, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	if err := domain.ValidateSlug(dto.Slug); err != nil {
		return nil, err
	}
	status := domain.OrganizationStatusActive
	if dto.Status != "" {
		status = dto.Status
	}
	org := &domain.Organization{
		Name:        dto.Name,
		Slug:        dto.Slug,
		Description: dto.Description,
		Status:      status,
	}
	if err := s.repo.Create(ctx, org); err != nil {
		return nil, err
	}
	if err := s.settingsRepo.Create(ctx, domain.NewDefaultOrganizationSettings(org.ID)); err != nil {
		return nil, fmt.Errorf("creating organization settings: %w", err)
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
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && (caller.OrgID == nil || *caller.OrgID != id) {
		return nil, domain.ErrForbidden
	}
	org, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	oldSlug := org.Slug
	// Shallow changed-fields diff so the audit entry records exactly what moved.
	changed := map[string]any{}
	if dto.Name != nil && *dto.Name != org.Name {
		changed["name"] = map[string]any{"from": org.Name, "to": *dto.Name}
		org.Name = *dto.Name
	}
	if dto.Slug != nil && *dto.Slug != org.Slug {
		if err := domain.ValidateSlug(*dto.Slug); err != nil {
			return nil, err
		}
		changed["slug"] = map[string]any{"from": org.Slug, "to": *dto.Slug}
		org.Slug = *dto.Slug
	}
	if dto.Description != nil && *dto.Description != org.Description {
		changed["description"] = map[string]any{"from": org.Description, "to": *dto.Description}
		org.Description = *dto.Description
	}
	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, org); err != nil {
			return err
		}
		return s.audit.Record(ctx, domain.AuditRecord{
			Action:      domain.AuditUpdated,
			TargetType:  domain.AuditTargetOrganization,
			TargetID:    &org.ID,
			TargetLabel: org.Name,
			OrgID:       &org.ID,
			Metadata:    map[string]any{"changed": changed},
		})
	})
	if err != nil {
		return nil, err
	}
	if org.Slug != oldSlug {
		s.bustTenant(ctx, oldSlug, org.Slug)
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
