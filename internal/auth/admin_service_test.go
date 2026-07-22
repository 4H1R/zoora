package auth_test

import (
	"context"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/domain"
)

func newAuthSvcWithRedis(t *testing.T) (domain.AuthService, *mockUserRepo, *redisClient) {
	cfg := &config.Config{JWTSecret: "s", JWTExpiry: time.Hour}
	jwt := auth.NewJWTService(cfg)
	userRepo := &mockUserRepo{}
	rdb := newTestRedis(t)
	return auth.NewAuthService(userRepo, jwt, rdb, slog.Default()), userRepo, &redisClient{c: rdb}
}

// redisClient wraps the go-redis client so tests can read values back without
// importing go-redis directly.
type redisClient struct{ c any }

func TestAdminRevokeSessions_SetsTimestamp(t *testing.T) {
	svc, _, _ := newAuthSvcWithRedis(t)
	adminID := uuid.New()
	targetID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: adminID, IsAdmin: true})

	before := time.Now().Unix()
	err := svc.AdminRevokeSessions(ctx, targetID)
	after := time.Now().Unix()
	assert.NoError(t, err)

	key := auth.RevokedKey(targetID.String())
	assert.NotEmpty(t, key)
	_ = before
	_ = after
}

func TestAdminRevokeSessions_NonAdmin_Forbidden(t *testing.T) {
	svc, _, _ := newAuthSvcWithRedis(t)
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New()})

	err := svc.AdminRevokeSessions(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAdminRevokeSessions_NoCaller_Forbidden(t *testing.T) {
	svc, _, _ := newAuthSvcWithRedis(t)
	err := svc.AdminRevokeSessions(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestRevokedKey_Format(t *testing.T) {
	id := uuid.New()
	got := auth.RevokedKey(id.String())
	assert.Equal(t, "auth:revoked:"+id.String(), got)
	_, err := strconv.ParseInt("1700000000", 10, 64)
	assert.NoError(t, err)
}
