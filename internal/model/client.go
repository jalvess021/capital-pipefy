package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Client struct {
    ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
    Nome            string         `gorm:"type:varchar(255);not null" json:"nome"`
    Email           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
    TipoSolicitacao string         `gorm:"type:varchar(100);not null" json:"tipo_solicitacao"`
    ValorPatrimonio float64        `gorm:"not null" json:"valor_patrimonio"`
    Status          string         `gorm:"type:varchar(50);default:'Aguardando Análise'" json:"status"`
    Prioridade      string         `gorm:"type:varchar(50);default:''" json:"prioridade"`
    CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (c *Client) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}

