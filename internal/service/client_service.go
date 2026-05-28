package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/apperrors"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/dto"
	"github.com/jalvess021/capital-pipefy/internal/logger"
	"github.com/jalvess021/capital-pipefy/internal/port"
	"github.com/jalvess021/capital-pipefy/internal/repository"
)

type ClientService struct {
	repo   repository.ClientRepository
	pipefy port.Pipefy
	log    *zap.Logger
}

func NewClientService(repo repository.ClientRepository, pipefy port.Pipefy, log *zap.Logger) *ClientService {
	return &ClientService{repo: repo, pipefy: pipefy, log: log}
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
		logger.RequestError(s.log, "failed to save client", err,
			zap.String("email", req.Email),
		)
		return nil, fmt.Errorf("failed to save client: %w", apperrors.ErrInternal)
	}

	if err := s.pipefy.CreateCard(context.Background(), client.Nome, client.Email, client.ValorPatrimonio); err != nil {
		logger.RequestError(s.log, "failed to sync card to pipefy", err,
			zap.String("email", client.Email),
		)
		return nil, fmt.Errorf("failed to sync card to pipefy: %w", apperrors.ErrInternal)
	}

	logger.RequestInfo(s.log, "client created",
		zap.String("email", client.Email),
		zap.String("prioridade", client.Prioridade),
	)

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
