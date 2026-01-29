package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/handlers"
	"github.com/ludovicopassari/api-gateway/internal/limiters"
	"github.com/ludovicopassari/api-gateway/internal/middlewares"
	"github.com/ludovicopassari/api-gateway/internal/monitoring"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(reg *prometheus.Registry, m *monitoring.Metrics, limiter *limiters.SlidingWindowLimiter) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.RateLimitMiddleware(limiter))
	r.Use(middlewares.PrometheusMonitoringMiddleware(m))

	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})))
	r.GET("/health", handlers.HealtHandler)

	return r
}
