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
// authenticated socket). Semantics mirror the HTTP CORS middleware config:
// "*" (the dev default) allows every origin; otherwise the Origin header must
// exactly match an allowed entry. Requests with NO Origin header are allowed —
// they come from non-browser clients, which cannot be CSWSH'd (the attack
// requires a victim browser attaching credentials), and the ?token= JWT is
// still required either way.
func originChecker(allowedOrigins []string) func(*http.Request) bool {
	allowAll := false
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		o = strings.TrimRight(strings.ToLower(strings.TrimSpace(o)), "/")
		if o == "*" {
			allowAll = true
		}
		if o != "" {
			allowed[o] = true
		}
	}
	return func(r *http.Request) bool {
		if allowAll {
			return true
		}
		origin := strings.TrimRight(strings.ToLower(r.Header.Get("Origin")), "/")
		if origin == "" {
			return true
		}
		return allowed[origin]
	}
}

// HandleWS upgrades the connection, authenticates via ?token=, and serves it.
// allowedOrigins is the browser Origin allow-list (share the CORS config).
// presence tracks per-user online state: the socket lifecycle marks the user
// online on connect and every heartbeat, offline when their last socket drops,
// and fans out presence_update events to the rooms the socket had joined.
func HandleWS(hub *Hub, bridge *Bridge, presence *Presence, jwt *auth.JWTService, allowedOrigins []string, logger *slog.Logger) gin.HandlerFunc {
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
		// a live socket. bridge.rdb is the same client auth.Middleware uses.
		if isTokenRevoked(c.Request.Context(), bridge.rdb, claims) {
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
