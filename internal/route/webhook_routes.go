package route

import (
	"github.com/gin-gonic/gin"
	"github.com/jalvess021/capital-pipefy/internal/handler"
)

func registerWebhookRoutes(rg *gin.Engine, h *handler.WebhookHandler) {
	{
		router := rg.Group("/webhooks")
		{
			pipefy := router.Group("/pipefy")
			{
				pipefy.POST("/card-updated", h.CardUpdated)
			}
		}
	}
}
