package postgres

import (
    "fmt"
    "time"

    "github.com/jalvess021/capital-pipefy/internal/domain"
    "github.com/jalvess021/capital-pipefy/internal/repository"
    "github.com/jalvess021/capital-pipefy/internal/repository/postgres/models"
    "gorm.io/gorm"
)

type eventRepository struct {
    db *gorm.DB
}

func NewEventRepository(db *gorm.DB) repository.EventRepository {
    return &eventRepository{db: db}
}

func (r *eventRepository) Save(event *domain.ProcessedEvent) error {
    m := toEventModel(event)
    if err := r.db.Create(&m).Error; err != nil {
        return fmt.Errorf("failed to save event: %w", err)
    }
    return nil
}

func (r *eventRepository) ExistsByEventID(eventID string) (bool, error) {
    var count int64
    err := r.db.Model(&models.ProcessedEventModel{}).
        Where("event_id = ?", eventID).
        Count(&count).Error
    if err != nil {
        return false, fmt.Errorf("failed to check event: %w", err)
    }
    return count > 0, nil
}

func toEventModel(d *domain.ProcessedEvent) models.ProcessedEventModel {
    return models.ProcessedEventModel{
        EventID:     d.EventID,
        CardID:      d.CardID,
        ProcessedAt: time.Now(),
    }
}