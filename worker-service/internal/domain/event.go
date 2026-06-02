package domain

import "context"

// Representa a mensagem que vem da fila
type UserEvent struct {
	Action string `json:"action"`
	Email  string `json:"email"`
}

// Interface que define o que a regra de negócio deve ser capaz de fazer
type EventProcessor interface {
	ProcessUserRegistration(ctx context.Context, email string) error
}
