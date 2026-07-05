package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// HeaderRequestID is the response/request header carrying the correlation id.
const HeaderRequestID = "X-Request-Id"

// RequestID assigns a correlation id to every request. It reuses an inbound
// X-Request-Id (set by an upstream proxy or another service) when present,
// otherwise generates a UUIDv7. The id is stored on the request context (so the
// logger tags every line for this request) and echoed back in the response
// header so clients and logs can be tied to the same request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			if v, err := uuid.NewV7(); err == nil {
				id = v.String()
			} else {
				id = uuid.NewString()
			}
		}

		c.Request = c.Request.WithContext(domain.WithRequestID(c.Request.Context(), id))
		c.Writer.Header().Set(HeaderRequestID, id)

		c.Next()
	}
}
