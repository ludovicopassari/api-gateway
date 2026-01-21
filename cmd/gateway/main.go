package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/limiters"
	"github.com/ludovicopassari/api-gateway/internal/limiters/storage"
	"github.com/ludovicopassari/api-gateway/internal/middlewares"
	"github.com/ludovicopassari/api-gateway/pkg/logger"
	"go.uber.org/zap"
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

	logger.Info("Starting API Gateway",
		zap.String("env", logCfg.Environment),
		zap.String("level", logCfg.Level),
		zap.Time("time", time.Now()),
	)

	// setup connection with in-memory DB
	rdb, err := storage.NewRedis("redis:6379")
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	limiter := limiters.NewFixedWindow(rdb, 10, time.Second*10)

	// setup Gin middlewares
	r.Use(middlewares.RateLimitMiddleware(limiter))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
