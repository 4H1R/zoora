package conversations

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// HubMembership adapts the member repo to the chathub's narrow read port.
type HubMembership struct {
	members domain.ConversationMemberRepository
}

func NewHubMembership(m domain.ConversationMemberRepository) *HubMembership {
	return &HubMembership{members: m}
}

func (h *HubMembership) IsMember(ctx context.Context, convID, userID uuid.UUID) (bool, error) {
	_, err := h.members.FindByConversationAndUser(ctx, convID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (h *HubMembership) ListUserIDs(ctx context.Context, convID uuid.UUID) ([]uuid.UUID, error) {
	return h.members.ListUserIDs(ctx, convID)
}
