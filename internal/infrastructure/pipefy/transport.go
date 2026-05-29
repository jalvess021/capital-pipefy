package pipefy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/logger"
)

func (c *Client) Execute(ctx context.Context, query string, variables any) error {
	payload := graphQLRequest{Query: query, Variables: variables}

	_, err := c.breaker.Execute(func() (any, error) {
		return nil, c.doWithRetry(ctx, payload)
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("pipefy call failed: %w", err)
	}
	return nil
}

func (c *Client) doWithRetry(ctx context.Context, payload graphQLRequest) error {
	var lastErr error
	delay := c.cfg.RetryDelay

	for attempt := 1; attempt <= c.cfg.MaxRetries; attempt++ {
		start := time.Now()
		err := c.do(ctx, payload)
		latency := time.Since(start)

		if err == nil {
			if attempt > 1 {
				logger.InfraInfo(c.log, "pipefy call succeeded after retry",
					zap.Int("attempt", attempt),
					zap.Duration("latency", latency),
				)
			}
			return nil
		}

		if !isTransient(err) {
			logger.InfraError(c.log, "pipefy non-transient error, aborting retry", err,
				zap.Int("attempt", attempt),
				zap.Duration("latency", latency),
			)
			return err
		}

		logger.InfraWarn(c.log, "pipefy transient error, will retry",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", c.cfg.MaxRetries),
			zap.Duration("latency", latency),
			zap.Duration("retry_in", delay),
			zap.Error(err),
		)

		lastErr = err

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			delay *= 2
		}
	}

	return fmt.Errorf("all %d attempts failed: %w", c.cfg.MaxRetries, lastErr)
}

func isTransient(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, pattern := range []string{
		"status 502", "status 503", "status 504",
		"connection reset", "EOF", "timeout",
		"unexpected status 502", "unexpected status 503", "unexpected status 504",
	} {
		if contains(msg, pattern) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
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
