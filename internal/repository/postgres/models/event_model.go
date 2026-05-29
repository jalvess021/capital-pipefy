package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type ProcessedEventModel struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
    EventID     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    CardID      string    `gorm:"type:varchar(255);not null"`
    ProcessedAt time.Time `gorm:"not null"`
    CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func (m *ProcessedEventModel) BeforeCreate(tx *gorm.DB) error {
    m.ID = uuid.New()
    return nil
}

func (m *ProcessedEventModel) TableName() string {
    return "processed_events"
}