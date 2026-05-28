package port

import "context"

type Pipefy interface {
	CreateCard(ctx context.Context, name, email string, assetValue float64) error
	UpdateCardField(ctx context.Context, cardID, fieldID, value string) error
}
