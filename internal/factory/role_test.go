package factory_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
)

func TestNewPermission(t *testing.T) {
	p := factory.NewPermission()

	assert.NotEmpty(t, p.Name)
}

func TestNewPermission_WithOverride(t *testing.T) {
	p := factory.NewPermission(func(p *domain.Permission) {
		p.Name = "meetings:create"
	})

	assert.Equal(t, domain.PermissionName("meetings:create"), p.Name)
}

func TestNewRole(t *testing.T) {
	orgID := uuid.New()
	r := factory.NewRole(&orgID)

	assert.Equal(t, &orgID, r.OrganizationID)
	assert.NotEmpty(t, r.Name)
}

func TestNewRole_WithOverride(t *testing.T) {
	orgID := uuid.New()
	r := factory.NewRole(&orgID, func(r *domain.Role) {
		r.Name = "Teacher"
	})

	assert.Equal(t, "Teacher", r.Name)
}
