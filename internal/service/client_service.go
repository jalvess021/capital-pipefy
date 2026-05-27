package service

import (
	"fmt"

	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/dto"
	"github.com/jalvess021/capital-pipefy/internal/repository"
)

type ClientService struct {
	repo repository.ClientRepository
}

func NewClientService(repo repository.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

func (s *ClientService) Create(req dto.CreateClientRequest) (*dto.ClientResponse, error) {
	_, err := s.repo.FindByEmail(req.Email)
	if err == nil {
		return nil, fmt.Errorf("client with email %s already exists", req.Email)
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
		return nil, fmt.Errorf("failed to save client: %w", err)
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