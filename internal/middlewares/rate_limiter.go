package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/limiters"
)

func RateLimitMiddleware(l limiters.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() // oppure userID

		allowed, _ := l.Allow(key)
		if !allowed {
			c.Header("Retry-After", "60") // 60 secondi
			c.Header("X-RateLimit-Limit", "100")
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(1*time.Minute).Unix()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()

	}
}
