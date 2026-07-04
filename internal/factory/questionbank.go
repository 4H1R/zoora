package factory

import (
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewQuestionBank(orgID uuid.UUID, opts ...func(*domain.QuestionBank)) *domain.QuestionBank {
	id := nextID()
	qb := &domain.QuestionBank{
		OrganizationID: orgID,
		Name:           fakeQuestionBankName(id),
		Description:    fakeSentence(8),
	}
	for _, o := range opts {
		o(qb)
	}
	return qb
}

func NewQuestionOption() domain.QuestionOption {
	return domain.QuestionOption{
		ID:    uuid.New().String(),
		Value: fakeSentence(3),
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
		Text:           fakeQuestionText(id),
		Type:           questionTypes[id%3],
		Metadata:       []domain.QuestionMetadata{},
		MinSeconds:     fake.IntRange(0, 20),
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
	// Negative marking only applies to choice questions; occasionally set a mode.
	if q.Type == domain.QuestionTypeChoice && q.NegativeMarkMode == "" && fake.Bool() {
		if fake.Bool() {
			q.NegativeMarkMode = domain.NegativeMarkPerWrong
			q.NegativeValue = domain.FractionFor(len(q.Options))
		} else {
			n := min(max(len(q.Options), 2), 5)
			q.NegativeMarkMode = domain.NegativeMarkAccumulative
			q.WrongsPerPoint = n
		}
	}
	if q.NegativeMarkMode == "" {
		q.NegativeMarkMode = domain.NegativeMarkNone
	}
	if q.Metadata == nil {
		q.Metadata = []domain.QuestionMetadata{}
	}
	return q
}
