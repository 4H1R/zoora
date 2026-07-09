package conversations

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// Notifier implements the service's notifier port using
// domain.NotificationService.SendSystem — the in-app/push fan-out entry
// point that bypasses caller auth + rate limiting (this is a system send).
type Notifier struct {
	notifications domain.NotificationService
	members       domain.ConversationMemberRepository
}

func NewNotifier(n domain.NotificationService, members domain.ConversationMemberRepository) *Notifier {
	return &Notifier{notifications: n, members: members}
}

// NotifyMessage sends in-app + push to the resolved recipients, skipping
// muted members. Errors are the caller's to log-and-continue on.
func (nx *Notifier) NotifyMessage(ctx context.Context, conv *domain.Conversation, msg *domain.ConversationMessage, recipientIDs []uuid.UUID) error {
	if len(recipientIDs) == 0 {
		return nil
	}
	// Filter muted.
	now := time.Now()
	var targets []uuid.UUID
	for _, uid := range recipientIDs {
		m, err := nx.members.FindByConversationAndUser(ctx, conv.ID, uid)
		if err != nil {
			continue
		}
		if m.MutedUntil != nil && m.MutedUntil.After(now) {
			continue
		}
		targets = append(targets, uid)
	}
	if len(targets) == 0 {
		return nil
	}
	title := conv.Name
	if conv.Type == domain.ConversationTypeDirect {
		title = "New message" // client localizes; DM has no name
	}
	body := msg.Content
	if r := []rune(body); len(r) > 140 { // rune-safe: byte slicing would split UTF-8
		body = string(r[:140])
	}
	action := fmt.Sprintf("/chat/%s", conv.ID.String())
	orgID := conv.OrganizationID
	return nx.notifications.SendSystem(ctx, domain.SystemNotificationInput{
		OrganizationID: &orgID,
		Category:       domain.NotificationCategoryOrg,
		Title:          title,
		Body:           body,
		ActionURL:      &action,
		Audience:       domain.NotificationAudience{Type: domain.AudienceUsers, UserIDs: targets},
	})
}
