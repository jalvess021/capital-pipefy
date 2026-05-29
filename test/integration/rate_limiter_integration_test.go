//go:build integration

package integration_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/middleware"
)

// Run with: go test -tags=integration ./test/integration/... -v
// Requires: REDIS_URL env var pointing to a real redis instance

func setupRateLimitRouter(t *testing.T, rps int) *gin.Engine {
	t.Helper()
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set — skipping rate limiter integration test")
	}

	rl, err := middleware.NewRateLimiter(url, rps, zap.NewNop())
	if err != nil {
		t.Skipf("redis unavailable, skipping rate limiter integration test: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(rl.Handle())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func sendRequest(r *gin.Engine) int {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	r := setupRateLimitRouter(t, 5)

	for i := 0; i < 5; i++ {
		if code := sendRequest(r); code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, code)
		}
	}
}

func TestRateLimiter_BlocksAfterLimit(t *testing.T) {
	r := setupRateLimitRouter(t, 3)

	for i := 0; i < 3; i++ {
		sendRequest(r)
	}

	if code := sendRequest(r); code != http.StatusTooManyRequests {
		t.Errorf("expected 429 after limit exceeded, got %d", code)
	}
}
