package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/limiters"
	"github.com/ludovicopassari/api-gateway/internal/limiters/storage"
	"github.com/ludovicopassari/api-gateway/internal/middlewares"
	"github.com/ludovicopassari/api-gateway/internal/monitoring"
	"github.com/ludovicopassari/api-gateway/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	// setup logging
	logCfg := logger.Config{
		Level:       getEnv("LOG_LEVEL", "debug"),
		Environment: getEnv("ENV", "development"),
		OutputPaths: []string{"stdout"},
	}

	if err := logger.Init(logCfg); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	reg := prometheus.NewRegistry()
	m := monitoring.NewMetrics(reg)

	// setup connection with in-memory DB
	rdb, err := storage.NewRedis("redis:6379")
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})))
	r.GET("/healt", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	api := r.Group("/api")
	limiter := limiters.NewSlidingWindowLimiter(rdb, 100, time.Minute)
	api.Use(middlewares.PrometheusMonitoringMiddleware(m))
	api.Use(middlewares.RateLimitMiddleware(limiter))
	{
		api.GET("/service", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"service": "payment service",
			})
		})
	}

	r.Run()

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
