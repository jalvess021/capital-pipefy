package repository

import "github.com/jalvess021/capital-pipefy/internal/domain"

type EventRepository interface {
	SaveIfNotExists(event *domain.ProcessedEvent) error
}