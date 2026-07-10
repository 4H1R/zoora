package organizations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) AdminList(ctx context.Context, q domain.AdminListOrganizationsQuery) ([]domain.Organization, int64, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, 0, err
	}
	if q.ListParams.Page < 1 {
		q.ListParams.Page = 1
	}
	if q.ListParams.PageSize <= 0 {
		q.ListParams.PageSize = domain.DefaultPageSize
	}
	return s.repo.AdminList(ctx, q)
}

func (s *service) AdminCreate(ctx context.Context, dto domain.AdminCreateOrganizationDTO) (*domain.Organization, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
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
	s.logger.Info("admin created organization",
		"org_id", org.ID.String(),
		"created_by", caller.UserID.String(),
	)
	return org, nil
}

func (s *service) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateOrganizationDTO) (*domain.Organization, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	org, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	oldSlug, oldStatus := org.Slug, org.Status
	if dto.Name != nil {
		org.Name = *dto.Name
	}
	if dto.Slug != nil {
		if err := domain.ValidateSlug(*dto.Slug); err != nil {
			return nil, err
		}
		org.Slug = *dto.Slug
	}
	if dto.Description != nil {
		org.Description = *dto.Description
	}
	if dto.Status != nil {
		org.Status = *dto.Status
	}
	if err := s.repo.Update(ctx, org); err != nil {
		return nil, err
	}
	if org.Slug != oldSlug || org.Status != oldStatus {
		s.bustTenant(ctx, oldSlug, org.Slug)
	}
	return org, nil
}

func (s *service) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.repo.HardDelete(ctx, id); err != nil {
		return err
	}
	s.logger.Warn("admin hard-deleted organization",
		"org_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	s.enqueueStorageCleanup(ctx, id)
	return nil
}

// enqueueStorageCleanup schedules deletion of the org's S3 objects under its
// key prefix. The DB org FK cascade already dropped the media rows; no FK
// covers the storage objects, so a background sweep purges them. Best-effort:
// a nil queue or enqueue failure is logged and never blocks the delete.
func (s *service) enqueueStorageCleanup(ctx context.Context, orgID uuid.UUID) {
	if s.queue == nil {
		return
	}
	payload, err := json.Marshal(domain.OrganizationCleanupPayload{OrganizationID: orgID})
	if err != nil {
		s.logger.Error("org storage cleanup enqueue: marshal payload", "org_id", orgID.String(), "error", err)
		return
	}
	if _, err := s.queue.Enqueue(asynq.NewTask(domain.TypeOrganizationCleanup, payload), asynq.Queue(domain.QueueMedia)); err != nil {
		s.logger.Error("org storage cleanup enqueue", "org_id", orgID.String(), "error", err)
	}
}

func (s *service) SetPlan(ctx context.Context, id uuid.UUID, dto domain.SetPlanDTO) (*domain.Organization, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !dto.Plan.Valid() {
		return nil, domain.NewValidationError(map[string]string{"plan": "unknown plan"})
	}
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, err // ErrNotFound bubbles
	}
	if err := s.repo.UpdatePlan(ctx, id, dto.Plan, dto.ExpiresAt); err != nil {
		return nil, err
	}
	// Invalidate the cached entitlement snapshot so the next request re-resolves.
	if s.redis != nil {
		if err := cache.InvalidateOrgPlan(ctx, s.redis, id); err != nil {
			s.logger.Warn("invalidating org plan cache", "org_id", id.String(), "error", err)
		}
	}
	s.logger.Info("admin set organization plan",
		"org_id", id.String(),
		"plan", string(dto.Plan),
		"set_by", caller.UserID.String(),
	)
	return s.repo.FindByID(ctx, id)
}

func (s *service) AdminRestore(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.repo.Restore(ctx, id); err != nil {
		return err
	}
	s.logger.Info("admin restored organization",
		"org_id", id.String(),
		"restored_by", caller.UserID.String(),
	)
	return nil
}
