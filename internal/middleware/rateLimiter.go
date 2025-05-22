package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/halftoothed/urlite/internal/redis"
)

const (
	rateLimit       = 2
	rateLimitWindow = time.Minute
)

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "limit:" + ip

		val, err := redis.Rdb.Get(redis.Ctx, key).Int()

		if err != nil && err.Error() != "redis: nil" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "redis error"})
			return
		}

		if val >= rateLimit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate Limit Exceeded"})
			return
		}

		pipe := redis.Rdb.TxPipeline()
		pipe.Incr(redis.Ctx, key)
		pipe.Expire(redis.Ctx, key, rateLimitWindow)
		_, _ = pipe.Exec(redis.Ctx)

		c.Next()

	}
}
