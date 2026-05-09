package roles

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

var _ domain.RoleService = (*service)(nil)

type service struct {
	roleRepo domain.RoleRepository
	permRepo domain.PermissionRepository
	tx       domain.Transactor
	redis    *redis.Client
	logger   *slog.Logger
}

func NewService(
	roleRepo domain.RoleRepository,
	permRepo domain.PermissionRepository,
	tx domain.Transactor,
	rdb *redis.Client,
	logger *slog.Logger,
) domain.RoleService {
	return &service{
		roleRepo: roleRepo,
		permRepo: permRepo,
		tx:       tx,
		redis:    rdb,
		logger:   logger,
	}
}

func (s *service) Create(ctx context.Context, dto domain.CreateRoleDTO) (*domain.Role, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	if dto.IsPreset && !caller.IsAdmin {
		return nil, domain.ErrForbidden
	}
	if dto.IsPreset && dto.OrganizationID != nil {
		return nil, fmt.Errorf("preset roles must not have organization_id: %w", domain.ErrValidation)
	}
	if !dto.IsPreset && dto.OrganizationID == nil {
		return nil, fmt.Errorf("organization_id is required for non-preset roles: %w", domain.ErrValidation)
	}

	perms, err := s.permRepo.FindByIDs(ctx, dto.PermissionIDs)
	if err != nil {
		return nil, fmt.Errorf("roles.service.Create find permissions: %w", err)
	}
	if len(perms) != len(dto.PermissionIDs) {
		return nil, fmt.Errorf("some permission IDs not found: %w", domain.ErrValidation)
	}

	role := &domain.Role{
		OrganizationID: dto.OrganizationID,
		Name:           dto.Name,
		IsPreset:       dto.IsPreset,
	}

	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.roleRepo.Create(ctx, role); err != nil {
			return err
		}
		return s.roleRepo.SetPermissions(ctx, role.ID, dto.PermissionIDs)
	})
	if err != nil {
		return nil, err
	}

	role.Permissions = perms
	s.logger.Info("role created",
		"role_id", role.ID.String(),
		"name", role.Name,
		"created_by", caller.UserID.String(),
	)
	return role, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.FindByID(ctx, id)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateRoleDTO) (*domain.Role, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsPreset && !caller.IsAdmin {
		return nil, domain.ErrForbidden
	}
	if dto.IsPreset != nil && !caller.IsAdmin {
		return nil, domain.ErrForbidden
	}

	if dto.Name != "" {
		role.Name = dto.Name
	}
	if dto.IsPreset != nil {
		role.IsPreset = *dto.IsPreset
	}

	var newPerms []domain.Permission
	if dto.PermissionIDs != nil {
		newPerms, err = s.permRepo.FindByIDs(ctx, dto.PermissionIDs)
		if err != nil {
			return nil, fmt.Errorf("roles.service.Update find permissions: %w", err)
		}
		if len(newPerms) != len(dto.PermissionIDs) {
			return nil, fmt.Errorf("some permission IDs not found: %w", domain.ErrValidation)
		}
	}

	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.roleRepo.Update(ctx, role); err != nil {
			return err
		}
		if dto.PermissionIDs != nil {
			return s.roleRepo.SetPermissions(ctx, id, dto.PermissionIDs)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if newPerms != nil {
		role.Permissions = newPerms
	}

	if s.redis != nil {
		_ = cache.InvalidateRolePermissions(ctx, s.redis, id)
	}

	return role, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if role.IsPreset && !caller.IsAdmin {
		return domain.ErrForbidden
	}

	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.redis != nil {
		_ = cache.InvalidateRolePermissions(ctx, s.redis, id)
	}
	s.logger.Info("role deleted", "role_id", id.String())
	return nil
}

func (s *service) List(ctx context.Context, f domain.RoleFilter) ([]domain.Role, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && caller.OrgID != nil {
		f.OrganizationID = caller.OrgID
		f.IncludePreset = true
	}
	return s.roleRepo.List(ctx, f)
}

func (s *service) AdminList(ctx context.Context, f domain.AdminRoleFilter) ([]domain.Role, int64, error) {
	return s.roleRepo.AdminList(ctx, f)
}

func (s *service) Stats(ctx context.Context) (*domain.RoleStats, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	var orgID *uuid.UUID
	if !caller.IsAdmin && caller.OrgID != nil {
		orgID = caller.OrgID
	}
	return s.roleRepo.Stats(ctx, orgID)
}

func (s *service) AdminStats(ctx context.Context, orgID *uuid.UUID) (*domain.RoleStats, error) {
	return s.roleRepo.Stats(ctx, orgID)
}
