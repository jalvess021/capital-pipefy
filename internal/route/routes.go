package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jalvess021/capital-pipefy/internal/bootstrap"
)

func SetupRouter(app *bootstrap.App) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	registerClientRoutes(router, app.Providers.ClientHandler)
	registerWebhookRoutes(router, app.Providers.WebhookHandler)

	return router
}
