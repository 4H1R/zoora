package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
)

const (
	loginLockTTL  = 15 * time.Minute
	loginMaxFails = 5

	revokedKeyPrefix = "auth:revoked:"
	revokedTTL       = 7 * 24 * time.Hour
)

func RevokedKey(userID string) string { return revokedKeyPrefix + userID }

type service struct {
	userRepo   domain.UserRepository
	jwtService *JWTService
	redis      *redis.Client
	logger     *slog.Logger
}

func NewAuthService(
	userRepo domain.UserRepository,
	jwtService *JWTService,
	rdb *redis.Client,
	logger *slog.Logger,
) domain.AuthService {
	return &service{
		userRepo:   userRepo,
		jwtService: jwtService,
		redis:      rdb,
		logger:     logger,
	}
}

func (s *service) Login(ctx context.Context, dto domain.LoginDTO, orgID *uuid.UUID) (*domain.User, string, error) {
	scope := "admin"
	if orgID != nil {
		scope = orgID.String()
	}
	lockKey := fmt.Sprintf("login:lock:%s:%s", scope, dto.Username)
	if locked, _ := s.redis.Get(ctx, lockKey).Int(); locked >= loginMaxFails {
		return nil, "", domain.ErrAccountLocked
	}

	var user *domain.User
	var err error
	if orgID != nil {
		user, err = s.userRepo.FindByUsernameAndOrg(ctx, dto.Username, *orgID)
	} else {
		user, err = s.userRepo.FindAdminByUsername(ctx, dto.Username)
	}
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrUnauthorized
		}
		return nil, "", fmt.Errorf("auth.Login lookup: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		if s.bumpLoginFail(ctx, lockKey) {
			return nil, "", domain.ErrAccountLocked
		}
		return nil, "", domain.ErrUnauthorized
	}

	if user.DisabledAt != nil {
		s.logger.Warn("disabled user login blocked", "user_id", user.ID.String())
		return nil, "", domain.ErrUserDisabled
	}

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("auth.Login generate token: %w", err)
	}

	s.redis.Del(ctx, lockKey)
	s.logger.Info("user logged in", "user_id", user.ID.String(), "scope", scope)
	return user, token, nil
}

// bumpLoginFail increments the failure counter and reports whether the account
// is now locked (this attempt reached the threshold).
func (s *service) bumpLoginFail(ctx context.Context, key string) bool {
	pipe := s.redis.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, loginLockTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Warn("login fail counter", "error", err)
		return false
	}
	if incr.Val() >= loginMaxFails {
		s.logger.Warn("login lockout triggered", "key", key)
		return true
	}
	return false
}

// SessionManager issues tokens and revokes a user's own sessions. It performs
// no caller authorization — unlike AdminRevokeSessions — because callers (e.g.
// self password change) have already proven identity. Satisfies
// domain.SessionTokenService.
type SessionManager struct {
	jwt   *JWTService
	redis *redis.Client
}

func NewSessionManager(jwt *JWTService, rdb *redis.Client) *SessionManager {
	return &SessionManager{jwt: jwt, redis: rdb}
}

func (m *SessionManager) GenerateToken(userID uuid.UUID) (string, error) {
	return m.jwt.GenerateToken(userID)
}

// RevokeUserSessions invalidates every token issued before now for userID by
// stamping the revocation marker the auth middleware checks (isRevoked). The
// caller should mint a replacement token afterwards so its own device — whose
// old token is now invalid — stays signed in.
func (m *SessionManager) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().Unix()
	if err := m.redis.Set(ctx, RevokedKey(userID.String()), now, revokedTTL).Err(); err != nil {
		return fmt.Errorf("auth.RevokeUserSessions: %w", err)
	}
	return nil
}

func (s *service) AdminRevokeSessions(ctx context.Context, userID uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.ErrForbidden
	}
	now := time.Now().Unix()
	key := RevokedKey(userID.String())
	if err := s.redis.Set(ctx, key, now, revokedTTL).Err(); err != nil {
		return fmt.Errorf("auth.AdminRevokeSessions: %w", err)
	}
	s.logger.Warn("admin revoked sessions",
		"target_user_id", userID.String(),
		"revoked_by", caller.UserID.String(),
	)
	return nil
}
