package card

import (
	"context"
	"fmt"
)

const createCardMutation = `
	mutation CreateCard($input: CreateCardInput!) {
		createCard(input: $input) {
			card {
				id
				title
				url
			}
		}
	}`

const updateCardFieldMutation = `
	mutation UpdateCardField($input: UpdateCardFieldInput!) {
		updateCardField(input: $input) {
			card {
				id
				title
			}
			success
		}
	}`

type Executor interface {
	Execute(ctx context.Context, query string, variables any) error
}

type Operations struct {
	executor Executor
	pipeID   string
}

func NewOperations(executor Executor, pipeID string) *Operations {
	return &Operations{executor: executor, pipeID: pipeID}
}

func (o *Operations) CreateCard(ctx context.Context, name, email string, assetValue float64) error {
	input := CreateCardInput{
		PipeID: o.pipeID,
		Title:  name,
		FieldsAttributes: []FieldAttribute{
			{FieldID: "nome", FieldValue: name},
			{FieldID: "email", FieldValue: email},
			{FieldID: "valor_patrimonio", FieldValue: fmt.Sprintf("%.2f", assetValue)},
		},
	}
	return o.executor.Execute(ctx, createCardMutation, map[string]any{"input": input})
}

func (o *Operations) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	input := UpdateCardFieldInput{
		CardID:   cardID,
		FieldID:  fieldID,
		NewValue: value,
	}
	return o.executor.Execute(ctx, updateCardFieldMutation, map[string]any{"input": input})
}
