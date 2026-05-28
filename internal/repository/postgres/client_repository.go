package postgres

import (
    "fmt"

    "github.com/jalvess021/capital-pipefy/internal/domain"
    "github.com/jalvess021/capital-pipefy/internal/repository"
    "github.com/jalvess021/capital-pipefy/internal/repository/postgres/models"
    "gorm.io/gorm"
)

type clientRepository struct {
    db *gorm.DB
}

func NewClientRepository(db *gorm.DB) repository.ClientRepository {
    return &clientRepository{db: db}
}

func (r *clientRepository) Save(client *domain.Client) error {
    m := toClientModel(client)
    if err := r.db.Create(&m).Error; err != nil {
        return fmt.Errorf("failed to save client: %w", err)
    }
    client.ID = m.ID.String()
    return nil
}

func (r *clientRepository) FindByEmail(email string) (*domain.Client, error) {
    var m models.ClientModel
    if err := r.db.Where("email = ?", email).First(&m).Error; err != nil {
        return nil, fmt.Errorf("client not found: %w", err)
    }
    c := toClientDomain(&m)
    return &c, nil
}

func (r *clientRepository) UpdateStatusAndPriority(email, status, prioridade string) error {
    result := r.db.Model(&models.ClientModel{}).
        Where("email = ?", email).
        Updates(map[string]interface{}{
            "status":     status,
            "prioridade": prioridade,
        })
    if result.Error != nil {
        return fmt.Errorf("failed to update client: %w", result.Error)
    }
    return nil
}

func toClientModel(d *domain.Client) models.ClientModel {
    return models.ClientModel{
        Nome:            d.Nome,
        Email:           d.Email,
        TipoSolicitacao: d.TipoSolicitacao,
        ValorPatrimonio: d.ValorPatrimonio,
        Status:          d.Status,
        Prioridade:      d.Prioridade,
    }
}

func toClientDomain(m *models.ClientModel) domain.Client {
    return domain.Client{
        ID:              m.ID.String(),
        Nome:            m.Nome,
        Email:           m.Email,
        TipoSolicitacao: m.TipoSolicitacao,
        ValorPatrimonio: m.ValorPatrimonio,
        Status:          m.Status,
        Prioridade:      m.Prioridade,
        CreatedAt:       m.CreatedAt,
        UpdatedAt:       m.UpdatedAt,
    }
}