package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProcessedEvent struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"event_id"`
	CardID      string    `gorm:"type:varchar(255);not null" json:"card_id"`
	ProcessedAt time.Time `gorm:"not null" json:"processed_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (e *ProcessedEvent) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New()
	return
}
