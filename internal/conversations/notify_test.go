package conversations

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func TestUnmutedRecipients_ExcludesMutedAndSender(t *testing.T) {
	sender, a, b, c := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	members := []domain.ConversationMember{
		{UserID: sender},                 // excluded: sender
		{UserID: a},                      // included: not muted
		{UserID: b, MutedUntil: &future}, // excluded: muted in the future
		{UserID: c, MutedUntil: &past},   // included: mute already expired
	}

	got := unmutedRecipients(members, sender, now)

	if len(got) != 2 {
		t.Fatalf("expected 2 recipients, got %d: %v", len(got), got)
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range got {
		seen[id] = true
	}
	if !seen[a] {
		t.Errorf("expected unmuted member A to be included")
	}
	if !seen[c] {
		t.Errorf("expected member C with expired mute to be included")
	}
	if seen[sender] {
		t.Errorf("sender must be excluded")
	}
	if seen[b] {
		t.Errorf("currently-muted member B must be excluded")
	}
}
