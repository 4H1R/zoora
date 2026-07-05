package factory_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
)

func TestNewQuiz(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	classID := uuid.New()
	q := factory.NewQuiz(orgID, userID, classID)

	assert.Equal(t, orgID, q.OrganizationID)
	assert.Equal(t, userID, q.UserID)
	assert.Equal(t, classID, q.ClassID)
	assert.NotEmpty(t, q.Title)
	assert.Greater(t, q.DurationMinutes, 0)
}

func TestNewQuizRule_Manual(t *testing.T) {
	quizID := uuid.New()
	r := factory.NewQuizRule(quizID, func(r *domain.QuizRule) {
		r.Type = domain.QuizRuleTypeManual
		r.QuestionIDs = []uuid.UUID{uuid.New(), uuid.New()}
	})

	assert.Equal(t, quizID, r.QuizID)
	assert.Equal(t, domain.QuizRuleTypeManual, r.Type)
	assert.Len(t, r.QuestionIDs, 2)
}

func TestNewQuizRule_Random(t *testing.T) {
	quizID := uuid.New()
	bankID := uuid.New()
	r := factory.NewQuizRule(quizID, func(r *domain.QuizRule) {
		r.Type = domain.QuizRuleTypeRandom
		r.BankID = &bankID
		r.Count = 5
	})

	assert.Equal(t, domain.QuizRuleTypeRandom, r.Type)
	assert.Equal(t, &bankID, r.BankID)
	assert.Equal(t, 5, r.Count)
}

func TestNewQuizRoom(t *testing.T) {
	quizID := uuid.New()
	sessionID := uuid.New()
	qr := factory.NewQuizRoom(quizID, sessionID)

	assert.Equal(t, quizID, qr.QuizID)
	assert.Equal(t, sessionID, qr.ClassSessionID)
}

func TestNewQuizSubmission(t *testing.T) {
	quizID := uuid.New()
	userID := uuid.New()
	s := factory.NewQuizSubmission(quizID, userID)

	assert.Equal(t, quizID, s.QuizID)
	assert.Equal(t, userID, s.UserID)
	assert.Equal(t, domain.SubmissionStatusSubmitted, s.Status)
	assert.False(t, s.StartedAt.IsZero())
	assert.NotNil(t, s.SubmittedAt)
	// question_set is NOT NULL in the DB; a nil slice serializes to SQL NULL
	// (bypassing the column default), so the factory must always emit at least [].
	assert.NotNil(t, s.QuestionSet)
}
