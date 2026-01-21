package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/limiter"
)

/*
this function needs to reject requests for a specific IP if its request limit is reached
*/

func RateLimitMiddleware(l limiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() // oppure userID

		allowed, retryAfter := l.Allow(key)
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
