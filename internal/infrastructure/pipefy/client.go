package pipefy

import (
	"context"
	"net/http"

	"github.com/jalvess021/capital-pipefy/internal/infrastructure/pipefy/card"
)

type Client struct {
	apiURL string
	token  string
	http   *http.Client
	Card   *card.Operations
}

func NewClient(apiURL, token, pipeID string) *Client {
	c := &Client{
		apiURL: apiURL,
		token:  token,
		http:   &http.Client{},
	}
	c.Card = card.NewOperations(c, pipeID)
	return c
}

func (c *Client) CreateCard(ctx context.Context, name, email string, assetValue float64) error {
	return c.Card.CreateCard(ctx, name, email, assetValue)
}

func (c *Client) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	return c.Card.UpdateCardField(ctx, cardID, fieldID, value)
}
