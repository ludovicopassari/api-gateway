package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/handlers"
	"github.com/ludovicopassari/api-gateway/internal/limiters"
	"github.com/ludovicopassari/api-gateway/internal/middlewares"
	"github.com/ludovicopassari/api-gateway/internal/monitoring"
)

func SetupRouter(m *monitoring.Metrics, limiter *limiters.SlidingWindowLimiter) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.RateLimitMiddleware(limiter))
	r.Use(middlewares.PrometheusMonitoringMiddleware(m))
	//r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(m.Registry(), promhttp.HandlerOpts{Registry: m.Registry()})))
	r.GET("/health", handlers.HealtHandler)

	return r
}
