package notifications

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// fakeQueue records enqueued task types so tests can assert that fan-out
// actually schedules the per-channel send tasks (the prod bug: a nil queue
// created delivery rows but silently enqueued nothing).
type fakeQueue struct{ enqueued []string }

func (q *fakeQueue) Enqueue(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	q.enqueued = append(q.enqueued, task.Type())
	return &asynq.TaskInfo{}, nil
}

func (q *fakeQueue) countByType(taskType string) int {
	n := 0
	for _, t := range q.enqueued {
		if t == taskType {
			n++
		}
	}
	return n
}

type mockRepo struct {
	domain.NotificationRepository // panic on unstubbed calls

	created        *domain.Notification
	senderCount    int64
	usersInOrg     int64
	usersInClasses int64
	roleInScope    bool

	found        *domain.Notification
	recipients   []domain.NotificationRecipient
	orgUserIDs   []uuid.UUID
	classUserIDs []uuid.UUID
	unread       int64
	markedRead   *uuid.UUID

	deliveriesCreated []domain.NotificationDelivery
	deliveriesByID    []domain.NotificationDelivery
	marked            []markedCall
	recipientCount    int64
}

type markedCall struct {
	ids    []uuid.UUID
	status domain.NotificationDeliveryStatus
	errMsg *string
}

func (m *mockRepo) CreateDeliveries(_ context.Context, d []domain.NotificationDelivery) error {
	m.deliveriesCreated = append(m.deliveriesCreated, d...)
	return nil
}

// ListPendingDeliveries mimics the post-insert read-back: it returns the
// created rows for the requested channel (with an ID assigned, as the real
// upsert would) so enqueueDeliveries can schedule a send task per row.
func (m *mockRepo) ListPendingDeliveries(_ context.Context, _ uuid.UUID, channel domain.ConnectorType) ([]domain.NotificationDelivery, error) {
	var rows []domain.NotificationDelivery
	for _, d := range m.deliveriesCreated {
		if d.Channel == channel {
			if d.ID == uuid.Nil {
				d.ID = uuid.New()
			}
			rows = append(rows, d)
		}
	}
	return rows, nil
}

func (m *mockRepo) ListDeliveriesByIDs(context.Context, []uuid.UUID) ([]domain.NotificationDelivery, error) {
	return m.deliveriesByID, nil
}

func (m *mockRepo) MarkDeliveries(_ context.Context, ids []uuid.UUID, status domain.NotificationDeliveryStatus, errMsg *string, _ time.Time) error {
	m.marked = append(m.marked, markedCall{ids: ids, status: status, errMsg: errMsg})
	return nil
}

func (m *mockRepo) CountRecipients(context.Context, uuid.UUID) (int64, error) {
	return m.recipientCount, nil
}

func (m *mockRepo) DeliveryReport(context.Context, uuid.UUID) ([]domain.NotificationChannelReport, error) {
	return nil, nil
}

type mockConnectorRepo struct {
	domain.UserConnectorRepository
	conns   []domain.UserConnector
	deleted []string
}

func (m *mockConnectorRepo) ListVerifiedEnabledByUsers(context.Context, []uuid.UUID) ([]domain.UserConnector, error) {
	return m.conns, nil
}

func (m *mockConnectorRepo) DeleteByTypeTarget(_ context.Context, _ domain.ConnectorType, target string) error {
	m.deleted = append(m.deleted, target)
	return nil
}

type mockOrgSettings struct{ smsEnabled bool }

func (m mockOrgSettings) GetByOrgID(context.Context, uuid.UUID) (*domain.OrganizationSettings, error) {
	return &domain.OrganizationSettings{SMSEnabled: m.smsEnabled}, nil
}

type fakeBotSender struct {
	sent    int
	lastMsg string
}

func (f *fakeBotSender) SendMessage(_ context.Context, _, msg string) error {
	f.sent++
	f.lastMsg = msg
	return nil
}

type fakePushSender struct {
	invalid   []string
	lastTitle string
}

func (f *fakePushSender) SendMulticast(_ context.Context, _ []string, title, _, _ string) ([]string, error) {
	f.lastTitle = title
	return f.invalid, nil
}

// mockOrgRepo resolves an org name for delivery-message tests. Only FindByID is
// exercised; the rest satisfy domain.OrganizationRepository.
type mockOrgRepo struct{ org *domain.Organization }

func (m mockOrgRepo) FindByID(context.Context, uuid.UUID) (*domain.Organization, error) {
	return m.org, nil
}
func (mockOrgRepo) Create(context.Context, *domain.Organization) error { return nil }
func (mockOrgRepo) FindBySlug(context.Context, string) (*domain.Organization, error) {
	return nil, nil
}
func (mockOrgRepo) Update(context.Context, *domain.Organization) error { return nil }
func (mockOrgRepo) Delete(context.Context, uuid.UUID) error            { return nil }
func (mockOrgRepo) List(context.Context, domain.OrganizationFilter) ([]domain.Organization, int64, error) {
	return nil, 0, nil
}
func (mockOrgRepo) GetStats(context.Context) (*domain.OrganizationStats, error) { return nil, nil }
func (mockOrgRepo) AdminList(context.Context, domain.AdminListOrganizationsQuery) ([]domain.Organization, int64, error) {
	return nil, 0, nil
}
func (mockOrgRepo) HardDelete(context.Context, uuid.UUID) error { return nil }
func (mockOrgRepo) Restore(context.Context, uuid.UUID) error    { return nil }
func (mockOrgRepo) UpdatePlan(context.Context, uuid.UUID, domain.Plan, *time.Time) error {
	return nil
}

func (m *mockRepo) Create(_ context.Context, n *domain.Notification) error {
	n.ID = uuid.New()
	m.created = n
	return nil
}

func (m *mockRepo) CountBySenderSince(context.Context, uuid.UUID, time.Time) (int64, error) {
	return m.senderCount, nil
}

func (m *mockRepo) CountActiveUsersByIDs(context.Context, []uuid.UUID, *uuid.UUID) (int64, error) {
	return m.usersInOrg, nil
}

func (m *mockRepo) CountUsersInClassesOwnedBy(context.Context, []uuid.UUID, uuid.UUID) (int64, error) {
	return m.usersInClasses, nil
}

func (m *mockRepo) RoleExistsInScope(context.Context, uuid.UUID, *uuid.UUID) (bool, error) {
	return m.roleInScope, nil
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Notification, error) {
	if m.found == nil {
		return nil, domain.ErrNotFound
	}
	return m.found, nil
}

func (m *mockRepo) CreateRecipients(_ context.Context, r []domain.NotificationRecipient) error {
	m.recipients = append(m.recipients, r...)
	return nil
}

func (m *mockRepo) ListUserIDsByOrg(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return m.orgUserIDs, nil
}

func (m *mockRepo) ListUserIDsByClass(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return m.classUserIDs, nil
}
func (m *mockRepo) CountUnread(context.Context, uuid.UUID) (int64, error) { return m.unread, nil }
func (m *mockRepo) MarkRead(_ context.Context, nID, uID uuid.UUID, t time.Time) error {
	m.markedRead = &nID
	return nil
}

type mockClassRepo struct {
	domain.ClassRepository
	class *domain.Class
}

func (m *mockClassRepo) FindByID(context.Context, uuid.UUID) (*domain.Class, error) {
	if m.class == nil {
		return nil, domain.ErrNotFound
	}
	return m.class, nil
}

func adminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
}

func managerCtx(orgID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(), OrgID: &orgID,
		Permissions: []string{string(domain.PermNotificationsSend), string(domain.PermNotificationsSendAny)},
	})
}

func teacherCtx(orgID uuid.UUID) (context.Context, uuid.UUID) {
	id := uuid.New()
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: id, OrgID: &orgID,
		Permissions: []string{string(domain.PermNotificationsSend)},
	}), id
}

func ctxWithCaller() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New()})
}

func dto(a domain.NotificationAudienceDTO) domain.SendNotificationDTO {
	return domain.SendNotificationDTO{Title: "t", Body: "b", Audience: a}
}

func TestSendAdminAllDerivesSystemCategory(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	n, err := svc.Send(adminCtx(), dto(domain.NotificationAudienceDTO{Type: "all"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Category != domain.NotificationCategorySystem {
		t.Fatalf("Category = %s, want system", n.Category)
	}
}

func TestSendManagerAllForbidden(t *testing.T) {
	orgID := uuid.New()
	svc := NewService(&mockRepo{}, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	_, err := svc.Send(managerCtx(orgID), dto(domain.NotificationAudienceDTO{Type: "all"}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSendManagerOrgForcedToOwnOrg(t *testing.T) {
	orgID, otherOrg := uuid.New(), uuid.New()
	repo := &mockRepo{}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	n, err := svc.Send(managerCtx(orgID), dto(domain.NotificationAudienceDTO{Type: "org", OrgID: &otherOrg}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Audience.OrgID == nil || *n.Audience.OrgID != orgID {
		t.Fatalf("audience org = %v, want caller org %v", n.Audience.OrgID, orgID)
	}
	if n.Category != domain.NotificationCategoryOrg {
		t.Fatalf("Category = %s, want org", n.Category)
	}
}

func TestSendManagerClassOutsideOrgForbidden(t *testing.T) {
	orgID := uuid.New()
	classID := uuid.New()
	classRepo := &mockClassRepo{class: &domain.Class{ID: classID, OrganizationID: uuid.New(), UserID: uuid.New()}}
	svc := NewService(&mockRepo{}, classRepo, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	_, err := svc.Send(managerCtx(orgID), dto(domain.NotificationAudienceDTO{Type: "class", ClassID: &classID}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSendTeacherOwnClassAllowed(t *testing.T) {
	orgID := uuid.New()
	ctx, teacherID := teacherCtx(orgID)
	classID := uuid.New()
	classRepo := &mockClassRepo{class: &domain.Class{ID: classID, OrganizationID: orgID, UserID: teacherID}}
	repo := &mockRepo{}
	svc := NewService(repo, classRepo, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	n, err := svc.Send(ctx, dto(domain.NotificationAudienceDTO{Type: "class", ClassID: &classID}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Category != domain.NotificationCategoryClass {
		t.Fatalf("Category = %s, want class", n.Category)
	}
}

func TestSendTeacherForeignClassForbidden(t *testing.T) {
	orgID := uuid.New()
	ctx, _ := teacherCtx(orgID)
	classID := uuid.New()
	classRepo := &mockClassRepo{class: &domain.Class{ID: classID, OrganizationID: orgID, UserID: uuid.New()}}
	svc := NewService(&mockRepo{}, classRepo, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	_, err := svc.Send(ctx, dto(domain.NotificationAudienceDTO{Type: "class", ClassID: &classID}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSendTeacherUsersMustBeInOwnedClasses(t *testing.T) {
	orgID := uuid.New()
	ctx, _ := teacherCtx(orgID)
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	// Only 1 of 2 ids is in an owned class → forbidden.
	svc := NewService(&mockRepo{usersInClasses: 1}, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	_, err := svc.Send(ctx, dto(domain.NotificationAudienceDTO{Type: "users", UserIDs: ids}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSendManagerRoleOutsideOrgForbidden(t *testing.T) {
	orgID := uuid.New()
	roleID := uuid.New()
	svc := NewService(&mockRepo{roleInScope: false}, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	_, err := svc.Send(managerCtx(orgID), dto(domain.NotificationAudienceDTO{Type: "role", RoleID: &roleID}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSendRateLimited(t *testing.T) {
	orgID := uuid.New()
	repo := &mockRepo{senderCount: 10}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	_, err := svc.Send(managerCtx(orgID), dto(domain.NotificationAudienceDTO{Type: "org"}))
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}

func TestSendWithoutPermissionForbidden(t *testing.T) {
	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgID})
	svc := NewService(&mockRepo{}, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	_, err := svc.Send(ctx, dto(domain.NotificationAudienceDTO{Type: "org"}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestFanoutOrgAudienceExcludesSender(t *testing.T) {
	orgID := uuid.New()
	senderID := uuid.New()
	u1, u2 := uuid.New(), uuid.New()
	n := &domain.Notification{
		ID:       uuid.New(),
		SenderID: &senderID,
		Audience: domain.NotificationAudience{Type: domain.AudienceOrg, OrgID: &orgID},
	}
	repo := &mockRepo{found: n, orgUserIDs: []uuid.UUID{u1, u2, senderID}}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)

	if err := svc.Fanout(context.Background(), n.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.recipients) != 2 {
		t.Fatalf("recipients = %d, want 2 (sender excluded)", len(repo.recipients))
	}
	for _, r := range repo.recipients {
		if r.UserID == senderID {
			t.Fatal("sender must not receive their own notification")
		}
	}
}

func TestStatusReturnsUnread(t *testing.T) {
	repo := &mockRepo{unread: 7}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	st, err := svc.Status(ctxWithCaller())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.UnreadCount != 7 {
		t.Fatalf("UnreadCount = %d, want 7", st.UnreadCount)
	}
}

func TestStatusRequiresCaller(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	if _, err := svc.Status(context.Background()); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestFanoutCreatesDeliveriesPerConnector(t *testing.T) {
	orgID := uuid.New()
	u1 := uuid.New()
	n := &domain.Notification{
		ID:             uuid.New(),
		OrganizationID: &orgID,
		Audience:       domain.NotificationAudience{Type: domain.AudienceOrg, OrgID: &orgID},
	}
	repo := &mockRepo{found: n, orgUserIDs: []uuid.UUID{u1}}
	connRepo := &mockConnectorRepo{conns: []domain.UserConnector{
		{UserID: u1, Type: domain.ConnectorTelegram, Target: "111"},
		{UserID: u1, Type: domain.ConnectorPush, Target: "tok"},
	}}
	svc := NewService(repo, &mockClassRepo{}, connRepo, mockOrgSettings{smsEnabled: false}, nil, nil, Senders{}, 10, nil, nil)
	if err := svc.Fanout(context.Background(), n.ID); err != nil {
		t.Fatalf("Fanout: %v", err)
	}
	if len(repo.deliveriesCreated) != 2 {
		t.Fatalf("deliveries = %d, want 2", len(repo.deliveriesCreated))
	}
	for _, d := range repo.deliveriesCreated {
		if d.Channel == domain.ConnectorSMS {
			t.Fatal("no SMS delivery expected")
		}
	}
}

func TestFanoutSMSGatedByOrgSetting(t *testing.T) {
	orgID := uuid.New()
	u1 := uuid.New()
	n := &domain.Notification{
		ID:             uuid.New(),
		OrganizationID: &orgID,
		Audience:       domain.NotificationAudience{Type: domain.AudienceOrg, OrgID: &orgID},
	}
	repo := &mockRepo{found: n, orgUserIDs: []uuid.UUID{u1}}
	connRepo := &mockConnectorRepo{conns: []domain.UserConnector{
		{UserID: u1, Type: domain.ConnectorSMS, Target: "09120000001"},
	}}
	svc := NewService(repo, &mockClassRepo{}, connRepo, mockOrgSettings{smsEnabled: false}, nil, nil, Senders{}, 10, nil, nil)
	if err := svc.Fanout(context.Background(), n.ID); err != nil {
		t.Fatalf("Fanout: %v", err)
	}
	if len(repo.deliveriesCreated) != 0 {
		t.Fatalf("deliveries = %d, want 0 (SMS gated off)", len(repo.deliveriesCreated))
	}
}

func TestFanoutSystemNotificationAllowsSMS(t *testing.T) {
	u1 := uuid.New()
	n := &domain.Notification{
		ID:             uuid.New(),
		OrganizationID: nil, // system notification — platform pays
		Audience:       domain.NotificationAudience{Type: domain.AudienceUsers, UserIDs: []uuid.UUID{u1}},
	}
	repo := &mockRepo{found: n}
	connRepo := &mockConnectorRepo{conns: []domain.UserConnector{
		{UserID: u1, Type: domain.ConnectorSMS, Target: "09120000001"},
	}}
	svc := NewService(repo, &mockClassRepo{}, connRepo, mockOrgSettings{smsEnabled: false}, nil, nil, Senders{}, 10, nil, nil)
	if err := svc.Fanout(context.Background(), n.ID); err != nil {
		t.Fatalf("Fanout: %v", err)
	}
	if len(repo.deliveriesCreated) != 1 || repo.deliveriesCreated[0].Channel != domain.ConnectorSMS {
		t.Fatalf("deliveries = %+v, want 1 SMS", repo.deliveriesCreated)
	}
}

func TestDeliverBotMarksSent(t *testing.T) {
	nID := uuid.New()
	n := &domain.Notification{ID: nID, Title: "t", Body: "b"}
	d := domain.NotificationDelivery{ID: uuid.New(), NotificationID: nID, Channel: domain.ConnectorTelegram, Target: "111"}
	repo := &mockRepo{found: n, deliveriesByID: []domain.NotificationDelivery{d}}
	bot := &fakeBotSender{}
	svc := NewService(repo, &mockClassRepo{}, &mockConnectorRepo{}, mockOrgSettings{}, nil, nil, Senders{Telegram: bot}, 10, nil, nil)
	if err := svc.DeliverBot(context.Background(), d.ID); err != nil {
		t.Fatalf("DeliverBot: %v", err)
	}
	if bot.sent != 1 {
		t.Fatalf("bot.sent = %d, want 1", bot.sent)
	}
	if len(repo.marked) != 1 || repo.marked[0].status != domain.DeliverySent {
		t.Fatalf("marked = %+v, want one sent", repo.marked)
	}
}

// TestDeliverIncludesOrgTitle guards that the sending organization's name is
// prefixed onto the delivered bot message and push title so recipients can tell
// which org a notification came from.
func TestDeliverIncludesOrgTitle(t *testing.T) {
	orgID := uuid.New()
	nID := uuid.New()
	n := &domain.Notification{ID: nID, OrganizationID: &orgID, Title: "Exam tomorrow", Body: "b"}
	orgRepo := mockOrgRepo{org: &domain.Organization{ID: orgID, Name: "Acme School"}}

	botD := domain.NotificationDelivery{ID: uuid.New(), NotificationID: nID, Channel: domain.ConnectorTelegram, Target: "111"}
	botRepo := &mockRepo{found: n, deliveriesByID: []domain.NotificationDelivery{botD}}
	bot := &fakeBotSender{}
	botSvc := NewService(botRepo, &mockClassRepo{}, &mockConnectorRepo{}, mockOrgSettings{}, orgRepo, nil, Senders{Telegram: bot}, 10, nil, nil)
	if err := botSvc.DeliverBot(context.Background(), botD.ID); err != nil {
		t.Fatalf("DeliverBot: %v", err)
	}
	if want := "Acme School · Exam tomorrow"; !strings.HasPrefix(bot.lastMsg, want) {
		t.Fatalf("bot message = %q, want prefix %q", bot.lastMsg, want)
	}

	pushD := domain.NotificationDelivery{ID: uuid.New(), NotificationID: nID, Channel: domain.ConnectorPush, Target: "tok"}
	pushRepo := &mockRepo{found: n, deliveriesByID: []domain.NotificationDelivery{pushD}}
	push := &fakePushSender{}
	pushSvc := NewService(pushRepo, &mockClassRepo{}, &mockConnectorRepo{}, mockOrgSettings{}, orgRepo, nil, Senders{Push: push}, 10, nil, nil)
	if err := pushSvc.DeliverPush(context.Background(), nID, []uuid.UUID{pushD.ID}); err != nil {
		t.Fatalf("DeliverPush: %v", err)
	}
	if want := "Acme School · Exam tomorrow"; push.lastTitle != want {
		t.Fatalf("push title = %q, want %q", push.lastTitle, want)
	}
}

func TestDeliverPushPrunesInvalidTokens(t *testing.T) {
	nID := uuid.New()
	n := &domain.Notification{ID: nID, Title: "t", Body: "b"}
	d1 := domain.NotificationDelivery{ID: uuid.New(), NotificationID: nID, Channel: domain.ConnectorPush, Target: "good"}
	d2 := domain.NotificationDelivery{ID: uuid.New(), NotificationID: nID, Channel: domain.ConnectorPush, Target: "bad"}
	repo := &mockRepo{found: n, deliveriesByID: []domain.NotificationDelivery{d1, d2}}
	connRepo := &mockConnectorRepo{}
	push := &fakePushSender{invalid: []string{"bad"}}
	svc := NewService(repo, &mockClassRepo{}, connRepo, mockOrgSettings{}, nil, nil, Senders{Push: push}, 10, nil, nil)
	if err := svc.DeliverPush(context.Background(), nID, []uuid.UUID{d1.ID, d2.ID}); err != nil {
		t.Fatalf("DeliverPush: %v", err)
	}
	if len(connRepo.deleted) != 1 || connRepo.deleted[0] != "bad" {
		t.Fatalf("deleted = %+v, want [bad]", connRepo.deleted)
	}
	var sawFailed, sawSent bool
	for _, mk := range repo.marked {
		switch mk.status {
		case domain.DeliveryFailed:
			sawFailed = true
			if len(mk.ids) != 1 || mk.ids[0] != d2.ID {
				t.Fatalf("failed ids = %+v, want [%s]", mk.ids, d2.ID)
			}
		case domain.DeliverySent:
			sawSent = true
			if len(mk.ids) != 1 || mk.ids[0] != d1.ID {
				t.Fatalf("sent ids = %+v, want [%s]", mk.ids, d1.ID)
			}
		}
	}
	if !sawFailed || !sawSent {
		t.Fatalf("marks = %+v, want a failed and a sent", repo.marked)
	}
}

// TestFanoutEnqueuesSendTasks is the regression guard for the prod incident
// where the worker built the notification service with a nil queue: delivery
// rows were created but no send task was ever enqueued, so bot deliveries sat
// stuck "pending" forever. With a wired queue, fan-out must enqueue one
// deliver-bot task per bot row and a batched task for SMS/push.
func TestFanoutEnqueuesSendTasks(t *testing.T) {
	u1, u2 := uuid.New(), uuid.New()
	n := &domain.Notification{
		ID:             uuid.New(),
		OrganizationID: nil, // system notification → SMS allowed
		Audience:       domain.NotificationAudience{Type: domain.AudienceUsers, UserIDs: []uuid.UUID{u1, u2}},
	}
	repo := &mockRepo{found: n}
	connRepo := &mockConnectorRepo{conns: []domain.UserConnector{
		{UserID: u1, Type: domain.ConnectorTelegram, Target: "111"},
		{UserID: u2, Type: domain.ConnectorBale, Target: "222"},
		{UserID: u1, Type: domain.ConnectorPush, Target: "tok"},
		{UserID: u2, Type: domain.ConnectorSMS, Target: "09120000001"},
	}}
	q := &fakeQueue{}
	svc := NewService(repo, &mockClassRepo{}, connRepo, mockOrgSettings{smsEnabled: false}, nil, q, Senders{}, 10, nil, nil)

	if err := svc.Fanout(context.Background(), n.ID); err != nil {
		t.Fatalf("Fanout: %v", err)
	}
	if got := q.countByType(domain.TypeNotificationDeliverBot); got != 2 {
		t.Fatalf("deliver-bot tasks = %d, want 2 (telegram + bale)", got)
	}
	if got := q.countByType(domain.TypeNotificationDeliverPush); got != 1 {
		t.Fatalf("deliver-push tasks = %d, want 1", got)
	}
	if got := q.countByType(domain.TypeNotificationDeliverSMS); got != 1 {
		t.Fatalf("deliver-sms tasks = %d, want 1", got)
	}
}

// TestFanoutNilQueueCreatesRowsButEnqueuesNothing documents the footgun that
// caused the incident: a nil queue is a silent no-op for enqueueing, so rows
// are created but never sent. This is the exact state to avoid in production
// wiring — see cmd/worker/main.go, which must pass the real queue client.
func TestFanoutNilQueueCreatesRowsButEnqueuesNothing(t *testing.T) {
	u1 := uuid.New()
	n := &domain.Notification{
		ID:             uuid.New(),
		OrganizationID: nil,
		Audience:       domain.NotificationAudience{Type: domain.AudienceUsers, UserIDs: []uuid.UUID{u1}},
	}
	repo := &mockRepo{found: n}
	connRepo := &mockConnectorRepo{conns: []domain.UserConnector{
		{UserID: u1, Type: domain.ConnectorTelegram, Target: "111"},
	}}
	svc := NewService(repo, &mockClassRepo{}, connRepo, mockOrgSettings{}, nil, nil, Senders{}, 10, nil, nil)

	if err := svc.Fanout(context.Background(), n.ID); err != nil {
		t.Fatalf("Fanout: %v", err)
	}
	if len(repo.deliveriesCreated) != 1 {
		t.Fatalf("deliveries created = %d, want 1 (rows still snapshotted)", len(repo.deliveriesCreated))
	}
}

func TestReportRequiresSenderOrAdmin(t *testing.T) {
	nID := uuid.New()
	senderID := uuid.New()
	n := &domain.Notification{ID: nID, SenderID: &senderID}
	repo := &mockRepo{found: n}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)

	strangerCtx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New()})
	if _, err := svc.Report(strangerCtx, nID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("stranger report err = %v, want ErrForbidden", err)
	}

	senderCtx := domain.WithCaller(context.Background(), domain.Caller{UserID: senderID})
	if _, err := svc.Report(senderCtx, nID); err != nil {
		t.Fatalf("sender report err = %v, want nil", err)
	}
}

func actionURLPtr(s string) *string { return &s }

func TestValidateActionURL(t *testing.T) {
	cases := []struct {
		name    string
		raw     *string
		wantErr bool
	}{
		{"nil", nil, false},
		{"empty", actionURLPtr(""), false},
		{"https", actionURLPtr("https://example.com/x"), false},
		{"http", actionURLPtr("http://example.com/x"), false},
		{"relative", actionURLPtr("/org/quizzes/123"), false},
		{"javascript", actionURLPtr("javascript:alert(1)"), true},
		{"data", actionURLPtr("data:text/html,<script>alert(1)</script>"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateActionURL(tc.raw)
			if tc.wantErr && err == nil {
				t.Fatalf("validateActionURL(%s) = nil, want error", tc.name)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateActionURL(%s) = %v, want nil", tc.name, err)
			}
		})
	}
}

func TestSendRejectsJavascriptActionURL(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	d := dto(domain.NotificationAudienceDTO{Type: "all"})
	d.ActionURL = actionURLPtr("javascript:alert(1)")
	_, err := svc.Send(adminCtx(), d)
	if err == nil {
		t.Fatalf("Send with javascript action_url err = nil, want validation error")
	}
	var verr *domain.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("err = %v, want *domain.ValidationError", err)
	}
	if repo.created != nil {
		t.Fatalf("notification was created despite invalid action_url")
	}
}

func TestSendAcceptsRelativeActionURL(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	d := dto(domain.NotificationAudienceDTO{Type: "all"})
	d.ActionURL = actionURLPtr("/org/quizzes/123")
	n, err := svc.Send(adminCtx(), d)
	if err != nil {
		t.Fatalf("Send with relative action_url err = %v, want nil", err)
	}
	if n.ActionURL == nil || *n.ActionURL != "/org/quizzes/123" {
		t.Fatalf("ActionURL = %v, want /org/quizzes/123", n.ActionURL)
	}
}

func TestSendSystemRejectsDataActionURL(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, &mockClassRepo{}, nil, nil, nil, nil, Senders{}, 10, nil, nil)
	err := svc.SendSystem(context.Background(), domain.SystemNotificationInput{
		Title:     "t",
		Body:      "b",
		ActionURL: actionURLPtr("data:text/html,<script>alert(1)</script>"),
		Audience:  domain.NotificationAudience{Type: domain.AudienceUsers, UserIDs: []uuid.UUID{uuid.New()}},
	})
	if err == nil {
		t.Fatalf("SendSystem with data action_url err = nil, want validation error")
	}
	if repo.created != nil {
		t.Fatalf("notification was created despite invalid action_url")
	}
}
