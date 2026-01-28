package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/pkg/logger"
	"go.uber.org/zap"
)

// this is the first middleware
func LoggingMiddleware(ctx *gin.Context) {
	start := time.Now()
	ctx.Set("start_time", start)

	ctx.Next()

	latency := time.Since(start)
	logger.Info("request served",
		zap.String("client_ip", ctx.ClientIP()),
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
		zap.Int("status", ctx.Writer.Status()),
		zap.Duration("latency", latency),
		zap.Int("rate_limit_remaining", ctx.GetInt("rl_remaining")),
		zap.Bool("rate_limited", ctx.GetBool("rate_limited")),
	)
}
