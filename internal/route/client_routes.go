package route

import (
	"github.com/gin-gonic/gin"
	"github.com/jalvess021/capital-pipefy/internal/handler"
)

func registerClientRoutes(r *gin.Engine, h *handler.ClientHandler) {
	{
		clients := r.Group("/clientes")
		{
			clients.POST("", h.Create)
		}
	}
}
