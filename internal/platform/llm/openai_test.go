package llm_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/llm"
)

func TestOpenAIGenerateParsesResponse(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("missing bearer auth: %q", r.Header.Get("Authorization"))
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"content":"{\"ok\":true}"}}],
			"usage":{"prompt_tokens":8,"completion_tokens":3}
		}`))
	}))
	defer srv.Close()

	client := llm.NewOpenAI(llm.AdapterConfig{APIKey: "k", Model: "gpt-4o-mini", BaseURL: srv.URL, MaxTokens: 256})
	resp, err := client.Generate(context.Background(), domain.LLMRequest{
		System:   "grade",
		Messages: []domain.LLMMessage{{Role: domain.LLMRoleUser, Content: "answer"}},
		JSONMode: true,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Text != `{"ok":true}` || resp.Usage.PromptTokens != 8 || resp.Usage.CompletionTokens != 3 {
		t.Fatalf("bad parse: %+v", resp)
	}
	msgs, _ := body["messages"].([]any)
	if len(msgs) != 2 {
		t.Fatalf("expected system+user messages, got %d", len(msgs))
	}
}
