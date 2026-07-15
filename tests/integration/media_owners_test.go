//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/media"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

type ownerRepos struct {
	db    *gorm.DB
	media domain.MediaRepository
	orgs  domain.OrganizationRepository
	users domain.UserRepository
}

func setupOwnersDB(t *testing.T) ownerRepos {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Media{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.ClassMember{},
		&domain.LiveRoom{},
		&domain.LiveRecording{},
		&domain.QuestionBank{},
		&domain.Question{},
	))
	return ownerRepos{
		db:    db,
		media: media.NewRepository(db),
		orgs:  organizations.NewRepository(db),
		users: users.NewRepository(db),
	}
}

// seedMediaFor creates a media row whose ModelID points at a real owner, so the
// resolver can walk it up to a class / bank / etc.
func seedMediaFor(t *testing.T, repo domain.MediaRepository, orgID uuid.UUID, modelType string, modelID uuid.UUID, name string, size int64) *domain.Media {
	t.Helper()
	m := &domain.Media{
		OrganizationID:   &orgID,
		ModelType:        modelType,
		ModelID:          modelID,
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

// TestIntegration_MediaRepo_OwnerResolution exercises the raw-SQL owner resolver
// end-to-end against Postgres: every media row (+ recordings) must roll up to
// the right bucket, and files must union + page in SQL.
func TestIntegration_MediaRepo_OwnerResolution(t *testing.T) {
	r := setupOwnersDB(t)
	ctx := context.Background()

	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "owner-teacher")

	// class -> session -> live_room, with a live_room file AND a recording.
	class := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Math 101"}
	require.NoError(t, r.db.Create(class).Error)
	session := &domain.ClassSession{ClassID: class.ID, Name: "Session 1", StartTime: time.Now()}
	require.NoError(t, r.db.Create(session).Error)
	room := &domain.LiveRoom{
		ClassSessionID:  session.ID,
		Name:            "Room A",
		LiveKitRoomName: uuid.NewString(),
		Config:          domain.LiveRoomConfig{},
	}
	require.NoError(t, r.db.Create(room).Error)
	seedMediaFor(t, r.media, org.ID, domain.MediaModelLiveRoom, room.ID, "slides.pdf", 100)
	rec := &domain.LiveRecording{
		LiveRoomID: room.ID,
		EgressID:   "egress-1",
		Status:     domain.LiveRecordingStatusCompleted,
		Size:       9000,
		StartedAt:  time.Now(),
	}
	require.NoError(t, r.db.Create(rec).Error)

	// question -> bank.
	bank := &domain.QuestionBank{OrganizationID: org.ID, Name: "Algebra Bank"}
	require.NoError(t, r.db.Create(bank).Error)
	question := &domain.Question{OrganizationID: org.ID, BankID: bank.ID, Text: "2+2?", Type: domain.QuestionTypeChoice}
	require.NoError(t, r.db.Create(question).Error)
	seedMediaFor(t, r.media, org.ID, domain.QuestionMediaModelType, question.ID, "q.png", 500)

	// shared + orphan (unknown model_type -> "other").
	seedMediaFor(t, r.media, org.ID, domain.MediaModelOrganization, org.ID, "shared.pdf", 10)
	seedMediaFor(t, r.media, org.ID, "some_dead_feature", uuid.New(), "orphan.pdf", 7)

	// Cross-tenant leak guard.
	seedMediaFor(t, r.media, uuid.New(), domain.MediaModelOrganization, uuid.New(), "leak.pdf", 999)

	t.Run("ListOwnerMedia buckets every row (recordings excluded)", func(t *testing.T) {
		owners, err := r.media.ListOwnerMedia(ctx, org.ID)
		require.NoError(t, err)

		type key struct {
			kind string
			id   string
		}
		got := map[key]domain.MediaOwner{}
		for _, o := range owners {
			id := ""
			if o.OwnerID != nil {
				id = o.OwnerID.String()
			}
			got[key{o.OwnerKind, id}] = o
		}

		cls := got[key{domain.MediaOwnerClass, class.ID.String()}]
		assert.Equal(t, int64(1), cls.FileCount)
		assert.Equal(t, int64(100), cls.TotalSize) // recording NOT counted here
		assert.Equal(t, "Math 101", cls.Name)

		qb := got[key{domain.MediaOwnerQuestionBank, bank.ID.String()}]
		assert.Equal(t, int64(500), qb.TotalSize)
		assert.Equal(t, "Algebra Bank", qb.Name)

		assert.Equal(t, int64(10), got[key{domain.MediaOwnerShared, ""}].TotalSize)
		assert.Equal(t, int64(7), got[key{domain.MediaOwnerOther, ""}].TotalSize)

		// No cross-tenant row leaked in.
		for _, o := range owners {
			assert.NotEqual(t, int64(999), o.TotalSize)
		}
	})

	t.Run("ListOwnerRecordings groups by class", func(t *testing.T) {
		owners, err := r.media.ListOwnerRecordings(ctx, org.ID)
		require.NoError(t, err)
		require.Len(t, owners, 1)
		assert.Equal(t, domain.MediaOwnerClass, owners[0].OwnerKind)
		require.NotNil(t, owners[0].OwnerID)
		assert.Equal(t, class.ID, *owners[0].OwnerID)
		assert.Equal(t, int64(9000), owners[0].TotalSize)
	})

	t.Run("ListOwnerFiles unions recordings for class, size-sorted", func(t *testing.T) {
		files, total, err := r.media.ListOwnerFiles(ctx, org.ID, domain.MediaOwnerClass, &class.ID, domain.ListParams{
			Page: 1, PageSize: 20, OrderBy: "size", OrderDir: "desc",
		})
		require.NoError(t, err)
		require.Equal(t, int64(2), total)
		require.Len(t, files, 2)
		// Recording (9000) sorts above the slide (100) and is read-only.
		assert.Equal(t, "recording", files[0].Source)
		assert.False(t, files[0].Deletable)
		assert.Equal(t, "media", files[1].Source)
		assert.True(t, files[1].Deletable)
	})

	t.Run("ListOwnerFiles paginates in SQL", func(t *testing.T) {
		files, total, err := r.media.ListOwnerFiles(ctx, org.ID, domain.MediaOwnerClass, &class.ID, domain.ListParams{
			Page: 1, PageSize: 1, OrderBy: "size", OrderDir: "desc",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, files, 1)
		assert.Equal(t, "recording", files[0].Source)
	})

	t.Run("ListOwnerFiles search narrows by name", func(t *testing.T) {
		files, total, err := r.media.ListOwnerFiles(ctx, org.ID, domain.MediaOwnerClass, &class.ID, domain.ListParams{
			Page: 1, PageSize: 20, Search: "Recording",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, files, 1)
		assert.Equal(t, "recording", files[0].Source)
	})

	t.Run("ListOwnerFiles for non-class owner excludes recordings", func(t *testing.T) {
		files, total, err := r.media.ListOwnerFiles(ctx, org.ID, domain.MediaOwnerQuestionBank, &bank.ID, domain.ListParams{
			Page: 1, PageSize: 20,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, files, 1)
		assert.Equal(t, "q.png", files[0].Name)
	})

	t.Run("ListOwnerFiles for shared and other buckets (nil owner id)", func(t *testing.T) {
		shared, total, err := r.media.ListOwnerFiles(ctx, org.ID, domain.MediaOwnerShared, nil, domain.ListParams{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, shared, 1)
		assert.Equal(t, "shared.pdf", shared[0].Name)

		other, total, err := r.media.ListOwnerFiles(ctx, org.ID, domain.MediaOwnerOther, nil, domain.ListParams{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, other, 1)
		assert.Equal(t, "orphan.pdf", other[0].Name)
	})
}
