package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/apperrors"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/dto"
	"github.com/jalvess021/capital-pipefy/internal/logger"
	"github.com/jalvess021/capital-pipefy/internal/port"
	"github.com/jalvess021/capital-pipefy/internal/repository"
)

type WebhookService struct {
	clientRepo repository.ClientRepository
	eventRepo  repository.EventRepository
	pipefy     port.Pipefy
	log        *zap.Logger
}

func NewWebhookService(
	clientRepo repository.ClientRepository,
	eventRepo repository.EventRepository,
	pipefy port.Pipefy,
	log *zap.Logger,
) *WebhookService {
	return &WebhookService{clientRepo: clientRepo, eventRepo: eventRepo, pipefy: pipefy, log: log}
}

func (s *WebhookService) ProcessCardUpdated(req dto.CardUpdatedWebhookRequest) error {
	exists, err := s.eventRepo.ExistsByEventID(req.EventID)
	if err != nil {
		logger.WebhookError(s.log, "failed to check event_id", err,
			zap.String("event_id", req.EventID),
		)
		return fmt.Errorf("failed to check event: %w", apperrors.ErrInternal)
	}
	if exists {
		logger.WebhookWarn(s.log, "duplicate event ignored",
			zap.String("event_id", req.EventID),
			zap.String("reason", "already processed"),
		)
		return fmt.Errorf("duplicate event %s: %w", req.EventID, apperrors.ErrConflict)
	}

	client, err := s.clientRepo.FindByEmail(req.ClienteEmail)
	if err != nil {
		logger.WebhookError(s.log, "client not found", err,
			zap.String("event_id", req.EventID),
			zap.String("cliente_email", req.ClienteEmail),
		)
		return fmt.Errorf("client not found: %w", apperrors.ErrNotFound)
	}

	priority := calculatePriority(client.ValorPatrimonio)

	if err := s.clientRepo.UpdateStatusAndPriority(req.ClienteEmail, "Processado", priority); err != nil {
		logger.WebhookError(s.log, "failed to update client", err,
			zap.String("event_id", req.EventID),
			zap.String("cliente_email", req.ClienteEmail),
		)
		return fmt.Errorf("failed to update client: %w", apperrors.ErrInternal)
	}

	if err := s.pipefy.UpdateCardField(context.Background(), req.CardID, "status", "Processado"); err != nil {
		logger.WebhookError(s.log, "failed to sync card status to pipefy", err,
			zap.String("event_id", req.EventID),
			zap.String("card_id", req.CardID),
		)
		return fmt.Errorf("failed to sync card to pipefy: %w", apperrors.ErrInternal)
	}

	event := &domain.ProcessedEvent{
		EventID:     req.EventID,
		CardID:      req.CardID,
		ProcessedAt: time.Now(),
	}
	if err := s.eventRepo.Save(event); err != nil {
		logger.WebhookError(s.log, "failed to save processed event", err,
			zap.String("event_id", req.EventID),
		)
		return fmt.Errorf("failed to save event: %w", apperrors.ErrInternal)
	}

	logger.WebhookInfo(s.log, "event processed",
		zap.String("event_id", req.EventID),
		zap.String("card_id", req.CardID),
		zap.String("cliente_email", req.ClienteEmail),
		zap.String("prioridade", priority),
	)

	return nil
}
