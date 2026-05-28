package pipefy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const createCardMutation = `
	mutation CreateCard($input: CreateCardInput!) {
		createCard(input: $input) {
			card {
				id
				title
				url
			}
		}
	}`

const updateCardFieldMutation = `
	mutation UpdateCardField($input: UpdateCardFieldInput!) {
		updateCardField(input: $input) {
			card {
				id
				title
			}
			success
		}
	}`

type Client struct {
	apiURL string
	token  string
	pipeID string
	http   *http.Client
}

func NewClient(apiURL, token, pipeID string) *Client {
	return &Client{
		apiURL: apiURL,
		token:  token,
		pipeID: pipeID,
		http:   &http.Client{},
	}
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func (c *Client) CreateCard(ctx context.Context, nome, email string, valorPatrimonio float64) error {
	payload := graphQLRequest{
		Query: createCardMutation,
		Variables: map[string]any{
			"input": map[string]any{
				"pipe_id": c.pipeID,
				"title":   nome,
				"fields_attributes": []map[string]any{
					{"field_id": "nome", "field_value": nome},
					{"field_id": "email", "field_value": email},
					{"field_id": "valor_patrimonio", "field_value": fmt.Sprintf("%.2f", valorPatrimonio)},
				},
			},
		},
	}
	return c.do(ctx, payload)
}

func (c *Client) UpdateCardField(ctx context.Context, cardID, status, priority string) error {
	payload := graphQLRequest{
		Query: updateCardFieldMutation,
		Variables: map[string]any{
			"input": map[string]any{
				"card_id":  cardID,
				"field_id": "status",
				"new_value": status,
			},
		},
	}
	return c.do(ctx, payload)
}

func (c *Client) do(ctx context.Context, payload graphQLRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send graphql request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pipefy returned unexpected status %d", resp.StatusCode)
	}

	return nil
}
