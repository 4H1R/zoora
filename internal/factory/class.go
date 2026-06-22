package factory

import (
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewClass(orgID, teacherID uuid.UUID, opts ...func(*domain.Class)) *domain.Class {
	id := nextID()
	c := &domain.Class{
		OrganizationID: orgID,
		UserID:         teacherID,
		Name:           fakeClassName(id),
		Description:    fakeSentence(8),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func NewClassSession(classID uuid.UUID, opts ...func(*domain.ClassSession)) *domain.ClassSession {
	id := nextID()
	s := &domain.ClassSession{
		ClassID:     classID,
		Name:        fakeSessionName(id),
		Description: fakeSentence(6),
		StartTime:   time.Now().Add(time.Duration(id) * 24 * time.Hour),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func NewClassMember(classID, userID uuid.UUID) *domain.ClassMember {
	return &domain.ClassMember{
		ClassID: classID,
		UserID:  userID,
	}
}
