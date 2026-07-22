package llm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/llm"
)

func TestAnthropicGenerateParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "k" {
			t.Errorf("missing x-api-key: %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"content":[{"type":"text","text":"{\"ok\":true}"}],
			"usage":{"input_tokens":9,"output_tokens":4}
		}`))
	}))
	defer srv.Close()

	client := llm.NewAnthropic(llm.AdapterConfig{APIKey: "k", Model: "claude-3-5-haiku", BaseURL: srv.URL, MaxTokens: 256})
	resp, err := client.Generate(context.Background(), domain.LLMRequest{
		System:   "grade",
		Messages: []domain.LLMMessage{{Role: domain.LLMRoleUser, Content: "answer"}},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Text != `{"ok":true}` || resp.Usage.PromptTokens != 9 || resp.Usage.CompletionTokens != 4 {
		t.Fatalf("bad parse: %+v", resp)
	}
}
