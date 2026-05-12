package users

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo     domain.UserRepository
	roleRepo domain.RoleRepository
	logger   *slog.Logger
}

func NewService(repo domain.UserRepository, roleRepo domain.RoleRepository, logger *slog.Logger) domain.UserService {
	return &service{repo: repo, roleRepo: roleRepo, logger: logger}
}

func (s *service) isStaffRole(ctx context.Context, roleID uuid.UUID) bool {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return false
	}
	return role.IsPreset && role.Name == domain.PresetRoleStaff
}

func (s *service) Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin {
		dto.IsAdmin = false
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("users.service.Create hash: %w", err)
	}

	if dto.RoleID != nil {
		if !caller.IsAdmin && !caller.HasPermission(domain.PermRolesUpdate) {
			dto.RoleID = nil
		} else if !caller.IsAdmin && s.isStaffRole(ctx, *dto.RoleID) {
			return nil, domain.ErrForbidden
		}
	}

	user := &domain.User{
		OrganizationID: dto.OrganizationID,
		Username:       dto.Username,
		Name:           dto.Name,
		Password:       string(hashed),
		IsAdmin:        dto.IsAdmin,
		RoleID:         dto.RoleID,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("user created",
		"user_id", user.ID.String(),
		"created_by", caller.UserID.String(),
	)
	return user, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	return s.repo.FindByID(ctx, id)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateUserDTO) (*domain.User, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !caller.IsAdmin && caller.OrgID != nil {
		if user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
			return nil, domain.ErrForbidden
		}
	}

	if dto.Username != nil {
		user.Username = *dto.Username
	}
	if dto.Name != nil {
		user.Name = *dto.Name
	}
	if dto.RoleID != nil && (caller.IsAdmin || caller.HasPermission(domain.PermRolesUpdate)) {
		if !caller.IsAdmin && s.isStaffRole(ctx, *dto.RoleID) {
			return nil, domain.ErrForbidden
		}
		user.RoleID = dto.RoleID
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	if caller.UserID == id {
		return domain.ErrForbidden
	}
	if !caller.IsAdmin && caller.OrgID != nil {
		user, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
			return domain.ErrForbidden
		}
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("user deleted", "user_id", id.String(), "deleted_by", caller.UserID.String())
	return nil
}

func (s *service) List(ctx context.Context, p domain.ListParams) ([]domain.User, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	return s.repo.List(ctx, scope, p)
}

// resolveListScope maps a Caller into the role-resolved UserListScope the
// repository understands. Super-admins see all rows; everyone else is
// scoped to their organization.
func (s *service) resolveListScope(caller domain.Caller) domain.UserListScope {
	if caller.IsAdmin {
		return domain.UserListScope{All: true}
	}
	return domain.UserListScope{OrganizationID: caller.OrgID}
}

func (s *service) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) ChangePassword(ctx context.Context, id uuid.UUID, dto domain.ChangePasswordDTO) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.CurrentPassword)); err != nil {
		return domain.ErrUnauthorized
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("users.service.ChangePassword hash: %w", err)
	}
	user.Password = string(hashed)
	if err := s.repo.Update(ctx, user); err != nil {
		return err
	}
	s.logger.Info("password changed", "user_id", id.String())
	return nil
}

func (s *service) UpdateProfile(ctx context.Context, id uuid.UUID, dto domain.UpdateProfileDTO) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Name != "" {
		user.Name = dto.Name
	}
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) AssignRole(ctx context.Context, userID uuid.UUID, dto domain.AssignRoleDTO) (*domain.User, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermRolesUpdate) {
		return nil, domain.ErrForbidden
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.RoleID = &dto.RoleID
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("role assigned", "user_id", userID.String(), "role_id", dto.RoleID.String())
	return user, nil
}

func (s *service) RemoveRole(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermRolesUpdate) {
		return nil, domain.ErrForbidden
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.RoleID = nil
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("role removed", "user_id", userID.String())
	return user, nil
}
