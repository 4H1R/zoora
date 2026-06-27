package attendance

import (
	"testing"

	"github.com/google/uuid"
)

func TestComputePresentByPercent(t *testing.T) {
	u1 := uuid.New()
	u2 := uuid.New()
	u3 := uuid.New()

	t.Run("marks users at or above threshold present", func(t *testing.T) {
		total := 1000
		userSeconds := map[uuid.UUID]int{
			u1: 800, // 80% -> present
			u2: 750, // 75% -> present (>=)
			u3: 740, // 74% -> absent
		}
		present, ok := computePresentByPercent(total, userSeconds, 75)
		if !ok {
			t.Fatalf("expected ok=true")
		}
		got := map[uuid.UUID]bool{}
		for _, id := range present {
			got[id] = true
		}
		if !got[u1] || !got[u2] || got[u3] {
			t.Fatalf("unexpected present set: %v", present)
		}
	})

	t.Run("zero total duration cannot compute", func(t *testing.T) {
		_, ok := computePresentByPercent(0, map[uuid.UUID]int{u1: 10}, 75)
		if ok {
			t.Fatalf("expected ok=false for zero total duration")
		}
	})

	t.Run("user beyond total is capped at present", func(t *testing.T) {
		present, ok := computePresentByPercent(100, map[uuid.UUID]int{u1: 250}, 75)
		if !ok || len(present) != 1 || present[0] != u1 {
			t.Fatalf("expected u1 present, got %v ok=%v", present, ok)
		}
	})
}
