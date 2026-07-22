package llm

import (
	"context"

	"github.com/4H1R/zoora/internal/domain"
)

// limiter caps the number of concurrent in-flight Generate calls across all
// consumers, protecting the provider's requests-per-minute ceiling. It is a
// process-global semaphore — one per active LLM client.
type limiter struct {
	inner domain.LLM
	sem   chan struct{}
}

// NewLimiter wraps an LLM so at most maxConcurrent calls run at once.
// maxConcurrent <= 0 disables limiting (returns inner unchanged).
func NewLimiter(inner domain.LLM, maxConcurrent int) domain.LLM {
	if maxConcurrent <= 0 {
		return inner
	}
	return &limiter{inner: inner, sem: make(chan struct{}, maxConcurrent)}
}

func (l *limiter) Generate(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	select {
	case l.sem <- struct{}{}:
		defer func() { <-l.sem }()
	case <-ctx.Done():
		return domain.LLMResponse{}, ctx.Err()
	}
	return l.inner.Generate(ctx, req)
}
