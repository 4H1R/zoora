package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/4H1R/zoora/internal/domain"
)

const openaiDefaultBase = "https://api.openai.com/v1"

type openai struct {
	cfg AdapterConfig
	hc  *http.Client
}

// NewOpenAI builds an OpenAI (chat-completions) adapter implementing domain.LLM.
func NewOpenAI(cfg AdapterConfig) domain.LLM {
	if cfg.BaseURL == "" {
		cfg.BaseURL = openaiDefaultBase
	}
	return &openai{cfg: cfg, hc: httpClient(cfg.Timeout, cfg.ProxyURL)}
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type openaiRespFormat struct {
	Type string `json:"type"`
}
type openaiRequest struct {
	Model          string            `json:"model"`
	Messages       []openaiMessage   `json:"messages"`
	MaxTokens      int               `json:"max_tokens,omitempty"`
	Temperature    float32           `json:"temperature,omitempty"`
	ResponseFormat *openaiRespFormat `json:"response_format,omitempty"`
}
type openaiResponse struct {
	Choices []struct {
		Message openaiMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (o *openai) Generate(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	msgs := make([]openaiMessage, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, openaiMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, openaiMessage{Role: string(m.Role), Content: m.Content})
	}
	or := openaiRequest{
		Model:       o.cfg.Model,
		Messages:    msgs,
		MaxTokens:   o.cfg.MaxTokens,
		Temperature: req.Temperature,
	}
	if req.JSONMode {
		or.ResponseFormat = &openaiRespFormat{Type: "json_object"}
	}
	headers := map[string]string{"Authorization": "Bearer " + o.cfg.APIKey}
	url := o.cfg.BaseURL + "/chat/completions"

	var out openaiResponse
	if err := doJSON(ctx, o.hc, url, headers, or, &out); err != nil {
		return domain.LLMResponse{}, err
	}
	if len(out.Choices) == 0 {
		return domain.LLMResponse{}, fmt.Errorf("llm(openai): empty response")
	}
	return domain.LLMResponse{
		Text:  out.Choices[0].Message.Content,
		Model: o.cfg.Model,
		Usage: domain.LLMUsage{PromptTokens: out.Usage.PromptTokens, CompletionTokens: out.Usage.CompletionTokens},
	}, nil
}
