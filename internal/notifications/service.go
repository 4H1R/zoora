package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

// QueueName isolates notification fan-out from critical live-session tasks.
const QueueName = "notifications"

// Enqueuer is the subset of the queue client the service needs. It is an
// interface (not the concrete *queue.Client) so that fan-out's downstream task
// enqueueing is observable in tests: a service wired with a nil queue silently
// created delivery rows but never enqueued the send tasks, leaving them stuck
// "pending" in prod. *queue.Client satisfies this.
type Enqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// Authorization always happens in the service layer so handlers stay thin.
// Send matrix: superAdmin → any audience; notifications:send_any → own-org
// audiences; notifications:send → owned classes and their members.
// Senders holds the external delivery ports. A nil field means that channel is
// disabled: fan-out still snapshots delivery rows for enabled channels only via
// the connector repo, and channel task handlers mark rows failed when their
// sender is nil.
type Senders struct {
	Telegram domain.BotSender // nil = channel disabled
	Bale     domain.BotSender
	SMS      domain.SMSSender
	Push     domain.PushSender
}

type service struct {
	repo          domain.NotificationRepository
	classRepo     domain.ClassRepository
	connectorRepo domain.UserConnectorRepository
	orgSettings   domain.OrganizationSettingsProvider
	orgRepo       domain.OrganizationRepository
	queue         Enqueuer
	senders       Senders
	ratePerHour   int
	rdb           *redis.Client // nil disables unread-count caching (unit tests)
	logger        *slog.Logger
}

func NewService(
	repo domain.NotificationRepository,
	classRepo domain.ClassRepository,
	connectorRepo domain.UserConnectorRepository, // nil ok: fan-out skips external channels
	orgSettings domain.OrganizationSettingsProvider, // nil ok: SMS gate treated as disabled
	orgRepo domain.OrganizationRepository, // nil ok: org title omitted from delivered messages
	queueClient Enqueuer,
	senders Senders,
	ratePerHour int,
	rdb *redis.Client, // nil ok: disables unread-count caching
	logger *slog.Logger,
) domain.NotificationService {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{
		repo:          repo,
		classRepo:     classRepo,
		connectorRepo: connectorRepo,
		orgSettings:   orgSettings,
		orgRepo:       orgRepo,
		queue:         queueClient,
		senders:       senders,
		ratePerHour:   ratePerHour,
		rdb:           rdb,
		logger:        logger,
	}
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

// SendSystem creates a system/reminder notification without a human caller.
// It skips CallerFromCtx, authorizeAudience, and rate-limiting — the caller is
// trusted server-side code (schedulers).
func (s *service) SendSystem(ctx context.Context, in domain.SystemNotificationInput) error {
	category := in.Category
	if category == "" {
		category = domain.NotificationCategoryReminder
	}
	n := &domain.Notification{
		SenderID:       nil, // system
		OrganizationID: in.OrganizationID,
		Category:       category,
		Title:          in.Title,
		Body:           in.Body,
		ActionURL:      in.ActionURL,
		Audience:       in.Audience,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}
	return s.enqueueFanout(ctx, n.ID)
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

// Status returns the caller's unread count. The bell badge polls this every
// ~30s per client, so it is cached behind a short TTL and invalidated on every
// write that changes the count (mark-read, mark-all-read, fan-out).
func (s *service) Status(ctx context.Context) (*domain.NotificationStatus, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	if s.rdb != nil {
		if unread, err := cache.GetUnreadCount(ctx, s.rdb, caller.UserID); err == nil {
			return &domain.NotificationStatus{UnreadCount: unread}, nil
		}
	}

	unread, err := s.repo.CountUnread(ctx, caller.UserID)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		if err := cache.SetUnreadCount(ctx, s.rdb, caller.UserID, unread); err != nil {
			s.logger.WarnContext(ctx, "caching unread count", "error", err)
		}
	}
	return &domain.NotificationStatus{UnreadCount: unread}, nil
}

func (s *service) MarkRead(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	// Server clock — never trust a client-sent timestamp.
	if err := s.repo.MarkRead(ctx, id, caller.UserID, time.Now()); err != nil {
		return err
	}
	s.invalidateUnread(ctx, caller.UserID)
	return nil
}

func (s *service) MarkAllRead(ctx context.Context) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	if err := s.repo.MarkAllRead(ctx, caller.UserID, time.Now()); err != nil {
		return err
	}
	s.invalidateUnread(ctx, caller.UserID)
	return nil
}

// invalidateUnread drops a user's cached unread count. Best-effort: the short
// TTL is the backstop, so a redis error is logged, not surfaced.
func (s *service) invalidateUnread(ctx context.Context, userID uuid.UUID) {
	if s.rdb == nil {
		return
	}
	if err := cache.InvalidateUnreadCount(ctx, s.rdb, userID); err != nil {
		s.logger.WarnContext(ctx, "invalidating unread count", "error", err, "user_id", userID)
	}
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
	recipientIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if n.SenderID != nil && id == *n.SenderID {
			continue // senders don't notify themselves
		}
		recipients = append(recipients, domain.NotificationRecipient{
			NotificationID: n.ID,
			UserID:         id,
		})
		recipientIDs = append(recipientIDs, id)
	}
	if err := s.repo.CreateRecipients(ctx, recipients); err != nil {
		return err
	}
	// New inbox rows raise these users' unread counts — drop their cached badge
	// so the next poll reflects the notification within the TTL window.
	if s.rdb != nil && len(recipientIDs) > 0 {
		if err := cache.InvalidateUnreadCounts(ctx, s.rdb, recipientIDs); err != nil {
			s.logger.WarnContext(ctx, "invalidating unread counts after fan-out", "error", err)
		}
	}
	s.logger.Info("notification fan-out complete",
		"notification_id", n.ID, "recipients", len(recipients))

	return s.fanoutDeliveries(ctx, n, recipientIDs)
}

// fanoutDeliveries snapshots per-channel delivery rows for the recipients'
// verified, enabled connectors and enqueues channel-send tasks. No-op when the
// connector repo isn't wired (unit tests / API-side construction).
func (s *service) fanoutDeliveries(ctx context.Context, n *domain.Notification, recipientIDs []uuid.UUID) error {
	if s.connectorRepo == nil {
		return nil
	}
	conns, err := s.connectorRepo.ListVerifiedEnabledByUsers(ctx, recipientIDs)
	if err != nil {
		return err
	}
	smsAllowed := n.OrganizationID == nil // system notifications: platform pays
	if !smsAllowed && s.orgSettings != nil {
		settings, err := s.orgSettings.GetByOrgID(ctx, *n.OrganizationID)
		if err == nil && settings != nil {
			smsAllowed = settings.SMSEnabled
		}
	}
	deliveries := make([]domain.NotificationDelivery, 0, len(conns))
	for _, c := range conns {
		if c.Type == domain.ConnectorSMS && !smsAllowed {
			continue
		}
		deliveries = append(deliveries, domain.NotificationDelivery{
			NotificationID: n.ID,
			UserID:         c.UserID,
			Channel:        c.Type,
			Target:         c.Target,
		})
	}
	if err := s.repo.CreateDeliveries(ctx, deliveries); err != nil {
		return err
	}
	return s.enqueueDeliveries(ctx, n.ID)
}

const (
	smsBatchSize  = 100
	pushBatchSize = 500
)

// enqueueDeliveries reads back pending rows (IDs are needed post-upsert) and
// enqueues channel tasks: one per row for bots, batched for SMS and push.
func (s *service) enqueueDeliveries(ctx context.Context, notificationID uuid.UUID) error {
	if s.queue == nil {
		return nil
	}
	for _, ch := range []domain.ConnectorType{domain.ConnectorTelegram, domain.ConnectorBale} {
		rows, err := s.repo.ListPendingDeliveries(ctx, notificationID, ch)
		if err != nil {
			return err
		}
		for _, row := range rows {
			payload, err := json.Marshal(domain.NotificationDeliverBotPayload{DeliveryID: row.ID})
			if err != nil {
				return fmt.Errorf("notifications.service.enqueueDeliveries bot payload: %w", err)
			}
			task := asynq.NewTask(domain.TypeNotificationDeliverBot, payload)
			if _, err := s.queue.Enqueue(task, asynq.Queue(QueueName), asynq.MaxRetry(5)); err != nil {
				return fmt.Errorf("notifications.service.enqueueDeliveries bot: %w", err)
			}
		}
	}

	if err := s.enqueueBatched(ctx, notificationID, domain.ConnectorSMS, domain.TypeNotificationDeliverSMS, smsBatchSize); err != nil {
		return err
	}
	return s.enqueueBatched(ctx, notificationID, domain.ConnectorPush, domain.TypeNotificationDeliverPush, pushBatchSize)
}

func (s *service) enqueueBatched(ctx context.Context, notificationID uuid.UUID, ch domain.ConnectorType, taskType string, batchSize int) error {
	rows, err := s.repo.ListPendingDeliveries(ctx, notificationID, ch)
	if err != nil {
		return err
	}
	for start := 0; start < len(rows); start += batchSize {
		end := min(start+batchSize, len(rows))
		ids := make([]uuid.UUID, 0, end-start)
		for _, row := range rows[start:end] {
			ids = append(ids, row.ID)
		}
		payload, err := json.Marshal(domain.NotificationDeliverBatchPayload{
			NotificationID: notificationID,
			DeliveryIDs:    ids,
		})
		if err != nil {
			return fmt.Errorf("notifications.service.enqueueBatched payload: %w", err)
		}
		task := asynq.NewTask(taskType, payload)
		if _, err := s.queue.Enqueue(task, asynq.Queue(QueueName), asynq.MaxRetry(5)); err != nil {
			return fmt.Errorf("notifications.service.enqueueBatched: %w", err)
		}
	}
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

// smsMaxLen caps SMS body at ~2 segments to bound provider cost.
const smsMaxLen = 320

// Report returns the delivery report. Only the notification's sender or a
// superAdmin may view it.
func (s *service) Report(ctx context.Context, notificationID uuid.UUID) (*domain.NotificationDeliveryReport, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	n, err := s.repo.FindByID(ctx, notificationID)
	if err != nil {
		return nil, err
	}
	if !caller.IsAdmin && (n.SenderID == nil || *n.SenderID != caller.UserID) {
		return nil, domain.ErrForbidden
	}
	recipients, err := s.repo.CountRecipients(ctx, notificationID)
	if err != nil {
		return nil, err
	}
	channels, err := s.repo.DeliveryReport(ctx, notificationID)
	if err != nil {
		return nil, err
	}
	return &domain.NotificationDeliveryReport{Recipients: recipients, Channels: channels}, nil
}

func botSender(senders Senders, ch domain.ConnectorType) domain.BotSender {
	switch ch {
	case domain.ConnectorTelegram:
		return senders.Telegram
	case domain.ConnectorBale:
		return senders.Bale
	default:
		return nil
	}
}

// DeliverBot sends one telegram/bale message for a single delivery row.
func (s *service) DeliverBot(ctx context.Context, deliveryID uuid.UUID) error {
	deliveries, err := s.repo.ListDeliveriesByIDs(ctx, []uuid.UUID{deliveryID})
	if err != nil {
		return err
	}
	if len(deliveries) == 0 {
		return nil // row gone (notification deleted) — nothing to do
	}
	d := deliveries[0]
	n, err := s.repo.FindByID(ctx, d.NotificationID)
	if err != nil {
		return err
	}
	sender := botSender(s.senders, d.Channel)
	if sender == nil {
		return s.markFailed(ctx, []uuid.UUID{d.ID}, "channel disabled")
	}
	if err := sender.SendMessage(ctx, d.Target, botMessage(n, s.orgTitle(ctx, n))); err != nil {
		// Return the error so Asynq retries; row stays pending until success or
		// retry exhaustion (visible in the report).
		return fmt.Errorf("notifications.service.DeliverBot: %w", err)
	}
	return s.repo.MarkDeliveries(ctx, []uuid.UUID{d.ID}, domain.DeliverySent, nil, time.Now())
}

// DeliverSMS sends one bulk SMS to a batch of delivery rows.
func (s *service) DeliverSMS(ctx context.Context, notificationID uuid.UUID, deliveryIDs []uuid.UUID) error {
	deliveries, err := s.repo.ListDeliveriesByIDs(ctx, deliveryIDs)
	if err != nil {
		return err
	}
	if len(deliveries) == 0 {
		return nil
	}
	n, err := s.repo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}
	ids := make([]uuid.UUID, 0, len(deliveries))
	phones := make([]string, 0, len(deliveries))
	for _, d := range deliveries {
		ids = append(ids, d.ID)
		phones = append(phones, d.Target)
	}
	if s.senders.SMS == nil {
		return s.markFailed(ctx, ids, "channel disabled")
	}
	if err := s.senders.SMS.SendBulk(ctx, phones, smsMessage(n, s.orgTitle(ctx, n))); err != nil {
		return fmt.Errorf("notifications.service.DeliverSMS: %w", err)
	}
	return s.repo.MarkDeliveries(ctx, ids, domain.DeliverySent, nil, time.Now())
}

// DeliverPush sends one FCM multicast to a batch of delivery rows, pruning
// tokens FCM reports as permanently invalid.
func (s *service) DeliverPush(ctx context.Context, notificationID uuid.UUID, deliveryIDs []uuid.UUID) error {
	deliveries, err := s.repo.ListDeliveriesByIDs(ctx, deliveryIDs)
	if err != nil {
		return err
	}
	if len(deliveries) == 0 {
		return nil
	}
	n, err := s.repo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}
	byToken := make(map[string]uuid.UUID, len(deliveries))
	allIDs := make([]uuid.UUID, 0, len(deliveries))
	tokens := make([]string, 0, len(deliveries))
	for _, d := range deliveries {
		byToken[d.Target] = d.ID
		allIDs = append(allIDs, d.ID)
		tokens = append(tokens, d.Target)
	}
	if s.senders.Push == nil {
		return s.markFailed(ctx, allIDs, "channel disabled")
	}
	link := "/"
	if n.ActionURL != nil && *n.ActionURL != "" {
		link = *n.ActionURL
	}
	invalidTokens, err := s.senders.Push.SendMulticast(ctx, tokens, displayTitle(s.orgTitle(ctx, n), n.Title), n.Body, link)
	if err != nil {
		return fmt.Errorf("notifications.service.DeliverPush: %w", err)
	}
	invalidIDs := make([]uuid.UUID, 0, len(invalidTokens))
	invalidSet := make(map[uuid.UUID]struct{}, len(invalidTokens))
	for _, tok := range invalidTokens {
		if id, ok := byToken[tok]; ok {
			invalidIDs = append(invalidIDs, id)
			invalidSet[id] = struct{}{}
		}
		if err := s.connectorRepo.DeleteByTypeTarget(ctx, domain.ConnectorPush, tok); err != nil {
			return err
		}
	}
	if err := s.markFailed(ctx, invalidIDs, "token unregistered"); err != nil {
		return err
	}
	sentIDs := make([]uuid.UUID, 0, len(allIDs)-len(invalidIDs))
	for _, id := range allIDs {
		if _, bad := invalidSet[id]; !bad {
			sentIDs = append(sentIDs, id)
		}
	}
	return s.repo.MarkDeliveries(ctx, sentIDs, domain.DeliverySent, nil, time.Now())
}

func (s *service) markFailed(ctx context.Context, ids []uuid.UUID, reason string) error {
	if len(ids) == 0 {
		return nil
	}
	msg := reason
	return s.repo.MarkDeliveries(ctx, ids, domain.DeliveryFailed, &msg, time.Time{})
}

// orgTitle resolves the sending organization's name for the given notification,
// so recipients can tell which org a delivered message is from. Returns "" for
// system notifications (no org) or when the org repo isn't wired / lookup fails
// — delivery must never be blocked just because the title couldn't be resolved.
func (s *service) orgTitle(ctx context.Context, n *domain.Notification) string {
	if s.orgRepo == nil || n.OrganizationID == nil {
		return ""
	}
	org, err := s.orgRepo.FindByID(ctx, *n.OrganizationID)
	if err != nil || org == nil {
		s.logger.WarnContext(ctx, "resolving org title for notification delivery",
			"error", err, "notification_id", n.ID, "org_id", *n.OrganizationID)
		return ""
	}
	return org.Name
}

// displayTitle prefixes the notification title with the organization name so the
// org is visible in the recipient's push/connector message. No-op when orgName
// is empty (system notifications).
func displayTitle(orgName, title string) string {
	if orgName == "" {
		return title
	}
	return orgName + " · " + title
}

func botMessage(n *domain.Notification, orgName string) string {
	msg := displayTitle(orgName, n.Title) + "\n\n" + n.Body
	if n.ActionURL != nil && *n.ActionURL != "" {
		msg += "\n" + *n.ActionURL
	}
	return msg
}

func smsMessage(n *domain.Notification, orgName string) string {
	msg := displayTitle(orgName, n.Title) + "\n" + n.Body
	// Rune-aware truncation: SMS bodies are often Persian (multibyte), so
	// slicing by byte could split a character.
	if r := []rune(msg); len(r) > smsMaxLen {
		msg = string(r[:smsMaxLen])
	}
	return msg
}
