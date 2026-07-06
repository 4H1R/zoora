package factory

import (
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewQAQuestion(userID uuid.UUID, modelType string, modelID uuid.UUID, opts ...func(*domain.QAQuestion)) *domain.QAQuestion {
	q := &domain.QAQuestion{
		UserID:    userID,
		ModelType: modelType,
		ModelID:   modelID,
		Text:      T("What is the deadline?", "مهلت چیست؟"),
		Status:    domain.QAStatusOpen,
	}
	for _, o := range opts {
		o(q)
	}
	return q
}

func NewQAVote(questionID, userID uuid.UUID, opts ...func(*domain.QAVote)) *domain.QAVote {
	v := &domain.QAVote{QuestionID: questionID, UserID: userID}
	for _, o := range opts {
		o(v)
	}
	return v
}
