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
	"github.com/4H1R/zoora/internal/questionbanks"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

type qbRepos struct {
	banks     domain.QuestionBankRepository
	questions domain.QuestionRepository
	users     domain.UserRepository
	orgs      domain.OrganizationRepository
}

func setupQuestionBanksDB(t *testing.T) qbRepos {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.QuestionBank{},
		&domain.Question{},
	))
	return qbRepos{
		banks:     questionbanks.NewRepository(db),
		questions: questionbanks.NewQuestionRepository(db),
		users:     users.NewRepository(db),
		orgs:      organizations.NewRepository(db),
	}
}

func seedBank(t *testing.T, repo domain.QuestionBankRepository, orgID uuid.UUID, name string) *domain.QuestionBank {
	t.Helper()
	b := &domain.QuestionBank{OrganizationID: orgID, Name: name, Description: "desc"}
	require.NoError(t, repo.Create(context.Background(), b))
	return b
}

func seedQuestion(t *testing.T, repo domain.QuestionRepository, bankID, orgID uuid.UUID, text string, qType domain.QuestionType) *domain.Question {
	t.Helper()
	q := &domain.Question{
		BankID:         bankID,
		OrganizationID: orgID,
		Text:           text,
		Type:           qType,
		Options: []domain.QuestionOption{
			{ID: "a", Value: "opt1", Score: 1},
			{ID: "b", Value: "opt2", Score: 0},
		},
	}
	require.NoError(t, repo.Create(context.Background(), q))
	return q
}

func TestIntegration_QuestionBankRepo_CRUD(t *testing.T) {
	r := setupQuestionBanksDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")

	bank := seedBank(t, r.banks, org.ID, "Physics")

	got, err := r.banks.FindByID(ctx, bank.ID)
	require.NoError(t, err)
	assert.Equal(t, "Physics", got.Name)

	got.Name = "Chemistry"
	require.NoError(t, r.banks.Update(ctx, got))

	updated, err := r.banks.FindByID(ctx, bank.ID)
	require.NoError(t, err)
	assert.Equal(t, "Chemistry", updated.Name)

	require.NoError(t, r.banks.Delete(ctx, bank.ID))
	_, err = r.banks.FindByID(ctx, bank.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	soft, err := r.banks.FindByIDIncludingDeleted(ctx, bank.ID)
	require.NoError(t, err)
	assert.True(t, soft.DeletedAt.Valid)

	require.NoError(t, r.banks.HardDelete(ctx, bank.ID))
	_, err = r.banks.FindByIDIncludingDeleted(ctx, bank.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_QuestionBankRepo_List_OrgScoped(t *testing.T) {
	r := setupQuestionBanksDB(t)
	ctx := context.Background()
	orgA := seedOrg(t, r.orgs, "A")
	orgB := seedOrg(t, r.orgs, "B")

	seedBank(t, r.banks, orgA.ID, "BankA1")
	seedBank(t, r.banks, orgA.ID, "BankA2")
	seedBank(t, r.banks, orgB.ID, "BankB1")

	bigPage := domain.ListParams{Page: 1, PageSize: 50}
	banks, total, err := r.banks.List(ctx, orgA.ID, bigPage)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, banks, 2)
}

func TestIntegration_QuestionBankRepo_AdminList_SearchAndFilter(t *testing.T) {
	r := setupQuestionBanksDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")

	seedBank(t, r.banks, org.ID, "Algebra")
	seedBank(t, r.banks, org.ID, "Algorithms")
	seedBank(t, r.banks, org.ID, "Biology")

	bigPage := domain.ListParams{
		Page:         1,
		PageSize:     50,
		Search:       "algo",
		SearchFields: []string{"name", "description"},
	}
	banks, total, err := r.banks.AdminList(ctx, domain.AdminListQuestionBanksQuery{ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Algorithms", banks[0].Name)

	q := domain.AdminListQuestionBanksQuery{
		OrganizationID: &org.ID,
		ListParams:     domain.ListParams{Page: 1, PageSize: 50},
	}
	_, total, err = r.banks.AdminList(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
}

func TestIntegration_QuestionRepo_CRUD(t *testing.T) {
	r := setupQuestionBanksDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	bank := seedBank(t, r.banks, org.ID, "Math")

	q := seedQuestion(t, r.questions, bank.ID, org.ID, "What is 2+2?", domain.QuestionTypeChoice)

	got, err := r.questions.FindByID(ctx, q.ID)
	require.NoError(t, err)
	assert.Equal(t, "What is 2+2?", got.Text)
	assert.Equal(t, domain.QuestionTypeChoice, got.Type)
	assert.Len(t, got.Options, 2)

	got.Text = "What is 3+3?"
	require.NoError(t, r.questions.Update(ctx, got))

	require.NoError(t, r.questions.Delete(ctx, q.ID))
	_, err = r.questions.FindByID(ctx, q.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	require.NoError(t, r.questions.HardDelete(ctx, q.ID))
}

func TestIntegration_QuestionRepo_ListByBank_FilterType(t *testing.T) {
	r := setupQuestionBanksDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	bank := seedBank(t, r.banks, org.ID, "Mixed")

	seedQuestion(t, r.questions, bank.ID, org.ID, "Describe X", domain.QuestionTypeDescriptive)
	seedQuestion(t, r.questions, bank.ID, org.ID, "Choose one", domain.QuestionTypeChoice)
	seedQuestion(t, r.questions, bank.ID, org.ID, "Short answer", domain.QuestionTypeShortAnswer)

	bigPage := domain.ListParams{Page: 1, PageSize: 50}

	_, total, err := r.questions.ListByBank(ctx, bank.ID, domain.ListQuestionsQuery{ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)

	choiceType := domain.QuestionTypeChoice
	got, total, err := r.questions.ListByBank(ctx, bank.ID, domain.ListQuestionsQuery{Type: &choiceType, ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, domain.QuestionTypeChoice, got[0].Type)
}

func TestIntegration_QuestionRepo_FindByIDs(t *testing.T) {
	r := setupQuestionBanksDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	bank := seedBank(t, r.banks, org.ID, "Bank")

	q1 := seedQuestion(t, r.questions, bank.ID, org.ID, "Q1", domain.QuestionTypeChoice)
	q2 := seedQuestion(t, r.questions, bank.ID, org.ID, "Q2", domain.QuestionTypeShortAnswer)
	seedQuestion(t, r.questions, bank.ID, org.ID, "Q3", domain.QuestionTypeDescriptive)

	got, err := r.questions.FindByIDs(ctx, []uuid.UUID{q1.ID, q2.ID})
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestIntegration_QuestionRepo_CountByBank(t *testing.T) {
	r := setupQuestionBanksDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	bank := seedBank(t, r.banks, org.ID, "Bank")

	seedQuestion(t, r.questions, bank.ID, org.ID, "Q1", domain.QuestionTypeChoice)
	seedQuestion(t, r.questions, bank.ID, org.ID, "Q2", domain.QuestionTypeChoice)

	count, err := r.questions.CountByBank(ctx, bank.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestIntegration_QuestionRepo_RandomByBank(t *testing.T) {
	r := setupQuestionBanksDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	bank := seedBank(t, r.banks, org.ID, "Bank")

	for i := 0; i < 5; i++ {
		seedQuestion(t, r.questions, bank.ID, org.ID, "Q", domain.QuestionTypeChoice)
	}

	got, err := r.questions.RandomByBank(ctx, bank.ID, 3)
	require.NoError(t, err)
	assert.Len(t, got, 3)
}
