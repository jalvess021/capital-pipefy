package pipefy

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"github.com/sony/gobreaker"
	"github.com/jalvess021/capital-pipefy/internal/config"
	"github.com/jalvess021/capital-pipefy/internal/infrastructure/pipefy/card"
	"github.com/jalvess021/capital-pipefy/internal/logger"
)

type Client struct {
	apiURL  string
	token   string
	cfg     config.PipefyConfig
	http    *http.Client
	breaker *gobreaker.CircuitBreaker
	log     *zap.Logger
	Card    *card.Operations
}

func NewClient(cfg config.PipefyConfig, log *zap.Logger) *Client {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "pipefy",
		MaxRequests: 1,
		Interval:    60 * cfg.CBOpenTimeout / 30,
		Timeout:     cfg.CBOpenTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.CBThreshold
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.InfraWarn(log, "circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	})

	c := &Client{
		apiURL:  cfg.APIURL,
		token:   cfg.Token,
		cfg:     cfg,
		http:    &http.Client{Timeout: cfg.HTTPTimeout},
		breaker: cb,
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
