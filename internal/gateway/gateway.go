package gateway

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ludovicopassari/api-gateway/internal/limiters"
	"github.com/ludovicopassari/api-gateway/internal/monitoring"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Gateway struct {
	logger  *zap.Logger
	redis   *redis.Client
	metrics *monitoring.Metrics
	limiter *limiters.RateLimiter
	router  *gin.Engine
	config  *GatewayConfig
}

type GatewayConfig struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RateLimit     int
	RatePeriod    time.Duration
}
