package middleware

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client, name string, limit redis_rate.Limit, disabled bool) gin.HandlerFunc {
	// Load-testing escape hatch; the caller is responsible for never enabling
	// this in production (main.go forces it off there).
	if disabled {
		return func(c *gin.Context) { c.Next() }
	}

	limiter := redis_rate.NewLimiter(rdb)

	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", name, c.ClientIP())

		res, err := limiter.Allow(c.Request.Context(), key, limit)
		if err != nil {
			c.Next()
			return
		}

		resetSecs := int(math.Ceil(res.ResetAfter.Seconds()))
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit.Rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		c.Header("X-RateLimit-Reset", strconv.Itoa(resetSecs))

		if res.Allowed == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "too many requests",
				},
			})
			return
		}

		c.Next()
	}
}

func GlobalRateLimit(rdb *redis.Client, disabled bool) gin.HandlerFunc {
	return RateLimit(rdb, "global", redis_rate.Limit{
		Rate:   100,
		Burst:  150,
		Period: time.Second,
	}, disabled)
}

func AuthRateLimit(rdb *redis.Client, disabled bool) gin.HandlerFunc {
	return RateLimit(rdb, "auth", redis_rate.Limit{
		Rate:   5,
		Burst:  2,
		Period: time.Minute,
	}, disabled)
}

func UploadRateLimit(rdb *redis.Client, disabled bool) gin.HandlerFunc {
	return RateLimit(rdb, "upload", redis_rate.Limit{
		Rate:   3,
		Burst:  3,
		Period: time.Minute,
	}, disabled)
}

// LeadRateLimit bounds the public lead-submit endpoint: a handful per IP per
// hour is plenty for a genuine "get started" form and starves bots.
func LeadRateLimit(rdb *redis.Client, disabled bool) gin.HandlerFunc {
	return RateLimit(rdb, "leads", redis_rate.Limit{
		Rate:   5,
		Burst:  5,
		Period: time.Hour,
	}, disabled)
}
