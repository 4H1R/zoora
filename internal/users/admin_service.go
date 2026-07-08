package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
)

func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) AdminGetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindByIDIncludingDeleted(ctx, id)
}

func (s *service) AdminList(ctx context.Context, q domain.AdminListUsersQuery) ([]domain.User, int64, error) {
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

func (s *service) AdminCreate(ctx context.Context, dto domain.AdminCreateUserDTO) (*domain.User, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("users.service.AdminCreate hash: %w", err)
	}

	user := &domain.User{
		OrganizationID: dto.OrganizationID,
		Username:       dto.Username,
		Name:           dto.Name,
		Password:       string(hashed),
		IsAdmin:        dto.IsAdmin,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("admin created user",
		"user_id", user.ID.String(),
		"created_by", caller.UserID.String(),
		"is_admin", dto.IsAdmin,
	)
	return user, nil
}

func (s *service) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateUserDTO) (*domain.User, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Username != nil {
		user.Username = *dto.Username
	}
	if dto.Name != nil {
		user.Name = *dto.Name
	}
	if dto.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*dto.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("users.service.AdminUpdate hash: %w", err)
		}
		user.Password = string(hashed)
	}
	if dto.IsAdmin != nil {
		user.IsAdmin = *dto.IsAdmin
	}
	if dto.RoleID != nil {
		user.RoleID = dto.RoleID
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.bustUser(ctx, id)
	s.logger.Info("admin updated user",
		"user_id", id.String(),
		"updated_by", caller.UserID.String(),
	)
	return user, nil
}

func (s *service) AdminForceResetPassword(ctx context.Context, id uuid.UUID, dto domain.AdminForceResetPasswordDTO) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("users.service.AdminForceResetPassword hash: %w", err)
	}
	user.Password = string(hashed)
	if err := s.repo.Update(ctx, user); err != nil {
		return err
	}

	s.bustUser(ctx, id)
	s.logger.Info("admin force-reset password",
		"user_id", id.String(),
		"reset_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if caller.UserID == id {
		return domain.ErrForbidden
	}
	if err := s.repo.HardDelete(ctx, id); err != nil {
		return err
	}
	s.bustUser(ctx, id)
	s.logger.Warn("admin hard-deleted user",
		"user_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}
