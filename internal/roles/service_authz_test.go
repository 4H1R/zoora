package roles_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/roles"
)

func nonAdminCaller(orgID uuid.UUID) domain.Caller {
	return domain.Caller{UserID: uuid.New(), OrgID: &orgID}
}

func TestGetByID_RejectsOtherOrgCustomRole(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	roleID := uuid.New()

	roleRepo := &mockRoleRepo{}
	roleRepo.On("FindByID", mock.Anything, roleID).Return(&domain.Role{ID: roleID, OrganizationID: &orgB}, nil)

	svc := roles.NewService(roleRepo, &mockPermRepo{}, noopTx{}, &auditSpy{}, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), nonAdminCaller(orgA))

	_, err := svc.GetByID(ctx, roleID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetByID_AllowsPresetRole(t *testing.T) {
	orgA := uuid.New()
	roleID := uuid.New()

	roleRepo := &mockRoleRepo{}
	roleRepo.On("FindByID", mock.Anything, roleID).Return(&domain.Role{ID: roleID, IsPreset: true}, nil)

	svc := roles.NewService(roleRepo, &mockPermRepo{}, noopTx{}, &auditSpy{}, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), nonAdminCaller(orgA))

	role, err := svc.GetByID(ctx, roleID)
	assert.NoError(t, err)
	assert.Equal(t, roleID, role.ID)
}

func TestGetByID_AllowsOwnOrgCustomRole(t *testing.T) {
	orgA := uuid.New()
	roleID := uuid.New()

	roleRepo := &mockRoleRepo{}
	roleRepo.On("FindByID", mock.Anything, roleID).Return(&domain.Role{ID: roleID, OrganizationID: &orgA}, nil)

	svc := roles.NewService(roleRepo, &mockPermRepo{}, noopTx{}, &auditSpy{}, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), nonAdminCaller(orgA))

	role, err := svc.GetByID(ctx, roleID)
	assert.NoError(t, err)
	assert.Equal(t, roleID, role.ID)
}

func TestUpdate_RejectsOtherOrgCustomRole(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	roleID := uuid.New()

	roleRepo := &mockRoleRepo{}
	roleRepo.On("FindByID", mock.Anything, roleID).Return(&domain.Role{ID: roleID, OrganizationID: &orgB}, nil)

	svc := roles.NewService(roleRepo, &mockPermRepo{}, noopTx{}, &auditSpy{}, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), nonAdminCaller(orgA))

	_, err := svc.Update(ctx, roleID, domain.UpdateRoleDTO{Name: "Hacked"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	roleRepo.AssertNotCalled(t, "Update")
}

func TestDelete_RejectsOtherOrgCustomRole(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	roleID := uuid.New()

	roleRepo := &mockRoleRepo{}
	roleRepo.On("FindByID", mock.Anything, roleID).Return(&domain.Role{ID: roleID, OrganizationID: &orgB}, nil)

	svc := roles.NewService(roleRepo, &mockPermRepo{}, noopTx{}, &auditSpy{}, nil, slog.Default())
	ctx := domain.WithCaller(context.Background(), nonAdminCaller(orgA))

	err := svc.Delete(ctx, roleID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	roleRepo.AssertNotCalled(t, "Delete")
}
