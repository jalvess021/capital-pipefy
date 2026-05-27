package service

import (
	"fmt"
	"time"

	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/dto"
	"github.com/jalvess021/capital-pipefy/internal/repository"
)

type WebhookService struct {
	clientRepo repository.ClientRepository
	eventRepo  repository.EventRepository
}

func NewWebhookService(
	clientRepo repository.ClientRepository,
	eventRepo repository.EventRepository,
) *WebhookService {
	return &WebhookService{
		clientRepo: clientRepo,
		eventRepo:  eventRepo,
	}
}

func (s *WebhookService) ProcessCardUpdated(req dto.CardUpdatedWebhookRequest) error {
	exists, err := s.eventRepo.ExistsByEventID(req.EventID)
	if err != nil {
		return fmt.Errorf("failed to check event: %w", err)
	}
	if exists {
		return fmt.Errorf("event %s already processed", req.EventID)
	}

	client, err := s.clientRepo.FindByEmail(req.ClienteEmail)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	priority := calculatePriority(client.ValorPatrimonio)

	if err := s.clientRepo.UpdateStatusAndPriority(req.ClienteEmail, "Processado", priority); err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	event := &domain.ProcessedEvent{
		EventID:     req.EventID,
		CardID:      req.CardID,
		ProcessedAt: time.Now(),
	}
	if err := s.eventRepo.Save(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}