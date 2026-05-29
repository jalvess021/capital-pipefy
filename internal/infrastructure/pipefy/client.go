package pipefy

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"github.com/redis/go-redis/v9"
	"github.com/jalvess021/capital-pipefy/internal/config"
	"github.com/jalvess021/capital-pipefy/internal/infrastructure/pipefy/card"
)

type Client struct {
	apiURL  string
	token   string
	cfg     config.PipefyConfig
	http    *http.Client
	breaker circuitBreaker
	log     *zap.Logger
	Card    *card.Operations
}

func NewClient(cfg config.PipefyConfig, rdb *redis.Client, log *zap.Logger) *Client {
	c := &Client{
		apiURL:  cfg.APIURL,
		token:   cfg.Token,
		cfg:     cfg,
		http:    &http.Client{Timeout: cfg.HTTPTimeout},
		breaker: newRedisCircuitBreaker(rdb, cfg.CBThreshold, cfg.CBOpenTimeout, log),
		log:     log,
	}
	c.Card = card.NewOperations(c, cfg.PipeID)
	return c
}

func (c *Client) CreateCard(ctx context.Context, name, email string, assetValue float64) error {
	return c.Card.CreateCard(ctx, name, email, assetValue)
}

func (c *Client) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	return c.Card.UpdateCardField(ctx, cardID, fieldID, value)
}
