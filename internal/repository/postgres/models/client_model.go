package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type ClientModel struct {
    ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
    Nome            string    `gorm:"type:varchar(255);not null"`
    Email           string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    TipoSolicitacao string    `gorm:"type:varchar(100);not null"`
    ValorPatrimonio float64   `gorm:"not null"`
    Status          string    `gorm:"type:varchar(50);default:'Aguardando Análise'"`
    Prioridade      string    `gorm:"type:varchar(50);default:''"`
    CreatedAt       time.Time `gorm:"autoCreateTime"`
    UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (m *ClientModel) BeforeCreate(tx *gorm.DB) error {
    m.ID = uuid.New()
    return nil
}

func (m *ClientModel) TableName() string {
    return "clients"
}