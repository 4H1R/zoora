package llm_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/llm"
)

func TestGeminiGenerateParsesResponse(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		if !strings.Contains(r.URL.Path, "gemini-2.0-flash") {
			t.Errorf("model not in path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"candidates":[{"content":{"parts":[{"text":"{\"ok\":true}"}]}}],
			"usageMetadata":{"promptTokenCount":12,"candidatesTokenCount":5}
		}`))
	}))
	defer srv.Close()

	client := llm.NewGemini(llm.AdapterConfig{
		APIKey:    "k",
		Model:     "gemini-2.0-flash",
		BaseURL:   srv.URL,
		MaxTokens: 256,
	})

	resp, err := client.Generate(context.Background(), domain.LLMRequest{
		System:   "grade it",
		Messages: []domain.LLMMessage{{Role: domain.LLMRoleUser, Content: "answer"}},
		JSONMode: true,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Text != `{"ok":true}` {
		t.Fatalf("unexpected text: %q", resp.Text)
	}
	if resp.Usage.PromptTokens != 12 || resp.Usage.CompletionTokens != 5 {
		t.Fatalf("usage not parsed: %+v", resp.Usage)
	}
	if resp.Model != "gemini-2.0-flash" {
		t.Fatalf("model not set: %q", resp.Model)
	}
	// System instruction must go in its own channel, not the user turn.
	if _, ok := gotBody["systemInstruction"]; !ok {
		t.Fatal("systemInstruction missing from request")
	}
}
