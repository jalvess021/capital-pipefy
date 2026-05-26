package interfaces

import "github.com/jalvess021/capital-pipefy/internal/domain"

type ClientRepository interface {
    Save(client *domain.Client) error
    FindByEmail(email string) (*domain.Client, error)
    UpdateStatusAndPriority(email, status, prioridade string) error
}