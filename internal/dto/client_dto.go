package dto

type CreateClientRequest struct {
	Nome            string  `json:"cliente_nome" binding:"required"`
	Email           string  `json:"cliente_email" binding:"required,email"`
	TipoSolicitacao string  `json:"tipo_solicitacao" binding:"required"`
	ValorPatrimonio float64 `json:"valor_patrimonio" binding:"required,gt=0"`
}

type ClientResponse struct {
	ID              string  `json:"id"`
	Nome            string  `json:"nome"`
	Email           string  `json:"email"`
	TipoSolicitacao string  `json:"tipo_solicitacao"`
	ValorPatrimonio float64 `json:"valor_patrimonio"`
	Status          string  `json:"status"`
	Prioridade      string  `json:"prioridade"`
}