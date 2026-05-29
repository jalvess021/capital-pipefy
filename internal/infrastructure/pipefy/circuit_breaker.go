package pipefy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/logger"
)

var ErrOpenCircuit = errors.New("circuit breaker open")

type noopCircuitBreaker struct{}

func (n *noopCircuitBreaker) Execute(fn func() (any, error)) (any, error) { return fn() }

type circuitBreaker interface {
	Execute(fn func() (any, error)) (any, error)
}

// redisCircuitBreaker compartilha estado entre multiplas instancias via Redis.
type redisCircuitBreaker struct {
	redis     *redis.Client
	threshold uint32
	timeout   time.Duration
	log       *zap.Logger
}

func newRedisCircuitBreaker(rdb *redis.Client, threshold uint32, timeout time.Duration, log *zap.Logger) circuitBreaker {
	if rdb == nil {
		return &noopCircuitBreaker{}
	}
	return &redisCircuitBreaker{
		redis:     rdb,
		threshold: threshold,
		timeout:   timeout,
		log:       log,
	}
}

const (
	cbStateKey    = "cb:pipefy:state"
	cbFailuresKey = "cb:pipefy:failures"
)

func (cb *redisCircuitBreaker) Execute(fn func() (any, error)) (any, error) {
	ctx := context.Background()

	state, _ := cb.redis.Get(ctx, cbStateKey).Result()
	if state == "open" {
		logger.InfraWarn(cb.log, "circuit breaker open, rejecting pipefy call")
		return nil, ErrOpenCircuit
	}

	result, err := fn()

	if err != nil {
		if isTransient(err) {
			count, _ := cb.redis.Incr(ctx, cbFailuresKey).Result()
			cb.redis.Expire(ctx, cbFailuresKey, cb.timeout)

			if uint32(count) >= cb.threshold {
				cb.redis.Set(ctx, cbStateKey, "open", cb.timeout)
				logger.InfraWarn(cb.log, "circuit breaker opened",
					zap.Int64("failures", count),
					zap.Duration("timeout", cb.timeout),
				)
			}
		}
		return nil, fmt.Errorf("pipefy call failed: %w", err)
	}

	cb.redis.Del(ctx, cbFailuresKey)
	return result, nil
}
