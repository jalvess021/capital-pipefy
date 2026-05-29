package postgres

import (
	"fmt"
	"time"

	"github.com/jalvess021/capital-pipefy/internal/apperrors"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/repository"
	"github.com/jalvess021/capital-pipefy/internal/repository/postgres/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) repository.EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) SaveIfNotExists(event *domain.ProcessedEvent) error {
	m := toEventModel(event)
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "event_id"}},
		DoNothing: true,
	}).Create(&m)

	if result.Error != nil {
		return fmt.Errorf("failed to save event: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("duplicate event %s: %w", event.EventID, apperrors.ErrConflict)
	}
	return nil
}

func toEventModel(d *domain.ProcessedEvent) models.ProcessedEventModel {
	return models.ProcessedEventModel{
		EventID:     d.EventID,
		CardID:      d.CardID,
		ProcessedAt: time.Now(),
	}
}