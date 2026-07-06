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

type mockSMS struct{ otpPhone, otpCode string }

func (m *mockSMS) SendBulk(context.Context, []string, string) error { return nil }
func (m *mockSMS) SendOTP(_ context.Context, phone, code string) error {
	m.otpPhone, m.otpCode = phone, code
	return nil
}

func testService(t *testing.T, repo domain.UserConnectorRepository, sms domain.SMSSender) (domain.ConnectorService, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := NewService(repo, rdb, sms, BotLinkConfig{TelegramBotUsername: "zoora_bot", BaleBotUsername: "zoora_bale_bot"}, nil)
	return svc, rdb
}

func callerCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: userID})
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

	if err := svc.CompleteLink(context.Background(), domain.ConnectorTelegram, resp.Token, "424242"); err != nil {
		t.Fatalf("CompleteLink: %v", err)
	}
	if repo.created == nil || repo.created.UserID != userID || repo.created.Target != "424242" {
		t.Fatalf("created = %+v", repo.created)
	}
	if repo.created.VerifiedAt == nil || !repo.created.Enabled {
		t.Fatal("bot link must be verified and enabled")
	}

	// Token is one-time.
	if err := svc.CompleteLink(context.Background(), domain.ConnectorTelegram, resp.Token, "424242"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("reuse err = %v, want ErrNotFound", err)
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
