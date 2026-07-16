package llm

import (
	"context"
	"log/slog"

	"github.com/4H1R/zoora/internal/domain"
)

// metered wraps an LLM and records one domain.AIUsageEvent per successful call.
// Recording failures are logged, never propagated — metering must not break a
// user-facing grade.
type metered struct {
	inner    domain.LLM
	recorder domain.AIUsageRecorder
	provider string
	logger   *slog.Logger
}

// NewMetered wraps inner so every call writes a usage event. recorder may be
// nil (metering disabled). provider is the active provider name (e.g. "gemini").
func NewMetered(inner domain.LLM, recorder domain.AIUsageRecorder, provider string) domain.LLM {
	return &metered{inner: inner, recorder: recorder, provider: provider, logger: slog.Default()}
}

func (m *metered) Generate(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	resp, err := m.inner.Generate(ctx, req)
	if err != nil {
		return resp, err
	}
	if m.recorder == nil {
		return resp, nil
	}
	ev := domain.AIUsageEvent{
		OrganizationID:   req.OrganizationID,
		Feature:          req.Feature,
		Provider:         m.provider,
		Model:            resp.Model,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		CostMicros:       CostMicros(resp.Model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens),
	}
	// Use a background context so recording still completes if the caller's ctx
	// is already cancelled after a successful generation.
	if recErr := m.recorder.Record(context.WithoutCancel(ctx), ev); recErr != nil {
		m.logger.Error("ai usage record failed", "org", req.OrganizationID.String(), "error", recErr)
	}
	return resp, nil
}
