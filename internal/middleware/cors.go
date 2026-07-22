package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	// A credentialed wildcard is an insecure browser-trust misconfiguration
	// (and browsers reject it anyway), so drop credentials whenever the
	// allow-list contains a bare "*". An explicit list — including the
	// "https://*.zoora.ir" subdomain wildcard — keeps credentials enabled.
	allowCredentials := true
	for _, o := range allowedOrigins {
		if o == "*" {
			allowCredentials = false
			break
		}
	}
	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		// Multi-tenant: each org is served from its own <org>.zoora.ir host, so
		// the allow-list uses a "https://*.zoora.ir" wildcard. Without this the
		// single "*" would be rejected as a bad origin at startup.
		AllowWildcard:    true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: allowCredentials,
		MaxAge:           12 * time.Hour,
	})
}
