package route

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/bootstrap"
	"github.com/jalvess021/capital-pipefy/internal/logger"
	"github.com/jalvess021/capital-pipefy/internal/middleware"
)

func SetupRouter(app *bootstrap.App) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(requestLogger(app.Log))
	applyRateLimiter(router, app)

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	registerClientRoutes(router, app.Providers.ClientHandler)
	registerWebhookRoutes(router, app.Providers.WebhookHandler)

	return router
}

func requestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		reqID, _ := c.Get(middleware.RequestIDKey)
		fields := []zap.Field{
			zap.String("request_id", reqID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
			zap.String("source", "router"),
		}
		switch {
		case status >= 500:
			logger.RequestError(log, "request", nil, fields...)
		case status >= 400:
			logger.RequestWarn(log, "request", fields...)
		default:
			logger.RequestInfo(log, "request", fields...)
		}
	}
}

func applyRateLimiter(r *gin.Engine, app *bootstrap.App) {
	if !app.Config.RateLimit.Enabled {
		logger.ApplicationWarn(app.Log, "rate limiter disabled, RATE_LIMIT_ENABLED=false")
		return
	}
	if app.Config.RateLimit.RedisURL == "" {
		logger.ApplicationWarn(app.Log, "rate limiter disabled, REDIS_URL not set")
		return
	}
	rl, err := middleware.NewRateLimiter(app.Config.RateLimit.RedisURL, app.Config.RateLimit.RPS, app.Log)
	if err != nil {
		logger.ApplicationWarn(app.Log, "rate limiter unavailable, running without it",
			zap.Error(err),
		)
		return
	}
	r.Use(rl.Handle())
}
