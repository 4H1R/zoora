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
}

func NewNotifier(n domain.NotificationService) *Notifier {
	return &Notifier{notifications: n}
}

// unmutedRecipients returns member UserIDs excluding the sender and any
// member whose MutedUntil is still in the future. Pure/no I/O; afterSend
// applies it once to the roster it already fetched, so the realtime per-user
// fanout and the notification recipients share one mute filter.
func unmutedRecipients(members []domain.ConversationMember, sender uuid.UUID, now time.Time) []uuid.UUID {
	var out []uuid.UUID
	for _, m := range members {
		if m.UserID == sender {
			continue
		}
		if m.MutedUntil != nil && m.MutedUntil.After(now) {
			continue
		}
		out = append(out, m.UserID)
	}
	return out
}

// NotifyMessage sends in-app + push to the resolved recipients. Recipients
// arrive already sender-excluded and mute-filtered (afterSend owns that
// filter). Errors are the caller's to log-and-continue on.
func (nx *Notifier) NotifyMessage(ctx context.Context, conv *domain.Conversation, msg *domain.ConversationMessage, recipientIDs []uuid.UUID) error {
	targets := recipientIDs
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
