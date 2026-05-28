package service

import (
	"context"
	"fmt"

	"github.com/jalvess021/capital-pipefy/internal/apperrors"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/dto"
	"github.com/jalvess021/capital-pipefy/internal/port"
	"github.com/jalvess021/capital-pipefy/internal/repository"
)

type ClientService struct {
	repo   repository.ClientRepository
	pipefy port.Pipefy
}

func NewClientService(repo repository.ClientRepository, pipefy port.Pipefy) *ClientService {
	return &ClientService{repo: repo, pipefy: pipefy}
}

func (s *ClientService) Create(req dto.CreateClientRequest) (*dto.ClientResponse, error) {
	_, err := s.repo.FindByEmail(req.Email)
	if err == nil {
		return nil, fmt.Errorf("client with email %s already exists: %w", req.Email, apperrors.ErrConflict)
	}

	client := &domain.Client{
		Nome:            req.Nome,
		Email:           req.Email,
		TipoSolicitacao: req.TipoSolicitacao,
		ValorPatrimonio: req.ValorPatrimonio,
		Status:          "Aguardando Análise",
		Prioridade:      calculatePriority(req.ValorPatrimonio),
	}

	if err := s.repo.Save(client); err != nil {
		return nil, fmt.Errorf("failed to save client: %w", apperrors.ErrInternal)
	}

	if err := s.pipefy.CreateCard(context.Background(), client.Nome, client.Email, client.ValorPatrimonio); err != nil {
		return nil, fmt.Errorf("failed to sync card to pipefy: %w", apperrors.ErrInternal)
	}

	return &dto.ClientResponse{
		ID:              client.ID,
		Nome:            client.Nome,
		Email:           client.Email,
		TipoSolicitacao: client.TipoSolicitacao,
		ValorPatrimonio: client.ValorPatrimonio,
		Status:          client.Status,
		Prioridade:      client.Prioridade,
	}, nil
}

func calculatePriority(assetValue float64) string {
	if assetValue >= 200_000 {
		return "prioridade_alta"
	}
	return "prioridade_normal"
}
