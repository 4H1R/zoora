package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
)

// RequestInfo stashes the client IP and user-agent into the request context so
// the audit recorder can attach them to entries. Relies on trusted-proxy config
// (set in main) for c.ClientIP() to resolve X-Forwarded-For correctly.
func RequestInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ri := domain.RequestInfo{
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}
		c.Request = c.Request.WithContext(domain.WithRequestInfo(c.Request.Context(), ri))
		c.Next()
	}
}
