//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/changelog"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/tests/testutil"
)

func setupChangelogDB(t *testing.T) (domain.ChangelogRepository, *domain.User) {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(&domain.ChangelogEntry{}, &domain.User{}))

	// A user whose last-seen marker is old, so freshly published entries count
	// as unseen.
	old := time.Now().Add(-24 * time.Hour)
	user := &domain.User{
		Username:            "changelog-reader",
		Name:                "Reader",
		Password:            "x",
		ChangelogLastSeenAt: &old,
	}
	require.NoError(t, db.Create(user).Error)
	return changelog.NewRepository(db), user
}

func TestIntegration_Changelog_PublishFeedAndUnseen(t *testing.T) {
	repo, user := setupChangelogDB(t)
	ctx := context.Background()

	// Draft: not visible in the published feed.
	version := "v2.0.0"
	entry := &domain.ChangelogEntry{
		Version: &version,
		TitleEn: "Big Release",
		BodyEn:  "## New\n\n- Stuff",
		IsMajor: true,
	}
	require.NoError(t, repo.Create(ctx, entry))

	items, total, err := repo.ListPublished(ctx, 20, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total, "draft must not appear in feed")
	assert.Empty(t, items)

	// Publish → visible, and it becomes the current version.
	now := time.Now()
	entry.PublishedAt = &now
	require.NoError(t, repo.Update(ctx, entry))

	items, total, err = repo.ListPublished(ctx, 20, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "Big Release", items[0].TitleEn)

	latest, err := repo.LatestPublished(ctx)
	require.NoError(t, err)
	require.NotNil(t, latest)
	require.NotNil(t, latest.Version)
	assert.Equal(t, "v2.0.0", *latest.Version)

	// Unseen relative to the user's old marker: 1, and the major is detected.
	seen, err := repo.GetLastSeen(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, seen)

	unseen, err := repo.CountUnseen(ctx, seen)
	require.NoError(t, err)
	assert.Equal(t, int64(1), unseen)

	major, err := repo.LatestMajorUnseen(ctx, seen)
	require.NoError(t, err)
	require.NotNil(t, major)
	assert.Equal(t, "Big Release", major.TitleEn)

	// After marking seen (now), nothing is unseen.
	require.NoError(t, repo.UpdateLastSeen(ctx, user.ID, time.Now().Add(time.Minute)))
	newSeen, err := repo.GetLastSeen(ctx, user.ID)
	require.NoError(t, err)

	unseen, err = repo.CountUnseen(ctx, newSeen)
	require.NoError(t, err)
	assert.Equal(t, int64(0), unseen, "marking seen clears unseen count")

	major, err = repo.LatestMajorUnseen(ctx, newSeen)
	require.NoError(t, err)
	assert.Nil(t, major)
}
