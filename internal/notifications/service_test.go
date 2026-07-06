package notifications

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

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
	svc := NewService(repo, &mockClassRepo{}, nil, 10, nil)
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
	svc := NewService(&mockRepo{}, &mockClassRepo{}, nil, 10, nil)
	_, err := svc.Send(managerCtx(orgID), dto(domain.NotificationAudienceDTO{Type: "all"}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSendManagerOrgForcedToOwnOrg(t *testing.T) {
	orgID, otherOrg := uuid.New(), uuid.New()
	repo := &mockRepo{}
	svc := NewService(repo, &mockClassRepo{}, nil, 10, nil)
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
	svc := NewService(&mockRepo{}, classRepo, nil, 10, nil)
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
	svc := NewService(repo, classRepo, nil, 10, nil)
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
	svc := NewService(&mockRepo{}, classRepo, nil, 10, nil)
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
	svc := NewService(&mockRepo{usersInClasses: 1}, &mockClassRepo{}, nil, 10, nil)
	_, err := svc.Send(ctx, dto(domain.NotificationAudienceDTO{Type: "users", UserIDs: ids}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSendManagerRoleOutsideOrgForbidden(t *testing.T) {
	orgID := uuid.New()
	roleID := uuid.New()
	svc := NewService(&mockRepo{roleInScope: false}, &mockClassRepo{}, nil, 10, nil)
	_, err := svc.Send(managerCtx(orgID), dto(domain.NotificationAudienceDTO{Type: "role", RoleID: &roleID}))
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestSendRateLimited(t *testing.T) {
	orgID := uuid.New()
	repo := &mockRepo{senderCount: 10}
	svc := NewService(repo, &mockClassRepo{}, nil, 10, nil)
	_, err := svc.Send(managerCtx(orgID), dto(domain.NotificationAudienceDTO{Type: "org"}))
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}

func TestSendWithoutPermissionForbidden(t *testing.T) {
	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgID})
	svc := NewService(&mockRepo{}, &mockClassRepo{}, nil, 10, nil)
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
	svc := NewService(repo, &mockClassRepo{}, nil, 10, nil)

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
	svc := NewService(repo, &mockClassRepo{}, nil, 10, nil)
	st, err := svc.Status(ctxWithCaller())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.UnreadCount != 7 {
		t.Fatalf("UnreadCount = %d, want 7", st.UnreadCount)
	}
}

func TestStatusRequiresCaller(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockClassRepo{}, nil, 10, nil)
	if _, err := svc.Status(context.Background()); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}
