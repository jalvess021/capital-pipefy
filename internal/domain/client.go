package domain

import "time"

type Client struct {
    ID              string
    Nome            string
    Email           string
    TipoSolicitacao string
    ValorPatrimonio float64
    Status          string
    Prioridade      string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}