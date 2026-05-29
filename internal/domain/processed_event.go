package domain

import "time"

type ProcessedEvent struct {
    ID          string
    EventID     string
    CardID      string
    ProcessedAt time.Time
    CreatedAt   time.Time
}