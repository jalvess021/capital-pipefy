package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/bootstrap"
	"github.com/jalvess021/capital-pipefy/internal/logger"
	"github.com/jalvess021/capital-pipefy/internal/middleware"
)

func SetupRouter(app *bootstrap.App) *gin.Engine {
	router := gin.Default()

	applyRateLimiter(router, app)

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	registerClientRoutes(router, app.Providers.ClientHandler)
	registerWebhookRoutes(router, app.Providers.WebhookHandler)

	return router
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
