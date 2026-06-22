package factory

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewUser(orgID uuid.UUID, opts ...func(*domain.User)) *domain.User {
	id := nextID()
	u := &domain.User{
		OrganizationID: &orgID,
		Username:       fmt.Sprintf("%s%d", fake.Username(), id),
		Name:           fakeName(),
		Password:       DefaultHashedPassword,
	}
	for _, o := range opts {
		o(u)
	}
	return u
}
