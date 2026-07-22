package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/4H1R/zoora/internal/domain"
)

const (
	anthropicDefaultBase = "https://api.anthropic.com/v1"
	anthropicVersion     = "2023-06-01"
)

type anthropic struct {
	cfg AdapterConfig
	hc  *http.Client
}

// NewAnthropic builds an Anthropic (messages API) adapter implementing domain.LLM.
func NewAnthropic(cfg AdapterConfig) domain.LLM {
	if cfg.BaseURL == "" {
		cfg.BaseURL = anthropicDefaultBase
	}
	return &anthropic{cfg: cfg, hc: httpClient(cfg.Timeout, cfg.ProxyURL)}
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type anthropicRequest struct {
	Model       string             `json:"model"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float32            `json:"temperature,omitempty"`
}
type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (a *anthropic) Generate(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	msgs := make([]anthropicMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, anthropicMessage{Role: string(m.Role), Content: m.Content})
	}
	maxTok := a.cfg.MaxTokens
	if maxTok <= 0 {
		maxTok = 512 // Anthropic requires max_tokens > 0
	}
	ar := anthropicRequest{
		Model:       a.cfg.Model,
		System:      req.System,
		Messages:    msgs,
		MaxTokens:   maxTok,
		Temperature: req.Temperature,
	}
	headers := map[string]string{
		"x-api-key":         a.cfg.APIKey,
		"anthropic-version": anthropicVersion,
	}
	url := a.cfg.BaseURL + "/messages"

	var out anthropicResponse
	if err := doJSON(ctx, a.hc, url, headers, ar, &out); err != nil {
		return domain.LLMResponse{}, err
	}
	if len(out.Content) == 0 {
		return domain.LLMResponse{}, fmt.Errorf("llm(anthropic): empty response")
	}
	return domain.LLMResponse{
		Text:  out.Content[0].Text,
		Model: a.cfg.Model,
		Usage: domain.LLMUsage{PromptTokens: out.Usage.InputTokens, CompletionTokens: out.Usage.OutputTokens},
	}, nil
}
