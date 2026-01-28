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

		c.Next()

		start := c.GetTime("start_time")
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

	}

}
