package pipefy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/config"
	pipefy "github.com/jalvess021/capital-pipefy/internal/infrastructure/pipefy"
)

func testConfig(serverURL string) config.PipefyConfig {
	return config.PipefyConfig{
		APIURL:        serverURL,
		Token:         "test-token",
		PipeID:        "pipe-123",
		HTTPTimeout:   5 * time.Second,
		MaxRetries:    3,
		RetryDelay:    10 * time.Millisecond,
		CBThreshold:   5,
		CBOpenTimeout: 30 * time.Second,
	}
}

func newTestClient(serverURL string) *pipefy.Client {
	return pipefy.NewClient(testConfig(serverURL), nil, zap.NewNop())
}

func TestExecute_RetryOnFailureThenSuccess(t *testing.T) {
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 — transitório
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.Execute(context.Background(), `{ test }`, nil)

	if err != nil {
		t.Errorf("expected success after retry, got: %v", err)
	}
	if calls.Load() != 3 {
		t.Errorf("expected 3 calls (2 failures + 1 success), got %d", calls.Load())
	}
}

func TestExecute_AllRetriesExhausted_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.Execute(context.Background(), `{ test }`, nil)

	if err == nil {
		t.Error("expected error after all retries exhausted")
	}
}

func TestExecute_NonTransientError_NoRetry(t *testing.T) {
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized) // 401 — erro permanente, sem retry
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.Execute(context.Background(), `{ test }`, nil)

	if err == nil {
		t.Error("expected error on 401")
	}
	if calls.Load() != 1 {
		t.Errorf("401 must not retry, expected 1 call, got %d", calls.Load())
	}
}

