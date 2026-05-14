package factory_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
)

func TestNewClass(t *testing.T) {
	orgID := uuid.New()
	teacherID := uuid.New()
	c := factory.NewClass(orgID, teacherID)

	assert.Equal(t, orgID, c.OrganizationID)
	assert.Equal(t, teacherID, c.UserID)
	assert.NotEmpty(t, c.Name)
	assert.NotEmpty(t, c.Description)
}

func TestNewClass_WithOverride(t *testing.T) {
	orgID := uuid.New()
	teacherID := uuid.New()
	c := factory.NewClass(orgID, teacherID, func(c *domain.Class) {
		c.Name = "Math 101"
		c.TotalUsers = 30
	})

	assert.Equal(t, "Math 101", c.Name)
	assert.Equal(t, 30, c.TotalUsers)
}

func TestNewClassSession(t *testing.T) {
	classID := uuid.New()
	s := factory.NewClassSession(classID)

	assert.Equal(t, classID, s.ClassID)
	assert.NotEmpty(t, s.Name)
	assert.False(t, s.StartTime.IsZero())
}

func TestNewClassSession_WithOverride(t *testing.T) {
	classID := uuid.New()
	s := factory.NewClassSession(classID, func(s *domain.ClassSession) {
		s.Name = "Custom"
	})

	assert.Equal(t, "Custom", s.Name)
}

func TestNewClassMember(t *testing.T) {
	classID := uuid.New()
	userID := uuid.New()
	m := factory.NewClassMember(classID, userID)

	assert.Equal(t, classID, m.ClassID)
	assert.Equal(t, userID, m.UserID)
}
