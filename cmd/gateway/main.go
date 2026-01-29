package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/handlers"
	"github.com/ludovicopassari/api-gateway/internal/middlewares"
	"github.com/ludovicopassari/api-gateway/internal/monitoring"
	"github.com/ludovicopassari/api-gateway/internal/ratelimit"
	"github.com/ludovicopassari/api-gateway/internal/storage"

	"github.com/ludovicopassari/api-gateway/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

/*

	Logging middleware (start)
	Rate limiter middleware
		Prometheus middleware
		Handler
		Prometheus
	Rate limiter
	Logging middleware (end)

*/

func main() {

	logger.SetupLogger()
	defer logger.Sync()

	logger.Info("Starting Gateway",
		zap.Time("timestamp", time.Now()),
	)

	reg := prometheus.NewRegistry()
	m := monitoring.NewMetrics(reg)

	storage, _ := storage.NewStorageFromEnv()
	limiter := ratelimit.NewSlidingWindow(storage, 100, time.Second*60)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.RateLimitMiddleware(limiter))
	r.Use(middlewares.PrometheusMonitoringMiddleware(m))

	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})))
	r.GET("/health", handlers.HealthHandler)

	r.Run(":" + "80")

}
