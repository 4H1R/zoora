package classes

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// requireAdmin pulls the caller from context and rejects anything that is not
// a verified super-admin. Defense-in-depth — the admin route group already
// mounts RequireAdmin middleware.
func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) AdminList(ctx context.Context, q domain.AdminListClassesQuery) ([]domain.Class, int64, error) {
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

func (s *service) AdminCreate(ctx context.Context, dto domain.AdminCreateClassDTO) (*domain.Class, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	class := &domain.Class{
		OrganizationID: dto.OrganizationID,
		UserID:         dto.UserID,
		Name:           dto.Name,
		Description:    dto.Description,
		TotalUsers:     dto.TotalUsers,
	}
	if err := s.repo.Create(ctx, class); err != nil {
		return nil, err
	}
	s.logger.Info("admin created class",
		"class_id", class.ID.String(),
		"created_by", caller.UserID.String(),
	)
	return class, nil
}

func (s *service) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateClassDTO) (*domain.Class, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	class, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Name != nil {
		class.Name = *dto.Name
	}
	if dto.Description != nil {
		class.Description = *dto.Description
	}
	if dto.TotalUsers != nil {
		class.TotalUsers = *dto.TotalUsers
	}
	if err := s.repo.Update(ctx, class); err != nil {
		return nil, err
	}
	return class, nil
}

func (s *service) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.repo.HardDelete(ctx, id); err != nil {
		return err
	}
	s.logger.Warn("admin hard-deleted class",
		"class_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) AdminHardDeleteSession(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.sessions.HardDelete(ctx, id); err != nil {
		return err
	}
	s.logger.Warn("admin hard-deleted class session",
		"session_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}
