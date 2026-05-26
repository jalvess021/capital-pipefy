package repository

import "github.com/jalvess021/capital-pipefy/internal/domain"

type EventRepository interface {
    Save(event *domain.ProcessedEvent) error
    ExistsByEventID(eventID string) (bool, error)
}