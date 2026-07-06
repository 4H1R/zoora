//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/qa"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

type qaRepos struct {
	questions domain.QARepository
	votes     domain.QAVoteRepository
	users     domain.UserRepository
	orgs      domain.OrganizationRepository
	modelID   uuid.UUID
}

func setupQADB(t *testing.T) qaRepos {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.QAQuestion{},
		&domain.QAVote{},
	))
	return qaRepos{
		questions: qa.NewRepository(db),
		votes:     qa.NewVoteRepository(db),
		users:     users.NewRepository(db),
		orgs:      organizations.NewRepository(db),
		modelID:   uuid.New(),
	}
}

func seedQAQuestion(t *testing.T, repo domain.QARepository, userID, modelID uuid.UUID, text, status string) *domain.QAQuestion {
	t.Helper()
	q := &domain.QAQuestion{
		UserID:    userID,
		ModelType: domain.QAModelLiveSession,
		ModelID:   modelID,
		Text:      text,
		Status:    status,
	}
	require.NoError(t, repo.Create(context.Background(), q))
	return q
}

func TestIntegration_QARepo_VoteOrderingAndVotedByMe(t *testing.T) {
	r := setupQADB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	author := seedTeacher(t, r.users, org.ID, "qa-author")
	voterA := seedTeacher(t, r.users, org.ID, "qa-voter-a")
	voterB := seedTeacher(t, r.users, org.ID, "qa-voter-b")

	qLow := seedQAQuestion(t, r.questions, author.ID, r.modelID, "low votes", domain.QAStatusOpen)
	qHigh := seedQAQuestion(t, r.questions, author.ID, r.modelID, "high votes", domain.QAStatusOpen)
	qClosed := seedQAQuestion(t, r.questions, author.ID, r.modelID, "resolved", domain.QAStatusResolved)

	// qHigh gets 2 votes, qLow gets 1, qClosed gets 3 (but must still sink below open).
	require.NoError(t, r.votes.Create(ctx, &domain.QAVote{QuestionID: qHigh.ID, UserID: voterA.ID}))
	require.NoError(t, r.votes.Create(ctx, &domain.QAVote{QuestionID: qHigh.ID, UserID: voterB.ID}))
	require.NoError(t, r.votes.Create(ctx, &domain.QAVote{QuestionID: qLow.ID, UserID: voterA.ID}))
	require.NoError(t, r.votes.Create(ctx, &domain.QAVote{QuestionID: qClosed.ID, UserID: voterA.ID}))
	require.NoError(t, r.votes.Create(ctx, &domain.QAVote{QuestionID: qClosed.ID, UserID: voterB.ID}))
	require.NoError(t, r.votes.Create(ctx, &domain.QAVote{QuestionID: qClosed.ID, UserID: author.ID}))

	modelType := domain.QAModelLiveSession
	views, total, err := r.questions.List(ctx, domain.QAListScope{
		ViewerID:  voterA.ID,
		ModelType: &modelType,
		ModelID:   &r.modelID,
	}, domain.ListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, views, 3)

	// Open-first, then by vote count desc: qHigh(open,2), qLow(open,1), qClosed(resolved,3).
	assert.Equal(t, qHigh.ID, views[0].ID)
	assert.Equal(t, 2, views[0].VoteCount)
	assert.Equal(t, qLow.ID, views[1].ID)
	assert.Equal(t, 1, views[1].VoteCount)
	assert.Equal(t, qClosed.ID, views[2].ID)
	assert.Equal(t, domain.QAStatusResolved, views[2].Status)

	// voted_by_me is per-viewer: voterA voted on all three.
	for _, v := range views {
		assert.True(t, v.VotedByMe, "voterA voted on %s", v.Text)
	}
	assert.Equal(t, "qa-author", views[0].AuthorName)

	// A different viewer who voted on nothing sees voted_by_me=false.
	viewsB, _, err := r.questions.List(ctx, domain.QAListScope{
		ViewerID:  author.ID, // author only voted on qClosed
		ModelType: &modelType,
		ModelID:   &r.modelID,
	}, domain.ListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	byID := map[uuid.UUID]domain.QAQuestionView{}
	for _, v := range viewsB {
		byID[v.ID] = v
	}
	assert.False(t, byID[qHigh.ID].VotedByMe)
	assert.True(t, byID[qClosed.ID].VotedByMe)
}

func TestIntegration_QARepo_VoteUniquenessAndToggle(t *testing.T) {
	r := setupQADB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	author := seedTeacher(t, r.users, org.ID, "qa-author")
	voter := seedTeacher(t, r.users, org.ID, "qa-voter")

	q := seedQAQuestion(t, r.questions, author.ID, r.modelID, "unique?", domain.QAStatusOpen)

	require.NoError(t, r.votes.Create(ctx, &domain.QAVote{QuestionID: q.ID, UserID: voter.ID}))
	// Second vote by same user violates the unique index -> ErrConflict.
	err := r.votes.Create(ctx, &domain.QAVote{QuestionID: q.ID, UserID: voter.ID})
	assert.ErrorIs(t, err, domain.ErrConflict)

	count, err := r.votes.CountByQuestion(ctx, q.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Toggle off.
	removed, err := r.votes.Delete(ctx, q.ID, voter.ID)
	require.NoError(t, err)
	assert.True(t, removed)

	// Deleting again removes nothing.
	removed, err = r.votes.Delete(ctx, q.ID, voter.ID)
	require.NoError(t, err)
	assert.False(t, removed)
}

func TestIntegration_QARepo_CountOpenByUser(t *testing.T) {
	r := setupQADB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	author := seedTeacher(t, r.users, org.ID, "qa-author")

	seedQAQuestion(t, r.questions, author.ID, r.modelID, "open 1", domain.QAStatusOpen)
	seedQAQuestion(t, r.questions, author.ID, r.modelID, "open 2", domain.QAStatusOpen)
	seedQAQuestion(t, r.questions, author.ID, r.modelID, "resolved", domain.QAStatusResolved)

	count, err := r.questions.CountOpenByUser(ctx, domain.QAModelLiveSession, r.modelID, author.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
