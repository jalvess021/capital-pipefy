package bootstrap

import (
	"context"
	"time"

	"go.uber.org/zap"
	"github.com/redis/go-redis/v9"
	"github.com/jalvess021/capital-pipefy/internal/config"
	"github.com/jalvess021/capital-pipefy/internal/database"
	"github.com/jalvess021/capital-pipefy/internal/handler"
	pipefyclient "github.com/jalvess021/capital-pipefy/internal/infrastructure/pipefy"
	"github.com/jalvess021/capital-pipefy/internal/logger"
	postgresrepo "github.com/jalvess021/capital-pipefy/internal/repository/postgres"
	"github.com/jalvess021/capital-pipefy/internal/service"
)

type Providers struct {
	ClientHandler  *handler.ClientHandler
	WebhookHandler *handler.WebhookHandler
}

func buildProviders(db *database.PostgresDB, cfg *config.Config, log *zap.Logger) *Providers {
	clientRepo := postgresrepo.NewClientRepository(db.GormDB())
	eventRepo := postgresrepo.NewEventRepository(db.GormDB())

	rdb := buildRedisClient(cfg.RateLimit.RedisURL, log)
	pipefy := pipefyclient.NewClient(cfg.Pipefy, rdb, log)

	clientSvc := service.NewClientService(clientRepo, pipefy, log)
	webhookSvc := service.NewWebhookService(clientRepo, eventRepo, pipefy, log)

	return &Providers{
		ClientHandler:  handler.NewClientHandler(clientSvc),
		WebhookHandler: handler.NewWebhookHandler(webhookSvc),
	}
}

func buildRedisClient(redisURL string, log *zap.Logger) *redis.Client {
	if redisURL == "" {
		return nil
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.ApplicationWarn(log, "invalid redis URL, circuit breaker disabled")
		return nil
	}
	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.ApplicationWarn(log, "redis unavailable, circuit breaker disabled")
		return nil
	}
	return rdb
}
