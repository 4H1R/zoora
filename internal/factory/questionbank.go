package factory

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewQuestionBank(orgID uuid.UUID, opts ...func(*domain.QuestionBank)) *domain.QuestionBank {
	id := nextID()
	qb := &domain.QuestionBank{
		OrganizationID: orgID,
		Name:           fmt.Sprintf("%s Question Bank %d", fake.Noun(), id),
		Description:    fake.Sentence(8),
	}
	for _, o := range opts {
		o(qb)
	}
	return qb
}

func NewQuestionOption() domain.QuestionOption {
	return domain.QuestionOption{
		ID:    uuid.New().String(),
		Value: fake.Sentence(3),
		Score: float64(fake.IntRange(0, 5)),
	}
}

func NewQuestion(bankID, orgID uuid.UUID, opts ...func(*domain.Question)) *domain.Question {
	questionTypes := []domain.QuestionType{
		domain.QuestionTypeChoice,
		domain.QuestionTypeShortAnswer,
		domain.QuestionTypeDescriptive,
	}
	id := nextID()
	q := &domain.Question{
		BankID:         bankID,
		OrganizationID: orgID,
		Text:           fmt.Sprintf("%s? (%d)", fake.Question(), id),
		Type:           questionTypes[id%3],
		Metadata:       []domain.QuestionMetadata{},
	}
	for _, o := range opts {
		o(q)
	}
	if len(q.Options) == 0 {
		switch q.Type {
		case domain.QuestionTypeChoice:
			count := fake.IntRange(3, 5)
			q.Options = make([]domain.QuestionOption, count)
			for i := range count {
				q.Options[i] = NewQuestionOption()
			}
			q.Options[0].Score = float64(fake.IntRange(1, 5))
		case domain.QuestionTypeShortAnswer:
			opt := NewQuestionOption()
			if opt.Score <= 0 {
				opt.Score = float64(fake.IntRange(1, 5))
			}
			q.Options = []domain.QuestionOption{opt}
		case domain.QuestionTypeDescriptive:
			q.Options = []domain.QuestionOption{{
				ID:    uuid.New().String(),
				Score: float64(fake.IntRange(1, 5)),
			}}
		}
	}
	if q.Metadata == nil {
		q.Metadata = []domain.QuestionMetadata{}
	}
	return q
}
