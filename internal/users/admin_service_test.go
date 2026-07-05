package users_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/users"
)

func adminUserCtx() (context.Context, uuid.UUID) {
	adminID := uuid.New()
	return domain.WithCaller(context.Background(), domain.Caller{UserID: adminID, IsAdmin: true}), adminID
}

func nonAdminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New()})
}

func TestAdminList_Forbidden_WhenNotAdmin(t *testing.T) {
	svc := users.NewService(&mockUserRepo{}, &mockRoleRepo{}, nil, slog.Default())
	_, _, err := svc.AdminList(nonAdminCtx(), domain.AdminListUsersQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAdminList_ClampsPagination(t *testing.T) {
	ctx, _ := adminUserCtx()
	repo := &mockUserRepo{}
	repo.On("AdminList", ctx, mock.MatchedBy(func(q domain.AdminListUsersQuery) bool {
		return q.ListParams.Page == 1 && q.ListParams.PageSize == domain.DefaultPageSize
	})).Return([]domain.User{}, int64(0), nil).Once()
	repo.On("AdminList", ctx, mock.MatchedBy(func(q domain.AdminListUsersQuery) bool {
		return q.ListParams.Page == 3 && q.ListParams.PageSize == 20
	})).Return([]domain.User{}, int64(0), nil).Once()

	svc := users.NewService(repo, &mockRoleRepo{}, nil, slog.Default())
	_, _, err := svc.AdminList(ctx, domain.AdminListUsersQuery{ListParams: domain.ListParams{Page: 0, PageSize: 0}})
	assert.NoError(t, err)
	_, _, err = svc.AdminList(ctx, domain.AdminListUsersQuery{ListParams: domain.ListParams{Page: 3, PageSize: 20}})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAdminCreate_HashesPasswordAndHonorsFlags(t *testing.T) {
	ctx, _ := adminUserCtx()
	repo := &mockUserRepo{}
	repo.On("Create", ctx, mock.MatchedBy(func(u *domain.User) bool {
		if !u.IsAdmin {
			return false
		}
		return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("Secret1A")) == nil
	})).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, slog.Default())
	user, err := svc.AdminCreate(ctx, domain.AdminCreateUserDTO{
		Username: "u", Name: "X", Password: "Secret1A",
		IsAdmin: true,
	})
	assert.NoError(t, err)
	assert.True(t, user.IsAdmin)
	repo.AssertExpectations(t)
}

func TestAdminCreate_Forbidden_WhenNotAdmin(t *testing.T) {
	svc := users.NewService(&mockUserRepo{}, &mockRoleRepo{}, nil, slog.Default())
	_, err := svc.AdminCreate(nonAdminCtx(), domain.AdminCreateUserDTO{
		Username: "u", Name: "X", Password: "Secret1A",
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAdminUpdate_MergesFieldsAndAllowsAdminFlag(t *testing.T) {
	ctx, _ := adminUserCtx()
	repo := &mockUserRepo{}
	userID := uuid.New()

	repo.On("FindByID", ctx, userID).Return(&domain.User{
		ID: userID, Name: "Old", IsAdmin: false,
	}, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.Name == "New" && u.IsAdmin
	})).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, slog.Default())
	newName := "New"
	isAdmin := true
	user, err := svc.AdminUpdate(ctx, userID, domain.AdminUpdateUserDTO{
		Name: &newName, IsAdmin: &isAdmin,
	})
	assert.NoError(t, err)
	assert.Equal(t, "New", user.Name)
	assert.True(t, user.IsAdmin)
	repo.AssertExpectations(t)
}

func TestAdminUpdate_NotFound(t *testing.T) {
	ctx, _ := adminUserCtx()
	repo := &mockUserRepo{}
	id := uuid.New()
	repo.On("FindByID", ctx, id).Return(nil, domain.ErrNotFound)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, slog.Default())
	_, err := svc.AdminUpdate(ctx, id, domain.AdminUpdateUserDTO{})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestAdminForceResetPassword_NoCurrentPasswordNeeded(t *testing.T) {
	ctx, _ := adminUserCtx()
	repo := &mockUserRepo{}
	userID := uuid.New()
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("irrelevant"), bcrypt.DefaultCost)

	repo.On("FindByID", ctx, userID).Return(&domain.User{ID: userID, Password: string(oldHash)}, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("NewPass1A")) == nil
	})).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, slog.Default())
	err := svc.AdminForceResetPassword(ctx, userID, domain.AdminForceResetPasswordDTO{NewPassword: "NewPass1A"})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAdminHardDelete_BlocksSelfDestruct(t *testing.T) {
	ctx, adminID := adminUserCtx()
	repo := &mockUserRepo{}

	svc := users.NewService(repo, &mockRoleRepo{}, nil, slog.Default())
	err := svc.AdminHardDelete(ctx, adminID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "HardDelete")
}

func TestAdminHardDelete_Success(t *testing.T) {
	ctx, _ := adminUserCtx()
	repo := &mockUserRepo{}
	target := uuid.New()

	repo.On("HardDelete", ctx, target).Return(nil)

	svc := users.NewService(repo, &mockRoleRepo{}, nil, slog.Default())
	err := svc.AdminHardDelete(ctx, target)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAdmin_NoCaller_Forbidden(t *testing.T) {
	repo := &mockUserRepo{}
	svc := users.NewService(repo, &mockRoleRepo{}, nil, slog.Default())

	_, _, err := svc.AdminList(context.Background(), domain.AdminListUsersQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	_, err2 := svc.AdminCreate(context.Background(), domain.AdminCreateUserDTO{})
	assert.ErrorIs(t, err2, domain.ErrForbidden)

	err3 := svc.AdminHardDelete(context.Background(), uuid.New())
	assert.ErrorIs(t, err3, domain.ErrForbidden)
}
