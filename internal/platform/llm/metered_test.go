package llm_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/llm"
)

type fakeLLM struct {
	resp domain.LLMResponse
}

func (f fakeLLM) Generate(context.Context, domain.LLMRequest) (domain.LLMResponse, error) {
	return f.resp, nil
}

type recorderSpy struct{ events []domain.AIUsageEvent }

func (r *recorderSpy) Record(_ context.Context, ev domain.AIUsageEvent) error {
	r.events = append(r.events, ev)
	return nil
}

func TestMeteredRecordsUsage(t *testing.T) {
	org := uuid.New()
	inner := fakeLLM{resp: domain.LLMResponse{
		Text:  "graded",
		Model: "gemini-2.0-flash",
		Usage: domain.LLMUsage{PromptTokens: 1_000_000, CompletionTokens: 1_000_000},
	}}
	spy := &recorderSpy{}
	metered := llm.NewMetered(inner, spy, "gemini")

	_, err := metered.Generate(context.Background(), domain.LLMRequest{
		Feature:        "ai_grading",
		OrganizationID: org,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(spy.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(spy.events))
	}
	ev := spy.events[0]
	if ev.OrganizationID != org || ev.Feature != "ai_grading" || ev.Provider != "gemini" {
		t.Fatalf("event context wrong: %+v", ev)
	}
	if ev.Model != "gemini-2.0-flash" || ev.CostMicros != 500 {
		t.Fatalf("expected cost 500 for known model, got model=%q cost=%d", ev.Model, ev.CostMicros)
	}
}
