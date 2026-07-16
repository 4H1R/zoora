package domain_test

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

func TestAIGradingConstants(t *testing.T) {
	if domain.AIGradingModeApply != "apply" || domain.AIGradingModeSuggest != "suggest" {
		t.Fatal("mode constants changed")
	}
	if domain.GradedByAI != "ai" || domain.GradedByManual != "manual" {
		t.Fatal("graded_by constants changed")
	}
	if domain.AIAnswerStatusFailed != "failed" {
		t.Fatal("answer status constants changed")
	}
}

func TestSubmissionAnswerHasAIFields(t *testing.T) {
	var a domain.SubmissionAnswer
	a.GradedBy = domain.GradedByManual
	a.AIStatus = domain.AIAnswerStatusScored
	score := 3.5
	a.SuggestedScore = &score
	a.AIRationale = "matches key points"
	if a.GradedBy != "manual" || *a.SuggestedScore != 3.5 {
		t.Fatal("AI fields not wired")
	}
}
