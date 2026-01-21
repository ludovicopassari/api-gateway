package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/limiter"
	"github.com/ludovicopassari/api-gateway/internal/middlewares"
	"github.com/ludovicopassari/api-gateway/pkg/logger"
	"go.uber.org/zap"
)

func main() {
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

	limiter := limiter.NewMemoryLimiter(10, time.Minute)

	r := gin.Default()
	r.Use(middlewares.RateLimitMiddleware(limiter))

	r.GET("/ping", func(c *gin.Context) {
		r := c.Request
		logger.Info("Request received",

			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", r.UserAgent()),
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	r.Run()

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
