package websocket

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/4H1R/zoora/internal/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Override in production with proper origin checking
	},
}

func HandleWebSocket(hub *Hub, jwt *auth.JWTService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authenticate via query param token
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

		room := c.Param("room")
		if room == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing room"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("websocket upgrade failed", "error", err)
			return
		}

		client := NewClient(hub, conn, claims.UserID.String(), room, logger)
		hub.register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
