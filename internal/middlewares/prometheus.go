package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/monitoring"
)

func PrometheusMonitoringMiddleware(m *monitoring.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.FullPath() == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		path := c.FullPath()

		if path == "" {
			path = "undefined"
		}

		m.RequestsCount.WithLabelValues(
			c.Request.Method,
			path,
			fmt.Sprint(status),
		).Inc()

		m.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)

		fmt.Printf("Request %s %s completed with status %d in %v\n",
			c.Request.Method, c.FullPath(), status, duration)
	}

}
