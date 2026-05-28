package pipefy

type graphQLRequest struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type graphQLResponse[T any] struct {
	Data   T              `json:"data,omitempty"`
	Errors []graphQLError `json:"errors,omitempty"`
}
