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
		key := c.ClientIP()

		c.Set("rate_limited", false)

		if pass, _ := l.Allow(key); !pass {
			c.Set("rate_limited", true)
			// headers
			c.Header("Retry-After", "60") // TODO dinamic config
			c.Header("X-RateLimit-Limit", "100")
			c.Header("X-RateLimit-Remaining", "0") // TODO update
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(1*time.Minute).Unix()))
			// response body
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort() // this prevents gin from calling pendin middlewares
			return
		}

		c.Next()

	}
}
