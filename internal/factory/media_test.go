package factory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
)

func TestNewMedia(t *testing.T) {
	m := factory.NewMedia()

	assert.NotEmpty(t, m.ModelType)
	assert.NotEmpty(t, m.FileName)
	assert.NotEmpty(t, m.MimeType)
	assert.Equal(t, "s3", m.Disk)
	assert.Greater(t, m.Size, int64(0))
}

func TestNewMedia_WithOverride(t *testing.T) {
	m := factory.NewMedia(func(m *domain.Media) {
		m.ModelType = "user"
		m.CollectionName = "avatar"
	})

	assert.Equal(t, "user", m.ModelType)
	assert.Equal(t, "avatar", m.CollectionName)
}
