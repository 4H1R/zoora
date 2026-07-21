package users

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
	"github.com/4H1R/zoora/internal/platform/cache"
)

type service struct {
	repo     domain.UserRepository
	roleRepo domain.RoleRepository
	ent      entitlements.Service
	redis    *redis.Client
	tokens   domain.SessionTokenService
	logger   *slog.Logger
}

func NewService(repo domain.UserRepository, roleRepo domain.RoleRepository, ent entitlements.Service, rdb *redis.Client, tokens domain.SessionTokenService, logger *slog.Logger) domain.UserService {
	return &service{repo: repo, roleRepo: roleRepo, ent: ent, redis: rdb, tokens: tokens, logger: logger}
}

// bustUser drops the auth-middleware cache for a user after any change to their
// row (profile, role, admin flag, disable/enable, delete) so the next request
// reloads fresh state instead of waiting out the TTL. Best-effort: a cache
// error is logged, not propagated.
func (s *service) bustUser(ctx context.Context, id uuid.UUID) {
	if s.redis == nil {
		return
	}
	if err := cache.InvalidateUser(ctx, s.redis, id); err != nil {
		s.logger.Warn("invalidating user cache", "user_id", id.String(), "error", err)
	}
}

func (s *service) isManagerRole(ctx context.Context, roleID uuid.UUID) bool {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return false
	}
	return role.IsPreset && role.Name == domain.PresetRoleManager
}

func (s *service) Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin {
		dto.IsAdmin = false
		if caller.OrgID != nil {
			dto.OrganizationID = caller.OrgID
		}
	}

	// Enforce the org's seat limit (grandfather: blocks new creation only).
	if caller.OrgID != nil && s.ent != nil {
		if err := s.ent.CheckUserLimit(ctx, *caller.OrgID, caller.Ent); err != nil {
			return nil, err
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("users.service.Create hash: %w", err)
	}

	if dto.RoleID != nil {
		if !caller.IsAdmin && !caller.HasPermission(domain.PermRolesUpdate) {
			dto.RoleID = nil
		} else if !caller.IsAdmin && s.isManagerRole(ctx, *dto.RoleID) {
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
	return user, nil
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
		if !caller.IsAdmin && s.isManagerRole(ctx, *dto.RoleID) {
			return nil, domain.ErrForbidden
		}
		user.RoleID = dto.RoleID
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	s.bustUser(ctx, id)
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
	s.bustUser(ctx, id)
	s.logger.Info("user deleted", "user_id", id.String(), "deleted_by", caller.UserID.String())
	return nil
}

func (s *service) List(ctx context.Context, p domain.ListParams, disabled *bool) ([]domain.User, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	scope.Disabled = disabled
	return s.repo.List(ctx, scope, p)
}

// StatusCounts returns the caller-scoped user totals split by lockout state,
// backing the all/active/disabled tabs. Scope mirrors List exactly so the
// counts match what the list can show.
func (s *service) StatusCounts(ctx context.Context) (domain.UserStatusCounts, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.UserStatusCounts{}, domain.ErrForbidden
	}
	return s.repo.StatusCounts(ctx, s.resolveListScope(caller))
}

// resolveListScope maps a Caller into the role-resolved UserListScope the
// repository understands. Super-admins see all rows; users:view_any sees the
// whole organization; plain users:view is scoped to the caller's own row.
func (s *service) resolveListScope(caller domain.Caller) domain.UserListScope {
	if caller.IsAdmin {
		return domain.UserListScope{All: true}
	}
	if caller.HasPermission(domain.PermUsersViewAny) {
		return domain.UserListScope{OrganizationID: caller.OrgID}
	}
	uid := caller.UserID
	return domain.UserListScope{UserID: &uid}
}

func (s *service) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.FindByIDWithPermissions(ctx, id)
}

func (s *service) ChangePassword(ctx context.Context, id uuid.UUID, dto domain.ChangePasswordDTO) (string, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.CurrentPassword)); err != nil {
		return "", domain.ErrUnauthorized
	}
	if dto.NewPassword == dto.CurrentPassword {
		return "", domain.NewValidationError(map[string]string{"new_password": "must differ from current password"})
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("users.service.ChangePassword hash: %w", err)
	}
	user.Password = string(hashed)
	if err := s.repo.Update(ctx, user); err != nil {
		return "", err
	}
	s.bustUser(ctx, id)
	s.logger.Info("password changed", "user_id", id.String())

	// Log out every other device (tokens issued before now), then mint a fresh
	// token so this device — whose old token is now revoked — stays signed in.
	// Revoke first: the new token's IssuedAt is >= the revocation stamp, so it
	// survives the middleware's `IssuedAt < revokedAt` check.
	if s.tokens == nil {
		return "", nil
	}
	if err := s.tokens.RevokeUserSessions(ctx, id); err != nil {
		return "", fmt.Errorf("users.service.ChangePassword revoke: %w", err)
	}
	token, err := s.tokens.GenerateToken(id)
	if err != nil {
		return "", fmt.Errorf("users.service.ChangePassword token: %w", err)
	}
	return token, nil
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

	if !caller.IsAdmin && caller.OrgID != nil {
		if user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
			return nil, domain.ErrForbidden
		}
	}
	if !caller.IsAdmin && s.isManagerRole(ctx, dto.RoleID) {
		return nil, domain.ErrForbidden
	}

	user.RoleID = &dto.RoleID
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.bustUser(ctx, userID)
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

	if !caller.IsAdmin && caller.OrgID != nil {
		if user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
			return nil, domain.ErrForbidden
		}
	}

	user.RoleID = nil
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.bustUser(ctx, userID)
	s.logger.Info("role removed", "user_id", userID.String())
	return user, nil
}

// disableScopeCheck enforces the shared guards for Disable/Enable: caller must
// exist, cannot target themselves, must share the org when not admin, and a
// non-admin can never lock out an admin.
func (s *service) disableScopeCheck(ctx context.Context, id uuid.UUID) (domain.Caller, *domain.User, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.Caller{}, nil, domain.ErrForbidden
	}
	if caller.UserID == id {
		return domain.Caller{}, nil, domain.ErrForbidden
	}
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Caller{}, nil, err
	}
	if !caller.IsAdmin {
		if caller.OrgID == nil || user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
			return domain.Caller{}, nil, domain.ErrForbidden
		}
		if user.IsAdmin {
			return domain.Caller{}, nil, domain.ErrForbidden
		}
	}
	return caller, user, nil
}

func (s *service) Disable(ctx context.Context, id uuid.UUID, dto domain.DisableUserDTO) (*domain.User, error) {
	caller, user, err := s.disableScopeCheck(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.DisabledAt != nil {
		return user, nil // idempotent: preserve original timestamp
	}
	now := time.Now()
	by := caller.UserID
	user.DisabledAt = &now
	user.DisabledBy = &by
	if dto.Reason != "" {
		reason := dto.Reason
		user.DisabledReason = &reason
	}
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	s.bustUser(ctx, id)
	s.logger.Info("user disabled", "user_id", id.String(), "disabled_by", caller.UserID.String())
	return user, nil
}

func (s *service) Enable(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	caller, user, err := s.disableScopeCheck(ctx, id)
	if err != nil {
		return nil, err
	}
	user.DisabledAt = nil
	user.DisabledBy = nil
	user.DisabledReason = nil
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	s.bustUser(ctx, id)
	s.logger.Info("user enabled", "user_id", id.String(), "enabled_by", caller.UserID.String())
	return user, nil
}
