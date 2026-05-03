package factory_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
)

func TestNewQuestionBank(t *testing.T) {
	orgID := uuid.New()
	qb := factory.NewQuestionBank(orgID)

	assert.Equal(t, orgID, qb.OrganizationID)
	assert.NotEmpty(t, qb.Name)
	assert.NotEmpty(t, qb.Description)
}

func TestNewQuestion_Choice(t *testing.T) {
	bankID := uuid.New()
	orgID := uuid.New()
	q := factory.NewQuestion(bankID, orgID, func(q *domain.Question) {
		q.Type = domain.QuestionTypeChoice
	})

	assert.Equal(t, bankID, q.BankID)
	assert.Equal(t, orgID, q.OrganizationID)
	assert.Equal(t, domain.QuestionTypeChoice, q.Type)
	assert.NotEmpty(t, q.Text)
	assert.NotEmpty(t, q.Options)
}

func TestNewQuestion_Descriptive(t *testing.T) {
	bankID := uuid.New()
	orgID := uuid.New()
	q := factory.NewQuestion(bankID, orgID, func(q *domain.Question) {
		q.Type = domain.QuestionTypeDescriptive
	})

	assert.Equal(t, domain.QuestionTypeDescriptive, q.Type)
	assert.Empty(t, q.Options)
}

func TestNewQuestion_DefaultType(t *testing.T) {
	bankID := uuid.New()
	orgID := uuid.New()
	q := factory.NewQuestion(bankID, orgID)

	assert.True(t, q.Type.Valid())
	assert.NotEmpty(t, q.Text)
}
