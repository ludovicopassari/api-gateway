package main

import (
	"time"

	"github.com/ludovicopassari/api-gateway/internal/limiters"
	"github.com/ludovicopassari/api-gateway/internal/limiters/storage"
	"github.com/ludovicopassari/api-gateway/internal/monitoring"
	"github.com/ludovicopassari/api-gateway/internal/routers"
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

	// setup connection with in-memory DB
	rdb, err := storage.NewRedis("redis:6379")
	if err != nil {
		panic(err)
	}

	limiter := limiters.NewSlidingWindowLimiter(rdb, 10, time.Minute)
	r := routers.SetupRouter(m, limiter)

	r.Run(":" + "80")

}
