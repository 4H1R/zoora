package quizzes

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// callerWithAI builds a caller on a plan that includes FeatureAI (Pro/Max).
func callerWithAI(t *testing.T) domain.Caller {
	t.Helper()
	return domain.Caller{
		UserID: uuid.New(),
		Ent:    domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
	}
}

func TestStartAIGradingRequiresFeature(t *testing.T) {
	svc := &service{} // llm nil, no deps needed — gate must fail before any use
	caller := domain.Caller{} // zero-value ⇒ Free plan ⇒ no FeatureAI
	ctx := domain.WithCaller(context.Background(), caller)

	_, err := svc.StartAIGrading(ctx, uuid.New(), domain.StartAIGradingDTO{Mode: domain.AIGradingModeSuggest})
	if err == nil {
		t.Fatal("expected feature-gate error for a Free-plan caller")
	}
}

func TestStartAIGradingRequiresConfiguredLLM(t *testing.T) {
	// A caller WITH the feature but no configured provider (llm nil) must get a
	// clear error, not a nil-pointer panic.
	svc := &service{llm: nil}
	caller := callerWithAI(t)
	ctx := domain.WithCaller(context.Background(), caller)

	_, err := svc.StartAIGrading(ctx, uuid.New(), domain.StartAIGradingDTO{Mode: domain.AIGradingModeSuggest})
	if err == nil {
		t.Fatal("expected error when LLM provider is not configured")
	}
}
