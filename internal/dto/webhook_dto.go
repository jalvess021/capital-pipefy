package dto

type CardUpdatedWebhookRequest struct {
	EventID      string `json:"event_id" binding:"required"`
	CardID       string `json:"card_id" binding:"required"`
	ClienteEmail string `json:"cliente_email" binding:"required,email"`
	Timestamp    string `json:"timestamp" binding:"required"`
}