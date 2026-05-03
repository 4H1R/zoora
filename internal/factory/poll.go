package factory

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewPoll(userID uuid.UUID, modelType string, modelID uuid.UUID, opts ...func(*domain.Poll)) *domain.Poll {
	id := nextID()
	p := &domain.Poll{
		UserID:              userID,
		ModelType:           modelType,
		ModelID:             modelID,
		Name:                fmt.Sprintf("%s Poll %d", fake.Noun(), id),
		AllowedAnswersCount: 1,
		Options: []domain.PollOption{
			{Label: "Yes", Value: "yes"},
			{Label: "No", Value: "no"},
		},
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func NewPollAnswer(userID, pollID uuid.UUID, option string, opts ...func(*domain.PollAnswer)) *domain.PollAnswer {
	a := &domain.PollAnswer{
		UserID: userID,
		PollID: pollID,
		Option: option,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}
