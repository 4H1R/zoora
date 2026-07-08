package connectors

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
)

const (
	linkTokenTTL  = 15 * time.Minute
	otpTTL        = 5 * time.Minute
	otpMaxPerHour = 3
)

type BotLinkConfig struct {
	TelegramBotUsername string
	BaleBotUsername     string
}

type service struct {
	repo   domain.UserConnectorRepository
	rdb    *redis.Client
	sms    domain.SMSSender
	bots   BotLinkConfig
	logger *slog.Logger
}

func NewService(repo domain.UserConnectorRepository, rdb *redis.Client, smsSender domain.SMSSender, bots BotLinkConfig, logger *slog.Logger) domain.ConnectorService {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{repo: repo, rdb: rdb, sms: smsSender, bots: bots, logger: logger}
}

func linkKey(t domain.ConnectorType, token string) string {
	return fmt.Sprintf("connector:link:%s:%s", t, token)
}

func otpKey(userID uuid.UUID) string   { return "connector:otp:" + userID.String() }
func otpRLKey(userID uuid.UUID) string { return "connector:otp-rl:" + userID.String() }

func (s *service) CreateLinkToken(ctx context.Context, t domain.ConnectorType) (*domain.LinkTokenResponse, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasFeature(domain.FeatureConnectors) {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureConnectors)
	}
	var botUsername, deepLinkBase string
	switch t {
	case domain.ConnectorTelegram:
		botUsername, deepLinkBase = s.bots.TelegramBotUsername, "https://t.me/"
	case domain.ConnectorBale:
		botUsername, deepLinkBase = s.bots.BaleBotUsername, "https://ble.ir/"
	default:
		return nil, domain.NewValidationError(map[string]string{"type": "must be telegram or bale"})
	}
	if botUsername == "" {
		return nil, domain.NewValidationError(map[string]string{"type": "channel is not configured"})
	}

	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("connectors.service.CreateLinkToken: %w", err)
	}
	token := hex.EncodeToString(raw)
	if err := s.rdb.Set(ctx, linkKey(t, token), caller.UserID.String(), linkTokenTTL).Err(); err != nil {
		return nil, fmt.Errorf("connectors.service.CreateLinkToken: storing token: %w", err)
	}
	return &domain.LinkTokenResponse{
		Token:     token,
		DeepLink:  fmt.Sprintf("%s%s?start=%s", deepLinkBase, botUsername, token),
		ExpiresAt: time.Now().Add(linkTokenTTL),
	}, nil
}

// CompleteLink runs in the worker (bot poller) — no Caller in ctx; identity
// comes from the one-time token.
func (s *service) CompleteLink(ctx context.Context, t domain.ConnectorType, token, chatID string) error {
	key := linkKey(t, token)
	val, err := s.rdb.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("connectors.service.CompleteLink: %w", err)
	}
	userID, err := uuid.Parse(val)
	if err != nil {
		return fmt.Errorf("connectors.service.CompleteLink: bad stored user id: %w", err)
	}
	now := time.Now()
	return s.repo.Create(ctx, &domain.UserConnector{
		UserID:     userID,
		Type:       t,
		Target:     chatID,
		VerifiedAt: &now,
		Enabled:    true,
	})
}

type otpRecord struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func (s *service) RequestSMSOTP(ctx context.Context, dto domain.RequestSMSOTPDTO) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasFeature(domain.FeatureConnectors) {
		return domain.NewFeatureError(caller.Ent.Plan, domain.FeatureConnectors)
	}
	if s.sms == nil {
		return domain.NewValidationError(map[string]string{"phone": "SMS channel is not configured"})
	}
	n, err := s.rdb.Incr(ctx, otpRLKey(caller.UserID)).Result()
	if err != nil {
		return fmt.Errorf("connectors.service.RequestSMSOTP: rate limit: %w", err)
	}
	if n == 1 {
		s.rdb.Expire(ctx, otpRLKey(caller.UserID), time.Hour)
	}
	if n > otpMaxPerHour {
		return domain.ErrRateLimited
	}

	code, err := sixDigitCode()
	if err != nil {
		return fmt.Errorf("connectors.service.RequestSMSOTP: %w", err)
	}
	rec, _ := json.Marshal(otpRecord{Phone: dto.Phone, Code: code})
	if err := s.rdb.Set(ctx, otpKey(caller.UserID), rec, otpTTL).Err(); err != nil {
		return fmt.Errorf("connectors.service.RequestSMSOTP: storing otp: %w", err)
	}
	return s.sms.SendOTP(ctx, dto.Phone, code)
}

func sixDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *service) VerifySMSOTP(ctx context.Context, dto domain.VerifySMSOTPDTO) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	raw, err := s.rdb.Get(ctx, otpKey(caller.UserID)).Result()
	if err == redis.Nil {
		return domain.NewValidationError(map[string]string{"code": "no pending verification — request a new code"})
	}
	if err != nil {
		return fmt.Errorf("connectors.service.VerifySMSOTP: %w", err)
	}
	var rec otpRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return fmt.Errorf("connectors.service.VerifySMSOTP: decoding: %w", err)
	}
	if rec.Code != dto.Code {
		return domain.NewValidationError(map[string]string{"code": "incorrect code"})
	}
	s.rdb.Del(ctx, otpKey(caller.UserID))
	now := time.Now()
	return s.repo.Create(ctx, &domain.UserConnector{
		UserID:     caller.UserID,
		Type:       domain.ConnectorSMS,
		Target:     rec.Phone,
		VerifiedAt: &now,
		Enabled:    true,
	})
}

func (s *service) RegisterPushToken(ctx context.Context, dto domain.RegisterPushTokenDTO) (*domain.UserConnector, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	now := time.Now()
	c := &domain.UserConnector{
		UserID:     caller.UserID,
		Type:       domain.ConnectorPush,
		Target:     dto.Token,
		VerifiedAt: &now,
		Enabled:    true,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *service) List(ctx context.Context) ([]domain.UserConnector, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListByUser(ctx, caller.UserID)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateConnectorDTO) (*domain.UserConnector, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.UserID != caller.UserID {
		return nil, domain.ErrForbidden
	}
	if dto.Enabled != nil {
		c.Enabled = *dto.Enabled
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *service) Unlink(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if c.UserID != caller.UserID {
		return domain.ErrForbidden
	}
	return s.repo.Delete(ctx, id)
}
