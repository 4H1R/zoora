package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/queue"
)

// QueueName isolates notification fan-out from critical live-session tasks.
const QueueName = "notifications"

// Authorization always happens in the service layer so handlers stay thin.
// Send matrix: superAdmin → any audience; notifications:send_any → own-org
// audiences; notifications:send → owned classes and their members.
type service struct {
	repo        domain.NotificationRepository
	classRepo   domain.ClassRepository
	queue       *queue.Client
	ratePerHour int
	logger      *slog.Logger
}

func NewService(
	repo domain.NotificationRepository,
	classRepo domain.ClassRepository,
	queueClient *queue.Client,
	ratePerHour int,
	logger *slog.Logger,
) domain.NotificationService {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{repo: repo, classRepo: classRepo, queue: queueClient, ratePerHour: ratePerHour, logger: logger}
}

func (s *service) Send(ctx context.Context, dto domain.SendNotificationDTO) (*domain.Notification, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	audience := domain.NotificationAudience{
		Type:    domain.NotificationAudienceType(dto.Audience.Type),
		OrgID:   dto.Audience.OrgID,
		ClassID: dto.Audience.ClassID,
		RoleID:  dto.Audience.RoleID,
		UserIDs: dto.Audience.UserIDs,
	}
	// A non-admin org audience is always scoped to the caller's own org, so the
	// request's org_id is ignored and pre-filled here — org_id stays optional
	// in the DTO for managers while authorizeAudience re-forces it below.
	if !caller.IsAdmin && audience.Type == domain.AudienceOrg && caller.OrgID != nil {
		audience.OrgID = caller.OrgID
	}
	if err := audience.Validate(); err != nil {
		return nil, err
	}

	if err := s.authorizeAudience(ctx, caller, &audience); err != nil {
		return nil, err
	}

	if !caller.IsAdmin && s.ratePerHour > 0 {
		n, err := s.repo.CountBySenderSince(ctx, caller.UserID, time.Now().Add(-time.Hour))
		if err != nil {
			return nil, err
		}
		if n >= int64(s.ratePerHour) {
			return nil, domain.ErrRateLimited
		}
	}

	notification := &domain.Notification{
		SenderID:       &caller.UserID,
		OrganizationID: caller.OrgID,
		Category:       deriveCategory(caller, audience),
		Title:          dto.Title,
		Body:           dto.Body,
		ActionURL:      dto.ActionURL,
		Audience:       audience,
	}
	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	if err := s.enqueueFanout(ctx, notification.ID); err != nil {
		// Notification row exists; fan-out can be re-triggered. Surface the error.
		return nil, err
	}
	return notification, nil
}

// authorizeAudience mutates the audience where scope is forced (manager org).
func (s *service) authorizeAudience(ctx context.Context, caller domain.Caller, a *domain.NotificationAudience) error {
	if caller.IsAdmin {
		if a.Type == domain.AudienceOrg && a.OrgID == nil {
			return domain.NewValidationError(map[string]string{"audience.org_id": "required"})
		}
		if a.Type == domain.AudienceClass {
			if _, err := s.classRepo.FindByID(ctx, *a.ClassID); err != nil {
				return err
			}
		}
		if a.Type == domain.AudienceRole {
			ok, err := s.repo.RoleExistsInScope(ctx, *a.RoleID, a.OrgID)
			if err != nil {
				return err
			}
			if !ok {
				return domain.ErrNotFound
			}
		}
		if a.Type == domain.AudienceUsers {
			n, err := s.repo.CountActiveUsersByIDs(ctx, a.UserIDs, nil)
			if err != nil {
				return err
			}
			if n != int64(len(a.UserIDs)) {
				return domain.NewValidationError(map[string]string{"audience.user_ids": "contains unknown or inactive users"})
			}
		}
		return nil
	}

	canAny := caller.HasPermission(domain.PermNotificationsSendAny)
	canOwn := caller.HasPermission(domain.PermNotificationsSend)
	if !canAny && !canOwn {
		return domain.ErrForbidden
	}
	if caller.OrgID == nil {
		return domain.ErrForbidden
	}

	switch a.Type {
	case domain.AudienceAll:
		return domain.ErrForbidden // platform-wide is superAdmin-only

	case domain.AudienceOrg:
		if !canAny {
			return domain.ErrForbidden
		}
		a.OrgID = caller.OrgID // forced: tenant scoping by construction

	case domain.AudienceClass:
		class, err := s.classRepo.FindByID(ctx, *a.ClassID)
		if err != nil {
			return err
		}
		if class.OrganizationID != *caller.OrgID {
			return domain.ErrForbidden
		}
		if !canAny && class.UserID != caller.UserID {
			return domain.ErrForbidden
		}

	case domain.AudienceRole:
		if !canAny {
			return domain.ErrForbidden
		}
		ok, err := s.repo.RoleExistsInScope(ctx, *a.RoleID, caller.OrgID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}
		a.OrgID = caller.OrgID // role resolution stays inside the org

	case domain.AudienceUsers:
		if canAny {
			n, err := s.repo.CountActiveUsersByIDs(ctx, a.UserIDs, caller.OrgID)
			if err != nil {
				return err
			}
			if n != int64(len(a.UserIDs)) {
				return domain.ErrForbidden
			}
		} else {
			n, err := s.repo.CountUsersInClassesOwnedBy(ctx, a.UserIDs, caller.UserID)
			if err != nil {
				return err
			}
			if n != int64(len(a.UserIDs)) {
				return domain.ErrForbidden
			}
		}
	}
	return nil
}

func deriveCategory(caller domain.Caller, a domain.NotificationAudience) domain.NotificationCategory {
	switch {
	case caller.IsAdmin:
		return domain.NotificationCategorySystem
	case a.Type == domain.AudienceClass:
		return domain.NotificationCategoryClass
	default:
		return domain.NotificationCategoryOrg
	}
}

func (s *service) enqueueFanout(ctx context.Context, notificationID uuid.UUID) error {
	if s.queue == nil {
		return nil // unit tests / worker-side construction
	}
	payload, err := json.Marshal(domain.NotificationFanoutPayload{NotificationID: notificationID})
	if err != nil {
		return fmt.Errorf("notifications.service.enqueueFanout payload: %w", err)
	}
	task := asynq.NewTask(domain.TypeNotificationFanout, payload)
	if _, err := s.queue.Enqueue(task, asynq.Queue(QueueName), asynq.MaxRetry(5)); err != nil {
		return fmt.Errorf("notifications.service.enqueueFanout: %w", err)
	}
	return nil
}

func (s *service) ListInbox(ctx context.Context, p domain.ListParams) ([]domain.NotificationInboxItem, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	return s.repo.ListInbox(ctx, caller.UserID, p.Limit(), p.Offset())
}

func (s *service) Status(ctx context.Context) (*domain.NotificationStatus, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	unread, err := s.repo.CountUnread(ctx, caller.UserID)
	if err != nil {
		return nil, err
	}
	return &domain.NotificationStatus{UnreadCount: unread}, nil
}

func (s *service) MarkRead(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	// Server clock — never trust a client-sent timestamp.
	return s.repo.MarkRead(ctx, id, caller.UserID, time.Now())
}

func (s *service) MarkAllRead(ctx context.Context) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	return s.repo.MarkAllRead(ctx, caller.UserID, time.Now())
}

func (s *service) ListSent(ctx context.Context, p domain.ListParams) ([]domain.Notification, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	return s.repo.ListBySender(ctx, caller.UserID, p.Limit(), p.Offset())
}

// Fanout resolves the audience to active user IDs and inserts inbox rows.
// Idempotent: CreateRecipients ignores conflicts, so Asynq retries are safe.
func (s *service) Fanout(ctx context.Context, notificationID uuid.UUID) error {
	n, err := s.repo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}
	ids, err := s.resolveAudience(ctx, n)
	if err != nil {
		return err
	}
	recipients := make([]domain.NotificationRecipient, 0, len(ids))
	for _, id := range ids {
		if n.SenderID != nil && id == *n.SenderID {
			continue // senders don't notify themselves
		}
		recipients = append(recipients, domain.NotificationRecipient{
			NotificationID: n.ID,
			UserID:         id,
		})
	}
	if err := s.repo.CreateRecipients(ctx, recipients); err != nil {
		return err
	}
	s.logger.Info("notification fan-out complete",
		"notification_id", n.ID, "recipients", len(recipients))
	return nil
}

func (s *service) resolveAudience(ctx context.Context, n *domain.Notification) ([]uuid.UUID, error) {
	a := n.Audience
	switch a.Type {
	case domain.AudienceAll:
		return s.repo.ListAllActiveUserIDs(ctx)
	case domain.AudienceOrg:
		return s.repo.ListUserIDsByOrg(ctx, *a.OrgID)
	case domain.AudienceClass:
		return s.repo.ListUserIDsByClass(ctx, *a.ClassID)
	case domain.AudienceRole:
		return s.repo.ListUserIDsByRole(ctx, *a.RoleID, a.OrgID)
	case domain.AudienceUsers:
		return a.UserIDs, nil
	default:
		return nil, fmt.Errorf("notifications.service.resolveAudience: unknown audience type %q", a.Type)
	}
}

// Report / DeliverBot / DeliverSMS / DeliverPush are the channel-delivery
// entries. Full behavior is added by the delivery pipeline (Task 6); these
// placeholders satisfy the interface until then.

func (s *service) Report(ctx context.Context, notificationID uuid.UUID) (*domain.NotificationDeliveryReport, error) {
	return nil, fmt.Errorf("notifications.service.Report: not implemented")
}

func (s *service) DeliverBot(ctx context.Context, deliveryID uuid.UUID) error {
	return fmt.Errorf("notifications.service.DeliverBot: not implemented")
}

func (s *service) DeliverSMS(ctx context.Context, notificationID uuid.UUID, deliveryIDs []uuid.UUID) error {
	return fmt.Errorf("notifications.service.DeliverSMS: not implemented")
}

func (s *service) DeliverPush(ctx context.Context, notificationID uuid.UUID, deliveryIDs []uuid.UUID) error {
	return fmt.Errorf("notifications.service.DeliverPush: not implemented")
}
