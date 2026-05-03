package factory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
)

func TestNewOrganization(t *testing.T) {
	org := factory.NewOrganization()

	assert.NotEmpty(t, org.Name)
	assert.NotEmpty(t, org.Description)
	assert.Zero(t, org.TotalUsers)
}

func TestNewOrganization_WithOverrides(t *testing.T) {
	org := factory.NewOrganization(func(o *domain.Organization) {
		o.Name = "Custom Org"
	})

	assert.Equal(t, "Custom Org", org.Name)
}

func TestNewOrganization_Uniqueness(t *testing.T) {
	org1 := factory.NewOrganization()
	org2 := factory.NewOrganization()

	assert.NotEqual(t, org1.Name, org2.Name)
}
