package chathub

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/auth"
)

// presenceUpdateEvent is the WS/Redis envelope type broadcast when a user's
// presence changes (joins a room, or their last socket disconnects).
const presenceUpdateEvent = "presence_update"

// presencePayload builds the presence_update event body fanned out to a room.
func presencePayload(userID uuid.UUID, online bool) map[string]any {
	return map[string]any{
		"user_id":   userID.String(),
		"online":    online,
		"last_seen": time.Now().UTC().Format(time.RFC3339),
	}
}

// originChecker builds the WS upgrade Origin allow-list check (the CSWSH
// guard: without it any website a logged-in user visits could open an
// authenticated socket). Semantics mirror the HTTP CORS middleware config
// (middleware.CORS with AllowWildcard): "*" (the dev default) allows every
// origin; a single "*" inside an entry is a wildcard matched by prefix+suffix
// (e.g. "https://*.zoora.ir" — the multi-tenant subdomain case, where each org
// is served from its own <org>.zoora.ir host); every other entry must match
// exactly. Requests with NO Origin header are allowed — they come from
// non-browser clients, which cannot be CSWSH'd (the attack requires a victim
// browser attaching credentials), and the ?token= JWT is still required either
// way.
func originChecker(allowedOrigins []string) func(*http.Request) bool {
	allowAll := false
	exact := make(map[string]bool, len(allowedOrigins))
	type wildcard struct{ prefix, suffix string }
	var wildcards []wildcard
	for _, o := range allowedOrigins {
		o = strings.TrimRight(strings.ToLower(strings.TrimSpace(o)), "/")
		if o == "" {
			continue
		}
		if o == "*" {
			allowAll = true
			continue
		}
		// Single-"*" wildcard, mirroring gin-contrib/cors AllowWildcard: split
		// into fixed prefix/suffix and match by both ends. Extra "*"s beyond the
		// first are treated as literal, which cannot match a real Origin — so a
		// malformed pattern fails closed rather than widening the allow-list.
		if i := strings.IndexByte(o, '*'); i >= 0 {
			wildcards = append(wildcards, wildcard{prefix: o[:i], suffix: o[i+1:]})
			continue
		}
		exact[o] = true
	}
	return func(r *http.Request) bool {
		if allowAll {
			return true
		}
		origin := strings.TrimRight(strings.ToLower(r.Header.Get("Origin")), "/")
		if origin == "" {
			return true
		}
		if exact[origin] {
			return true
		}
		for _, w := range wildcards {
			if len(origin) >= len(w.prefix)+len(w.suffix) &&
				strings.HasPrefix(origin, w.prefix) &&
				strings.HasSuffix(origin, w.suffix) {
				return true
			}
		}
		return false
	}
}

// HandleWS upgrades the connection, authenticates via ?token=, and serves it.
// allowedOrigins is the browser Origin allow-list (share the CORS config).
// presence tracks per-user online state: the socket lifecycle marks the user
// online on connect and every heartbeat, offline when their last socket drops,
// and fans out presence_update events to the rooms the socket had joined.
// authRedis is the CACHE-role Redis client (the same one auth.Middleware and the
// auth service use). The token-revocation key lives there, NOT on the pub/sub
// client the Bridge uses — with scale-out those are different instances, so the
// revoke check must read authRedis or it silently fails open/closed.
func HandleWS(hub *Hub, bridge *Bridge, presence *Presence, jwt *auth.JWTService, authRedis *redis.Client, allowedOrigins []string, logger *slog.Logger) gin.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     originChecker(allowedOrigins),
	}
	hooks := presenceHooks{
		onConnect: func(userID uuid.UUID) {
			if _, err := presence.Connect(context.Background(), userID); err != nil {
				logger.Warn("chathub presence Connect failed", "user_id", userID, "error", err)
			}
		},
		onHeartbeat: func(userID uuid.UUID) {
			if err := presence.Refresh(context.Background(), userID); err != nil {
				logger.Warn("chathub presence Refresh failed", "user_id", userID, "error", err)
			}
		},
		onJoin: func(userID, convID uuid.UUID) {
			bridge.ToConversation(context.Background(), convID, presenceUpdateEvent, presencePayload(userID, true))
		},
		onDisconnect: func(userID uuid.UUID, rooms []uuid.UUID) {
			offline, err := presence.Disconnect(context.Background(), userID)
			if err != nil {
				logger.Warn("chathub presence Disconnect failed", "user_id", userID, "error", err)
				return
			}
			// Only broadcast offline once the user's LAST socket across all
			// instances is gone — otherwise a user still connected elsewhere
			// would be shown offline (the multi-instance bug this fixes).
			if !offline {
				return
			}
			for _, convID := range rooms {
				bridge.ToConversation(context.Background(), convID, presenceUpdateEvent, presencePayload(userID, false))
			}
		},
	}
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// auth.Middleware rejects revoked/logged-out tokens by checking Redis
		// (see internal/auth/middleware.go's isRevoked); a bare ValidateToken
		// does not. Replicate that check here so a logged-out token can't open
		// a live socket. Use authRedis (cache role) — the revoke key is written
		// there, not on the Bridge's pub/sub client.
		if isTokenRevoked(c.Request.Context(), authRedis, claims) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("chathub upgrade failed", "error", err)
			return
		}
		// Detach from the request ctx so the socket outlives the HTTP handler.
		hub.serve(context.Background(), ws, claims.UserID, bridge.PublishTyping, hooks)
	}
}

// isTokenRevoked mirrors auth.Middleware's isRevoked check (internal/auth/middleware.go):
// a token is revoked if the org/session-wide revocation timestamp stored at
// auth.RevokedKey(userID) is newer than the token's IssuedAt. rdb == nil fails
// closed (treated as revoked) since this is a security-critical gate with no
// caller-facing fallback path like Middleware's `rdb != nil &&` guard.
func isTokenRevoked(ctx context.Context, rdb *redis.Client, claims *auth.Claims) bool {
	if rdb == nil {
		return true
	}
	val, err := rdb.Get(ctx, auth.RevokedKey(claims.UserID.String())).Result()
	if err != nil {
		return !errors.Is(err, redis.Nil)
	}
	revokedAt, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return true
	}
	if claims.IssuedAt == nil {
		return true
	}
	return claims.IssuedAt.Unix() < revokedAt
}
