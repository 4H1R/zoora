package quizzes

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/domain"
)

func TestSanitizeQuestionForTaking_ImageModeWithholdsText(t *testing.T) {
	bodyImg := uuid.New()
	optImg := uuid.New()
	q := domain.Question{
		Type:               domain.QuestionTypeChoice,
		Text:               "secret question text",
		ModelAnswer:        "the answer",
		RenderAsImage:      true,
		SystemImageMediaID: &bodyImg,
		Options: []domain.QuestionOption{
			{ID: "a", Value: "correct answer", Score: 5, SystemImageMediaID: &optImg},
		},
	}

	got := sanitizeQuestionForTaking(q)

	assert.Empty(t, got.Text, "body text must be withheld in image mode")
	assert.Empty(t, got.Options[0].Value, "option value must be withheld in image mode")
	assert.Zero(t, got.Options[0].Score, "score is always stripped")
	assert.Empty(t, got.ModelAnswer)
	// Image ids are what the client renders — they must survive.
	assert.Equal(t, &bodyImg, got.SystemImageMediaID)
	assert.Equal(t, &optImg, got.Options[0].SystemImageMediaID)
	assert.Equal(t, "a", got.Options[0].ID, "option id kept for answering/grading")
}

func TestSanitizeQuestionForTaking_NonImageKeepsText(t *testing.T) {
	q := domain.Question{
		Type: domain.QuestionTypeChoice,
		Text: "visible question",
		Options: []domain.QuestionOption{
			{ID: "a", Value: "visible option", Score: 5},
		},
	}

	got := sanitizeQuestionForTaking(q)

	assert.Equal(t, "visible question", got.Text)
	assert.Equal(t, "visible option", got.Options[0].Value)
	assert.Zero(t, got.Options[0].Score, "score still stripped")
}
