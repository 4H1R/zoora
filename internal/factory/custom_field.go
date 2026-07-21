package factory

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewCustomFieldDefinition(orgID uuid.UUID, opts ...func(*domain.UserCustomFieldDefinition)) *domain.UserCustomFieldDefinition {
	id := nextID()
	d := &domain.UserCustomFieldDefinition{
		OrganizationID: orgID,
		Label:          fmt.Sprintf("Field %d", id),
		FieldType:      domain.CustomFieldTypeText,
		Options:        []string{},
	}
	for _, o := range opts {
		o(d)
	}
	return d
}
