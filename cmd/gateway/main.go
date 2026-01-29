package main

import (
	"time"

	"github.com/ludovicopassari/api-gateway/internal/limiters"
	"github.com/ludovicopassari/api-gateway/internal/monitoring"
	"github.com/ludovicopassari/api-gateway/internal/routers"
	redistorage "github.com/ludovicopassari/api-gateway/internal/storage/redis"
	"github.com/ludovicopassari/api-gateway/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
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

	if err := redistorage.Init(
		redistorage.WithAddr("redis:6379"),
		redistorage.WithPassword(""),
	); err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redistorage.Close()

	redisdb, err := redistorage.RedisClient()
	if err != nil {
		logger.Fatal("Failed to get Redis client", zap.Error(err))
	}

	limiter := limiters.NewSlidingWindowLimiter(redisdb, 100, time.Minute)
	r := routers.SetupRouter(reg, m, limiter)

	r.Run(":" + "80")

}
