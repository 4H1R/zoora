package factory

import (
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewOrganizationSettings(orgID uuid.UUID, opts ...func(*domain.OrganizationSettings)) *domain.OrganizationSettings {
	s := domain.NewDefaultOrganizationSettings(orgID)
	for _, opt := range opts {
		opt(s)
	}
	return s
}
