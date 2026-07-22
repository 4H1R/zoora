package llm_test

import (
	"testing"

	"github.com/4H1R/zoora/internal/platform/llm"
)

func TestCostMicrosKnownModel(t *testing.T) {
	// gemini-2.0-flash: 100 in-micros/MTok, 400 out-micros/MTok (see pricing.go).
	// 1,000,000 prompt + 1,000,000 completion tokens => 100 + 400 = 500 micros.
	got := llm.CostMicros("gemini-2.0-flash", 1_000_000, 1_000_000)
	if got != 500 {
		t.Fatalf("expected 500 micros, got %d", got)
	}
}

func TestCostMicrosUnknownModelIsZero(t *testing.T) {
	if got := llm.CostMicros("mystery-model", 1000, 1000); got != 0 {
		t.Fatalf("unknown model should cost 0, got %d", got)
	}
}
