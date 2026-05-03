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

func (s *service) Login(ctx context.Context, dto domain.LoginDTO) (*domain.User, string, error) {
	lockKey := fmt.Sprintf("login:lock:%s", dto.Username)
	if locked, _ := s.redis.Get(ctx, lockKey).Int(); locked >= loginMaxFails {
		return nil, "", domain.ErrUnauthorized
	}

	user, err := s.userRepo.FindByUsername(ctx, dto.Username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrUnauthorized
		}
		return nil, "", fmt.Errorf("auth.Login lookup: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		s.bumpLoginFail(ctx, lockKey)
		return nil, "", domain.ErrUnauthorized
	}

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("auth.Login generate token: %w", err)
	}

	s.redis.Del(ctx, lockKey)
	s.logger.Info("user logged in", "user_id", user.ID.String())
	return user, token, nil
}

func (s *service) bumpLoginFail(ctx context.Context, key string) {
	pipe := s.redis.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, loginLockTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Warn("login fail counter", "error", err)
		return
	}
	if incr.Val() >= loginMaxFails {
		s.logger.Warn("login lockout triggered", "key", key)
	}
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
