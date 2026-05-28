package pipefy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) Execute(ctx context.Context, query string, variables any) error {
	return c.do(ctx, graphQLRequest{Query: query, Variables: variables})
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

	var gqlResp graphQLResponse[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return fmt.Errorf("failed to decode graphql response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("pipefy graphql error: %s", gqlResp.Errors[0].Message)
	}

	return nil
}
