package connectors

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
)

type mockConnRepo struct {
	domain.UserConnectorRepository
	created *domain.UserConnector
	byID    *domain.UserConnector
}

func (m *mockConnRepo) Create(_ context.Context, c *domain.UserConnector) error {
	m.created = c
	return nil
}

func (m *mockConnRepo) FindByID(context.Context, uuid.UUID) (*domain.UserConnector, error) {
	if m.byID == nil {
		return nil, domain.ErrNotFound
	}
	return m.byID, nil
}

type mockUserRepo struct {
	domain.UserRepository
	user *domain.User
}

func (m *mockUserRepo) FindByID(context.Context, uuid.UUID) (*domain.User, error) {
	if m.user == nil {
		return nil, domain.ErrNotFound
	}
	return m.user, nil
}

type mockOrgRepo struct {
	domain.OrganizationRepository
	org *domain.Organization
}

func (m *mockOrgRepo) FindByID(context.Context, uuid.UUID) (*domain.Organization, error) {
	if m.org == nil {
		return nil, domain.ErrNotFound
	}
	return m.org, nil
}

type mockSMS struct{ otpPhone, otpCode string }

func (m *mockSMS) SendBulk(context.Context, []string, string) error { return nil }
func (m *mockSMS) SendOTP(_ context.Context, phone, code string) error {
	m.otpPhone, m.otpCode = phone, code
	return nil
}

// fakeTransactor runs fn inline with no real DB — unit tests exercise the audit
// same-tx wiring without a database.
type fakeTransactor struct{}

func (fakeTransactor) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// auditSpy captures the records a service emits so tests can assert on them.
type auditSpy struct{ records []domain.AuditRecord }

func (a *auditSpy) Record(_ context.Context, r domain.AuditRecord) error {
	a.records = append(a.records, r)
	return nil
}

func (a *auditSpy) RecordDenied(_ context.Context, _ domain.AuditRecord) error { return nil }

func testService(t *testing.T, repo domain.UserConnectorRepository, sms domain.SMSSender) (domain.ConnectorService, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := NewService(repo, nil, nil, rdb, sms, BotLinkConfig{TelegramBotUsername: "zoora_bot", BaleBotUsername: "zoora_bale_bot"}, fakeTransactor{}, &auditSpy{}, nil)
	return svc, rdb
}

func callerCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: userID,
		Ent:    domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
	})
}

func TestCreateLinkTokenAndCompleteLink(t *testing.T) {
	repo := &mockConnRepo{}
	svc, _ := testService(t, repo, &mockSMS{})
	userID := uuid.New()

	resp, err := svc.CreateLinkToken(callerCtx(userID), domain.ConnectorTelegram)
	if err != nil {
		t.Fatalf("CreateLinkToken: %v", err)
	}
	if resp.DeepLink != "https://t.me/zoora_bot?start="+resp.Token {
		t.Fatalf("deep link = %s", resp.DeepLink)
	}

	if _, err := svc.CompleteLink(context.Background(), domain.ConnectorTelegram, resp.Token, "424242"); err != nil {
		t.Fatalf("CompleteLink: %v", err)
	}
	if repo.created == nil || repo.created.UserID != userID || repo.created.Target != "424242" {
		t.Fatalf("created = %+v", repo.created)
	}
	if repo.created.VerifiedAt == nil || !repo.created.Enabled {
		t.Fatal("bot link must be verified and enabled")
	}

	// Token is one-time.
	if _, err := svc.CompleteLink(context.Background(), domain.ConnectorTelegram, resp.Token, "424242"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("reuse err = %v, want ErrNotFound", err)
	}
}

func TestCompleteLinkReturnsAccountGreeting(t *testing.T) {
	repo := &mockConnRepo{}
	orgID := uuid.New()
	userID := uuid.New()
	userRepo := &mockUserRepo{user: &domain.User{ID: userID, Username: "ali", Name: "Ali A", OrganizationID: &orgID}}
	orgRepo := &mockOrgRepo{org: &domain.Organization{ID: orgID, Name: "Acme"}}
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	audit := &auditSpy{}
	svc := NewService(repo, userRepo, orgRepo, rdb, &mockSMS{}, BotLinkConfig{TelegramBotUsername: "zoora_bot"}, fakeTransactor{}, audit, nil)

	resp, err := svc.CreateLinkToken(callerCtx(userID), domain.ConnectorTelegram)
	if err != nil {
		t.Fatalf("CreateLinkToken: %v", err)
	}
	res, err := svc.CompleteLink(context.Background(), domain.ConnectorTelegram, resp.Token, "424242")
	if err != nil {
		t.Fatalf("CompleteLink: %v", err)
	}
	if res.Username != "ali" || res.Name != "Ali A" || res.OrgName != "Acme" {
		t.Fatalf("result = %+v, want ali/Ali A/Acme", res)
	}

	// The link is audited under the user's org, labelled by provider.
	if len(audit.records) != 1 {
		t.Fatalf("audit records = %d, want 1", len(audit.records))
	}
	rec := audit.records[0]
	if rec.Action != domain.AuditCreated || rec.TargetType != domain.AuditTargetConnector {
		t.Fatalf("audit action/type = %s/%s", rec.Action, rec.TargetType)
	}
	if rec.TargetLabel != string(domain.ConnectorTelegram) {
		t.Fatalf("audit label = %q, want telegram", rec.TargetLabel)
	}
	if rec.OrgID == nil || *rec.OrgID != orgID {
		t.Fatalf("audit org = %v, want %v", rec.OrgID, orgID)
	}
}

func TestConnectedMessage(t *testing.T) {
	tests := []struct {
		name string
		res  *domain.ConnectorLinkResult
		want string
	}{
		{"nil falls back", nil, "✅ Connected! You will now receive Zoora notifications here."},
		{"no username falls back", &domain.ConnectorLinkResult{}, "✅ Connected! You will now receive Zoora notifications here."},
		{
			"username + name + org", &domain.ConnectorLinkResult{Username: "ali", Name: "Ali A", OrgName: "Acme"},
			"✅ Connected as @ali (Ali A) · Acme.\nYou will now receive Zoora notifications for this account here.",
		},
		{
			"username only", &domain.ConnectorLinkResult{Username: "ali"},
			"✅ Connected as @ali.\nYou will now receive Zoora notifications for this account here.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := connectedMessage(tt.res); got != tt.want {
				t.Fatalf("connectedMessage = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSMSOTPRoundTrip(t *testing.T) {
	repo := &mockConnRepo{}
	sms := &mockSMS{}
	svc, _ := testService(t, repo, sms)
	userID := uuid.New()
	ctx := callerCtx(userID)

	if err := svc.RequestSMSOTP(ctx, domain.RequestSMSOTPDTO{Phone: "09120000001"}); err != nil {
		t.Fatalf("RequestSMSOTP: %v", err)
	}
	if sms.otpPhone != "09120000001" || len(sms.otpCode) != 6 {
		t.Fatalf("otp sent to %s code %s", sms.otpPhone, sms.otpCode)
	}

	if err := svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: sms.otpCode}); err != nil {
		t.Fatalf("VerifySMSOTP: %v", err)
	}
	if repo.created == nil || repo.created.Type != domain.ConnectorSMS || repo.created.Target != "09120000001" {
		t.Fatalf("created = %+v", repo.created)
	}
}

func TestVerifySMSOTPWrongCode(t *testing.T) {
	svc, _ := testService(t, &mockConnRepo{}, &mockSMS{})
	ctx := callerCtx(uuid.New())
	_ = svc.RequestSMSOTP(ctx, domain.RequestSMSOTPDTO{Phone: "09120000001"})
	if err := svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: "000000"}); err == nil {
		t.Fatal("expected error for wrong code")
	}
}

func TestVerifySMSOTPAttemptCapInvalidatesCode(t *testing.T) {
	repo := &mockConnRepo{}
	sms := &mockSMS{}
	svc, _ := testService(t, repo, sms)
	ctx := callerCtx(uuid.New())

	if err := svc.RequestSMSOTP(ctx, domain.RequestSMSOTPDTO{Phone: "09120000001"}); err != nil {
		t.Fatalf("RequestSMSOTP: %v", err)
	}
	// Exhaust the attempt budget with wrong guesses. Use a code that can never
	// collide with the real 6-digit numeric code.
	for i := range otpMaxAttempts {
		if err := svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: "wrong"}); err == nil {
			t.Fatalf("wrong guess %d: expected error", i)
		}
	}
	// The code is now burned: even the correct code returns "no pending verification".
	err := svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: sms.otpCode})
	var verr *domain.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("err = %v, want validation error", err)
	}
	if verr.Fields["code"] != "no pending verification — request a new code" {
		t.Fatalf("field = %q, want no-pending message", verr.Fields["code"])
	}
	if repo.created != nil {
		t.Fatalf("no connector should be created after cap, got %+v", repo.created)
	}
}

func TestVerifySMSOTPSuccessClearsAttemptCounter(t *testing.T) {
	repo := &mockConnRepo{}
	sms := &mockSMS{}
	svc, rdb := testService(t, repo, sms)
	userID := uuid.New()
	ctx := callerCtx(userID)

	if err := svc.RequestSMSOTP(ctx, domain.RequestSMSOTPDTO{Phone: "09120000001"}); err != nil {
		t.Fatalf("RequestSMSOTP: %v", err)
	}
	// A wrong guess within budget, then the correct code succeeds.
	if err := svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: "wrong"}); err == nil {
		t.Fatal("expected error for wrong code")
	}
	if err := svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: sms.otpCode}); err != nil {
		t.Fatalf("VerifySMSOTP: %v", err)
	}
	if repo.created == nil || repo.created.Type != domain.ConnectorSMS {
		t.Fatalf("created = %+v", repo.created)
	}
	// Both keys are cleared after success.
	if n, _ := rdb.Exists(ctx, otpKey(userID)).Result(); n != 0 {
		t.Fatal("otp key must be deleted on success")
	}
	if n, _ := rdb.Exists(ctx, otpAttemptsKey(userID)).Result(); n != 0 {
		t.Fatal("attempt counter must be deleted on success")
	}
}

func TestRequestSMSOTPResetsAttemptCounter(t *testing.T) {
	repo := &mockConnRepo{}
	sms := &mockSMS{}
	svc, _ := testService(t, repo, sms)
	ctx := callerCtx(uuid.New())

	if err := svc.RequestSMSOTP(ctx, domain.RequestSMSOTPDTO{Phone: "09120000001"}); err != nil {
		t.Fatalf("RequestSMSOTP: %v", err)
	}
	// Burn most of the budget without hitting the cap.
	for range otpMaxAttempts - 1 {
		_ = svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: "wrong"})
	}
	// A fresh request resets the counter, so the user gets a full budget again.
	if err := svc.RequestSMSOTP(ctx, domain.RequestSMSOTPDTO{Phone: "09120000001"}); err != nil {
		t.Fatalf("re-RequestSMSOTP: %v", err)
	}
	// otpMaxAttempts-1 more wrong guesses must NOT burn the code (counter was reset).
	for range otpMaxAttempts - 1 {
		_ = svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: "wrong"})
	}
	if err := svc.VerifySMSOTP(ctx, domain.VerifySMSOTPDTO{Code: sms.otpCode}); err != nil {
		t.Fatalf("correct code after reset should succeed, got %v", err)
	}
}

func TestUpdateForeignConnectorForbidden(t *testing.T) {
	other := &domain.UserConnector{ID: uuid.New(), UserID: uuid.New()}
	repo := &mockConnRepo{byID: other}
	svc, _ := testService(t, repo, &mockSMS{})
	enabled := false
	_, err := svc.Update(callerCtx(uuid.New()), other.ID, domain.UpdateConnectorDTO{Enabled: &enabled})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}
