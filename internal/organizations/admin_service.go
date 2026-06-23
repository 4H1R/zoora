package organizations

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
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
	return nil
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
