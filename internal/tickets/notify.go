package tickets

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// Notifier implements the service's notifier port via
// domain.NotificationService.SendSystem (system send: no sender, no rate
// limit). Mirrors internal/conversations/notify.go.
type Notifier struct {
	notifications domain.NotificationService
}

func NewNotifier(n domain.NotificationService) *Notifier {
	return &Notifier{notifications: n}
}

func (nx *Notifier) TicketCreated(ctx context.Context, t *domain.Ticket, body string, teacherID uuid.UUID) error {
	return nx.send(ctx, t, body, teacherID)
}

func (nx *Notifier) TicketReplied(ctx context.Context, t *domain.Ticket, m *domain.TicketMessage, recipientID uuid.UUID) error {
	return nx.send(ctx, t, m.Body, recipientID)
}

func (nx *Notifier) send(ctx context.Context, t *domain.Ticket, body string, recipientID uuid.UUID) error {
	if r := []rune(body); len(r) > 140 { // rune-safe truncation
		body = string(r[:140])
	}
	action := fmt.Sprintf("/org/tickets?ticket=%s", t.ID.String())
	orgID := t.OrganizationID
	return nx.notifications.SendSystem(ctx, domain.SystemNotificationInput{
		OrganizationID: &orgID,
		Category:       domain.NotificationCategoryClass,
		Title:          t.Title,
		Body:           body,
		ActionURL:      &action,
		Audience:       domain.NotificationAudience{Type: domain.AudienceUsers, UserIDs: []uuid.UUID{recipientID}},
	})
}
