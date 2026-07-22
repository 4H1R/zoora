package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/4H1R/zoora/internal/domain"
)

const geminiDefaultBase = "https://generativelanguage.googleapis.com/v1beta"

type gemini struct {
	cfg AdapterConfig
	hc  *http.Client
}

// NewGemini builds a Gemini adapter implementing domain.LLM.
func NewGemini(cfg AdapterConfig) domain.LLM {
	if cfg.BaseURL == "" {
		cfg.BaseURL = geminiDefaultBase
	}
	return &gemini{cfg: cfg, hc: httpClient(cfg.Timeout, cfg.ProxyURL)}
}

type geminiPart struct {
	Text string `json:"text"`
}
type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}
type geminiGenConfig struct {
	MaxOutputTokens  int     `json:"maxOutputTokens,omitempty"`
	Temperature      float32 `json:"temperature,omitempty"`
	ResponseMIMEType string  `json:"responseMimeType,omitempty"`
}
type geminiRequest struct {
	SystemInstruction *geminiContent   `json:"systemInstruction,omitempty"`
	Contents          []geminiContent  `json:"contents"`
	GenerationConfig  *geminiGenConfig `json:"generationConfig,omitempty"`
}
type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

func (g *gemini) Generate(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	contents := make([]geminiContent, 0, len(req.Messages))
	for _, m := range req.Messages {
		role := "user"
		if m.Role != domain.LLMRoleUser {
			role = "user" // Gemini only accepts user/model turns; system goes in systemInstruction
		}
		contents = append(contents, geminiContent{Role: role, Parts: []geminiPart{{Text: m.Content}}})
	}
	gr := geminiRequest{
		Contents: contents,
		GenerationConfig: &geminiGenConfig{
			MaxOutputTokens: g.cfg.MaxTokens,
			Temperature:     req.Temperature,
		},
	}
	if req.System != "" {
		gr.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: req.System}}}
	}
	if req.JSONMode {
		gr.GenerationConfig.ResponseMIMEType = "application/json"
	}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.cfg.BaseURL, g.cfg.Model, g.cfg.APIKey)

	var out geminiResponse
	if err := doJSON(ctx, g.hc, url, nil, gr, &out); err != nil {
		return domain.LLMResponse{}, err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return domain.LLMResponse{}, fmt.Errorf("llm(gemini): empty response")
	}
	return domain.LLMResponse{
		Text:  out.Candidates[0].Content.Parts[0].Text,
		Model: g.cfg.Model,
		Usage: domain.LLMUsage{
			PromptTokens:     out.UsageMetadata.PromptTokenCount,
			CompletionTokens: out.UsageMetadata.CandidatesTokenCount,
		},
	}, nil
}
