package health

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/platform/storage"
)

type Checker struct {
	db      *gorm.DB
	redis   *redis.Client
	storage *storage.Client
}

func NewChecker(db *gorm.DB, redis *redis.Client, storage *storage.Client) *Checker {
	return &Checker{db: db, redis: redis, storage: storage}
}

func (c *Checker) Readiness(ctx context.Context) map[string]error {
	checks := map[string]error{
		"database": c.db.WithContext(ctx).Exec("SELECT 1").Error,
		"redis":    c.redis.Ping(ctx).Err(),
		"s3":       c.storage.HeadBucket(ctx),
	}
	return checks
}

func (c *Checker) LivenessHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Checker) ReadinessHandler(ctx *gin.Context) {
	checks := c.Readiness(ctx.Request.Context())
	status := http.StatusOK
	result := make(map[string]string, len(checks))

	for name, err := range checks {
		if err != nil {
			status = http.StatusServiceUnavailable
			result[name] = err.Error()
		} else {
			result[name] = "ok"
		}
	}

	ctx.JSON(status, gin.H{
		"status": statusString(status),
		"checks": result,
	})
}

func statusString(code int) string {
	if code == http.StatusOK {
		return "ok"
	}
	return "unavailable"
}
