package factory

import (
	"fmt"

	"github.com/4H1R/zoora/internal/domain"
)

func NewOrganization(opts ...func(*domain.Organization)) *domain.Organization {
	id := nextID()
	o := &domain.Organization{
		Name:        fakeOrgName(id),
		Slug:        fmt.Sprintf("org-%d", id), // unique per nextID(); valid DNS label
		Description: fakeSentence(8),
		Status:      domain.OrganizationStatusActive,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
