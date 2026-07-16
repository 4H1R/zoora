package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/4H1R/zoora/internal/domain"
)

// AdapterConfig configures a single provider adapter.
type AdapterConfig struct {
	APIKey    string
	Model     string
	BaseURL   string // optional override; each adapter has a sane default
	MaxTokens int
	Timeout   time.Duration
}

func httpClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{Timeout: timeout}
}

// doJSON marshals reqBody, POSTs it, and unmarshals a 2xx JSON response into out.
// Non-2xx returns an error including a truncated body for diagnosis.
func doJSON(ctx context.Context, hc *http.Client, url string, headers map[string]string, reqBody, out any) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("llm: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("llm: http: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("llm: provider status %d: %s", resp.StatusCode, string(raw))
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("llm: decode response: %w", err)
	}
	return nil
}

// New builds the active LLM client from config, wrapping the selected provider
// adapter with metering and concurrency limiting. Returns (nil, nil) when no
// provider/API key is configured — the caller treats a nil LLM as "AI disabled".
func New(cfg AdapterConfig, provider string, recorder domain.AIUsageRecorder, maxConcurrent int) (domain.LLM, error) {
	if provider == "" || cfg.APIKey == "" {
		return nil, nil
	}
	var adapter domain.LLM
	switch provider {
	case "gemini":
		adapter = NewGemini(cfg)
	case "openai":
		adapter = NewOpenAI(cfg)
	case "anthropic":
		adapter = NewAnthropic(cfg)
	default:
		return nil, fmt.Errorf("llm: unknown provider %q", provider)
	}
	return NewLimiter(NewMetered(adapter, recorder, provider), maxConcurrent), nil
}
