//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/config"
	pipefy "github.com/jalvess021/capital-pipefy/internal/infrastructure/pipefy"
)

func newCBTestClient(t *testing.T, serverURL string) *pipefy.Client {
	t.Helper()
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set — skipping circuit breaker integration test")
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("invalid REDIS_URL: %v", err)
	}
	rdb := redis.NewClient(opts)

	// limpa estado anterior
	rdb.Del(context.Background(), "cb:pipefy:state", "cb:pipefy:failures")

	cfg := config.PipefyConfig{
		APIURL:        serverURL,
		Token:         "test-token",
		PipeID:        "test-pipe",
		HTTPTimeout:   2 * time.Second,
		MaxRetries:    1,
		RetryDelay:    10 * time.Millisecond,
		CBThreshold:   3,
		CBOpenTimeout: 5 * time.Second,
	}
	return pipefy.NewClient(cfg, rdb, zap.NewNop())
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := newCBTestClient(t, srv.URL)

	for i := 0; i < 3; i++ {
		c.Execute(context.Background(), `{ test }`, nil)
	}

	callsBefore := calls.Load()
	err := c.Execute(context.Background(), `{ test }`, nil)

	if err == nil {
		t.Error("expected circuit breaker error")
	}
	if calls.Load() != callsBefore {
		t.Errorf("expected no HTTP calls when circuit open, got %d new calls", calls.Load()-callsBefore)
	}
}

func TestCircuitBreaker_ClosesAfterTimeout(t *testing.T) {
	var calls atomic.Int32
	succeed := atomic.Bool{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if succeed.Load() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{}}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer srv.Close()

	c := newCBTestClient(t, srv.URL)

	// abre o CB
	for i := 0; i < 3; i++ {
		c.Execute(context.Background(), `{ test }`, nil)
	}

	// aguarda timeout (5s configurado)
	time.Sleep(6 * time.Second)
	succeed.Store(true)

	err := c.Execute(context.Background(), `{ test }`, nil)
	if err != nil {
		t.Errorf("expected CB to close after timeout, got %v", err)
	}
}
