//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
)

func seedOrgMedia(t *testing.T, repo domain.MediaRepository, orgID uuid.UUID, modelType, name string, size int64) *domain.Media {
	t.Helper()
	m := &domain.Media{
		OrganizationID:   &orgID,
		ModelType:        modelType,
		ModelID:          uuid.New(),
		CollectionName:   "attachments",
		Name:             name,
		FileName:         name,
		MimeType:         "application/pdf",
		Disk:             "s3",
		Size:             size,
		CustomProperties: json.RawMessage(`{}`),
	}
	require.NoError(t, repo.Create(context.Background(), m))
	return m
}

func TestIntegration_MediaRepo_ListFoldersGroupsByModelType(t *testing.T) {
	r := setupEngagementDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	// A bare UUID stands in for another tenant: media.organization_id has no FK,
	// and seeding a second org trips the pre-existing empty-slug unique collision.
	otherOrgID := uuid.New()

	seedOrgMedia(t, r.media, org.ID, "live_room", "a.pdf", 10)
	seedOrgMedia(t, r.media, org.ID, "live_room", "b.pdf", 20)
	seedOrgMedia(t, r.media, org.ID, "practice", "c.pdf", 5)
	seedOrgMedia(t, r.media, otherOrgID, "live_room", "leak.pdf", 99)

	folders, err := r.media.ListFolders(ctx, org.ID)
	require.NoError(t, err)
	require.Len(t, folders, 2)
	byType := map[string]domain.MediaFolder{}
	for _, f := range folders {
		byType[f.ModelType] = f
	}
	assert.Equal(t, int64(2), byType["live_room"].FileCount)
	assert.Equal(t, int64(30), byType["live_room"].TotalSize)
	assert.Equal(t, int64(1), byType["practice"].FileCount)
	assert.Equal(t, int64(5), byType["practice"].TotalSize)
}

func TestIntegration_MediaRepo_ListFilesPaginatesAndSearches(t *testing.T) {
	r := setupEngagementDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	// A bare UUID stands in for another tenant: media.organization_id has no FK,
	// and seeding a second org trips the pre-existing empty-slug unique collision.
	otherOrgID := uuid.New()

	seedOrgMedia(t, r.media, org.ID, "live_room", "algebra-notes.pdf", 10)
	seedOrgMedia(t, r.media, org.ID, "live_room", "geometry-notes.pdf", 10)
	seedOrgMedia(t, r.media, org.ID, "practice", "algebra-homework.pdf", 10)
	seedOrgMedia(t, r.media, otherOrgID, "live_room", "algebra-leak.pdf", 10)

	// Only the requested org + model_type.
	items, total, err := r.media.ListFiles(ctx, org.ID, "live_room", domain.ListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	// Search narrows within the folder.
	items, total, err = r.media.ListFiles(ctx, org.ID, "live_room", domain.ListParams{
		Page: 1, PageSize: 50,
		Search: "algebra", SearchFields: []string{"name", "file_name"},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "algebra-notes.pdf", items[0].FileName)

	// Pagination caps the page.
	items, total, err = r.media.ListFiles(ctx, org.ID, "live_room", domain.ListParams{Page: 1, PageSize: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 1)
}
