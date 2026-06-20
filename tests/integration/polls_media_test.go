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
	"github.com/4H1R/zoora/internal/media"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/polls"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

type engagementRepos struct {
	polls       domain.PollRepository
	answers     domain.PollAnswerRepository
	media       domain.MediaRepository
	users       domain.UserRepository
	orgs        domain.OrganizationRepository
	mediaModel  uuid.UUID
	pollModelID uuid.UUID
}

func setupEngagementDB(t *testing.T) engagementRepos {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Poll{},
		&domain.PollAnswer{},
		&domain.Media{},
	))
	return engagementRepos{
		polls:       polls.NewRepository(db),
		answers:     polls.NewAnswerRepository(db),
		media:       media.NewRepository(db),
		users:       users.NewRepository(db),
		orgs:        organizations.NewRepository(db),
		mediaModel:  uuid.New(),
		pollModelID: uuid.New(),
	}
}

func seedPoll(t *testing.T, repo domain.PollRepository, userID, modelID uuid.UUID, name string) *domain.Poll {
	t.Helper()
	poll := &domain.Poll{
		UserID:              userID,
		ModelType:           "live_room",
		ModelID:             modelID,
		Name:                name,
		AllowedAnswersCount: 2,
		Options: []domain.PollOption{
			{Label: "Option A", Value: "a"},
			{Label: "Option B", Value: "b"},
			{Label: "Option C", Value: "c"},
		},
	}
	require.NoError(t, repo.Create(context.Background(), poll))
	return poll
}

func TestIntegration_PollRepo_CRUDScopeAndSoftDelete(t *testing.T) {
	r := setupEngagementDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	owner := seedTeacher(t, r.users, org.ID, "poll-owner")
	otherOwner := seedTeacher(t, r.users, org.ID, "poll-other")

	poll := seedPoll(t, r.polls, owner.ID, r.pollModelID, "Check-in")
	seedPoll(t, r.polls, otherOwner.ID, uuid.New(), "Other")

	got, err := r.polls.FindByID(ctx, poll.ID)
	require.NoError(t, err)
	assert.Equal(t, "Check-in", got.Name)
	assert.Len(t, got.Options, 3)

	got.Name = "Updated check-in"
	require.NoError(t, r.polls.Update(ctx, got))

	modelType := "live_room"
	scoped, total, err := r.polls.List(ctx, domain.PollListScope{
		OwnerID:   &owner.ID,
		ModelType: &modelType,
		ModelID:   &r.pollModelID,
	}, domain.ListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Updated check-in", scoped[0].Name)

	all, total, err := r.polls.List(ctx, domain.PollListScope{AllOrgs: true}, domain.ListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, all, 2)

	require.NoError(t, r.polls.Delete(ctx, poll.ID))
	_, err = r.polls.FindByID(ctx, poll.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	deleted, err := r.polls.FindByIDIncludingDeleted(ctx, poll.ID)
	require.NoError(t, err)
	assert.True(t, deleted.DeletedAt.Valid)

	_, total, err = r.polls.List(ctx, domain.PollListScope{
		OwnerID:        &owner.ID,
		IncludeDeleted: true,
	}, domain.ListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)

	require.NoError(t, r.polls.HardDelete(ctx, poll.ID))
	_, err = r.polls.FindByIDIncludingDeleted(ctx, poll.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_PollAnswerRepo_ListFilterAndDeleteByUser(t *testing.T) {
	r := setupEngagementDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	owner := seedTeacher(t, r.users, org.ID, "answer-owner")
	student := seedTeacher(t, r.users, org.ID, "answer-student")
	otherStudent := seedTeacher(t, r.users, org.ID, "answer-other")
	poll := seedPoll(t, r.polls, owner.ID, r.pollModelID, "Vote")

	for _, answer := range []domain.PollAnswer{
		{PollID: poll.ID, UserID: student.ID, Option: "a"},
		{PollID: poll.ID, UserID: student.ID, Option: "b"},
		{PollID: poll.ID, UserID: otherStudent.ID, Option: "c"},
	} {
		require.NoError(t, r.answers.Create(ctx, &answer))
	}

	ownAnswers, err := r.answers.FindByPollAndUser(ctx, poll.ID, student.ID)
	require.NoError(t, err)
	assert.Len(t, ownAnswers, 2)

	filtered, total, err := r.answers.ListByPoll(ctx, poll.ID, domain.ListPollAnswersQuery{
		UserID:     &student.ID,
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, filtered, 2)

	require.NoError(t, r.answers.DeleteByPollAndUser(ctx, poll.ID, student.ID))
	ownAnswers, err = r.answers.FindByPollAndUser(ctx, poll.ID, student.ID)
	require.NoError(t, err)
	assert.Empty(t, ownAnswers)

	remaining, total, err := r.answers.ListByPoll(ctx, poll.ID, domain.ListPollAnswersQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, otherStudent.ID, remaining[0].UserID)
}

func TestIntegration_MediaRepo_ListByModelOrdersAndFilters(t *testing.T) {
	r := setupEngagementDB(t)
	ctx := context.Background()
	modelID := r.mediaModel
	props := json.RawMessage(`{"width":1200,"height":800}`)

	items := []*domain.Media{
		{
			ModelType:        "practice_room",
			ModelID:          modelID,
			CollectionName:   "attachments",
			Name:             "Second",
			FileName:         "second.pdf",
			MimeType:         "application/pdf",
			Disk:             "s3",
			Size:             20,
			CustomProperties: props,
			OrderColumn:      2,
		},
		{
			ModelType:        "practice_room",
			ModelID:          modelID,
			CollectionName:   "attachments",
			Name:             "First",
			FileName:         "first.png",
			MimeType:         "image/png",
			Disk:             "s3",
			Size:             10,
			CustomProperties: props,
			OrderColumn:      1,
		},
		{
			ModelType:        "practice_room",
			ModelID:          modelID,
			CollectionName:   "solutions",
			Name:             "Solution",
			FileName:         "solution.txt",
			MimeType:         "text/plain",
			Disk:             "s3",
			Size:             5,
			CustomProperties: props,
			OrderColumn:      0,
		},
	}
	for _, item := range items {
		require.NoError(t, r.media.Create(ctx, item))
	}

	all, err := r.media.ListByModel(ctx, "practice_room", modelID, "")
	require.NoError(t, err)
	require.Len(t, all, 3)
	assert.Equal(t, "solution.txt", all[0].FileName)
	assert.Equal(t, "first.png", all[1].FileName)
	assert.Equal(t, "second.pdf", all[2].FileName)

	attachments, err := r.media.ListByModel(ctx, "practice_room", modelID, "attachments")
	require.NoError(t, err)
	require.Len(t, attachments, 2)
	assert.Equal(t, "first.png", attachments[0].FileName)
	assert.Equal(t, "second.pdf", attachments[1].FileName)

	got, err := r.media.FindByID(ctx, items[1].ID)
	require.NoError(t, err)
	assert.JSONEq(t, string(props), string(got.CustomProperties))
	assert.Equal(t, "practice_room/"+modelID.String()+"/attachments/first.png", got.S3Key())

	require.NoError(t, r.media.Delete(ctx, items[1].ID))
	_, err = r.media.FindByID(ctx, items[1].ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
