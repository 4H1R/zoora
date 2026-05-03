package factory_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
)

func TestNewUser(t *testing.T) {
	orgID := uuid.New()
	u := factory.NewUser(orgID)

	assert.NotEmpty(t, u.Username)
	assert.NotEmpty(t, u.Name)
	assert.Equal(t, &orgID, u.OrganizationID)
	assert.False(t, u.IsAdmin)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("password")))
}

func TestNewUser_WithOverrides(t *testing.T) {
	orgID := uuid.New()
	u := factory.NewUser(orgID, func(u *domain.User) {
		u.Username = "custom"
		u.IsAdmin = true
	})

	assert.Equal(t, "custom", u.Username)
	assert.True(t, u.IsAdmin)
}

func TestNewUser_UniqueFields(t *testing.T) {
	orgID := uuid.New()
	u1 := factory.NewUser(orgID)
	u2 := factory.NewUser(orgID)

	assert.NotEqual(t, u1.Username, u2.Username)
}
