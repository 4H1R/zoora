package domain_test

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

func TestLLMMessageRolesAreStable(t *testing.T) {
	if domain.LLMRoleSystem != "system" || domain.LLMRoleUser != "user" {
		t.Fatalf("role constants changed: %q %q", domain.LLMRoleSystem, domain.LLMRoleUser)
	}
}

func TestLLMRequestZeroValueIsUsable(t *testing.T) {
	req := domain.LLMRequest{
		System:   "you are a grader",
		Messages: []domain.LLMMessage{{Role: domain.LLMRoleUser, Content: "hi"}},
	}
	if len(req.Messages) != 1 || req.Messages[0].Role != domain.LLMRoleUser {
		t.Fatal("request not constructable")
	}
}
