package factory

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewPermission(opts ...func(*domain.Permission)) *domain.Permission {
	id := nextID()
	p := &domain.Permission{
		Name: domain.PermissionName(fmt.Sprintf("resource%d:%s", id, fake.RandomString([]string{"create", "view", "update", "delete"}))),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func NewRole(orgID *uuid.UUID, opts ...func(*domain.Role)) *domain.Role {
	id := nextID()
	r := &domain.Role{
		OrganizationID: orgID,
		Name:           fmt.Sprintf("%s Role %d", fake.JobTitle(), id),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}
