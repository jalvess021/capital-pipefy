package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/logger"
)

// Mesmo contador para varias instancias do app
type RateLimiter struct {
	limiter *redis_rate.Limiter
	limit   int
	log     *zap.Logger
}

func NewRateLimiter(redisURL string, rps int, log *zap.Logger) (*RateLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RateLimiter{limiter: redis_rate.NewLimiter(client), limit: rps, log: log}, nil
}

func (rl *RateLimiter) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		ctx := c.Request.Context()

		limit := redis_rate.Limit{Rate: rl.limit, Burst: rl.limit * 2, Period: time.Second}
		res, err := rl.limiter.Allow(ctx, fmt.Sprintf("rate:%s", ip), limit)
		if err != nil {
			logger.InfraWarn(rl.log, "redis unavailable, skipping rate limit",
				zap.Error(err),
			)
			c.Next()
			return
		}

		if res.Allowed == 0 {
			logger.RequestWarn(rl.log, "request blocked",
				zap.String("ip", ip),
				zap.String("reason", "rate_limit_exceeded"),
				zap.Int("remaining", res.Remaining),
				zap.Int("limit", rl.limit),
			)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
