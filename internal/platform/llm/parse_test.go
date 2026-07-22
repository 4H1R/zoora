package llm_test

import (
	"testing"

	"github.com/4H1R/zoora/internal/platform/llm"
)

func TestExtractJSON(t *testing.T) {
	cases := map[string]string{
		"plain":        `{"a":1}`,
		"fenced":       "```json\n{\"a\":1}\n```",
		"fenced_plain": "```\n{\"a\":1}\n```",
		"prose_around": "Sure, here it is:\n{\"a\":1}\nHope that helps.",
		"nested":       `{"scores":[{"question_id":"x","score":3}]}`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := llm.ExtractJSON(in)
			if err != nil {
				t.Fatalf("ExtractJSON(%q): %v", in, err)
			}
			if out[0] != '{' || out[len(out)-1] != '}' {
				t.Fatalf("not a JSON object: %q", out)
			}
		})
	}
}

func TestExtractJSONNoObject(t *testing.T) {
	if _, err := llm.ExtractJSON("no json here"); err == nil {
		t.Fatal("expected error when no JSON object present")
	}
}
