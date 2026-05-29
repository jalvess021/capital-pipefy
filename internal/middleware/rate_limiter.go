package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/logger"
)

// Mesmo contador para varias instancias do app
type RateLimiter struct {
	redis *redis.Client
	limit int
	log   *zap.Logger
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

	return &RateLimiter{redis: client, limit: rps, log: log}, nil
}

func (rl *RateLimiter) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate:%s", ip)
		ctx := c.Request.Context()

		count, err := rl.redis.Incr(ctx, key).Result()
		if err != nil {
			logger.InfraWarn(rl.log, "redis unavailable, skipping rate limit",
				zap.Error(err),
			)
			c.Next()
			return
		}

		if count == 1 {
			rl.redis.Expire(ctx, key, time.Second)
		}

		if int(count) > rl.limit {
			logger.RequestWarn(rl.log, "request blocked",
				zap.String("ip", ip),
				zap.String("reason", "rate_limit_exceeded"),
				zap.Int64("count", count),
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
