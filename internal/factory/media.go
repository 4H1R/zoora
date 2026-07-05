package factory

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewMedia(opts ...func(*domain.Media)) *domain.Media {
	id := nextID()
	orgID := uuid.New()
	m := &domain.Media{
		OrganizationID:   &orgID,
		ModelType:        fake.RandomString([]string{"user", "class", "organization"}),
		ModelID:          uuid.New(),
		CollectionName:   fake.RandomString([]string{"avatar", "documents", "attachments"}),
		Name:             fmt.Sprintf("file-%d", id),
		FileName:         fmt.Sprintf("file-%d.%s", id, fake.FileExtension()),
		MimeType:         fake.FileMimeType(),
		Disk:             "s3",
		Size:             int64(fake.IntRange(1024, 10485760)),
		CustomProperties: json.RawMessage(`{}`),
		OrderColumn:      0,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}
