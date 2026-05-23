package factory

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewQuiz(orgID, userID, classID uuid.UUID, opts ...func(*domain.Quiz)) *domain.Quiz {
	id := nextID()
	q := &domain.Quiz{
		OrganizationID:   orgID,
		UserID:           userID,
		ClassID:          classID,
		Title:            fmt.Sprintf("%s Quiz %d", fake.Noun(), id),
		Description:      fake.Sentence(8),
		DurationMinutes:  fake.IntRange(10, 60),
		NoBackNavigation: fake.Bool(),
		ShuffleQuestions: fake.Bool(),
	}
	for _, o := range opts {
		o(q)
	}
	return q
}

func NewQuizRule(quizID uuid.UUID, opts ...func(*domain.QuizRule)) *domain.QuizRule {
	r := &domain.QuizRule{
		QuizID: quizID,
		Type:   domain.QuizRuleTypeManual,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func NewQuizRoom(quizID, sessionID uuid.UUID, opts ...func(*domain.QuizRoom)) *domain.QuizRoom {
	// Window centered on now: opened a bit in past, closes a few hours later.
	openOffset := time.Duration(fake.IntRange(5, 30)) * time.Minute
	windowLen := time.Duration(fake.IntRange(60, 240)) * time.Minute
	start := time.Now().Add(-openOffset)
	end := start.Add(windowLen)
	qr := &domain.QuizRoom{
		QuizID:         quizID,
		ClassSessionID: sessionID,
		StartedAt:      &start,
		EndedAt:        &end,
	}
	for _, o := range opts {
		o(qr)
	}
	return qr
}

func NewQuizSubmission(quizID, userID uuid.UUID, opts ...func(*domain.QuizSubmission)) *domain.QuizSubmission {
	now := time.Now()
	submittedAt := now.Add(time.Duration(fake.IntRange(5, 30)) * time.Minute)
	s := &domain.QuizSubmission{
		QuizID:      quizID,
		UserID:      userID,
		Status:      domain.SubmissionStatusSubmitted,
		Answers:     []domain.SubmissionAnswer{},
		TotalScore:  0,
		StartedAt:   now,
		SubmittedAt: &submittedAt,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}
