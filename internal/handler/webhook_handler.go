package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jalvess021/capital-pipefy/internal/apperrors"
	"github.com/jalvess021/capital-pipefy/internal/dto"
	"github.com/jalvess021/capital-pipefy/internal/service"
)

type WebhookHandler struct {
	service *service.WebhookService
}

func NewWebhookHandler(service *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

func (h *WebhookHandler) CardUpdated(c *gin.Context) {
	var req dto.CardUpdatedWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ProcessCardUpdated(req)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			// duplicate event — idempotent, return 200
			c.Status(http.StatusOK)
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusOK)
}
