package domain

import "context"

// Entidade principal
type User struct {
	ID       uint
	Email    string
	Password string
}

// Portas (Interfaces) que a infraestrutura terá que implementar
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type EventPublisher interface {
	PublishUserRegistered(ctx context.Context, email string) error
}
