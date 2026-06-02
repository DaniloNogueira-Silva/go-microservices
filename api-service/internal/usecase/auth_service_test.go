package usecase_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/domain"
	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/usecase"
)

// Mock do Repositório
type MockUserRepository struct {
	CreateFunc func(ctx context.Context, user *domain.User) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.CreateFunc(ctx, user)
}
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

// Mock do Publisher
type MockEventPublisher struct {
	PublishFunc func(ctx context.Context, email string) error
}

func (m *MockEventPublisher) PublishUserRegistered(ctx context.Context, email string) error {
	return m.PublishFunc(ctx, email)
}

func TestAuthService_Register_Success(t *testing.T) {
	// Setup dos mocks
	mockRepo := &MockUserRepository{
		CreateFunc: func(ctx context.Context, user *domain.User) error {
			user.ID = 1 // Simula o ID gerado pelo banco
			return nil
		},
	}
	mockPub := &MockEventPublisher{
		PublishFunc: func(ctx context.Context, email string) error { return nil },
	}

	// Logger descartável para o teste
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := usecase.NewAuthService(mockRepo, mockPub, logger)

	// Execução
	err := service.Register(context.Background(), "test@test.com", "senha123")

	// Asserções
	if err != nil {
		t.Errorf("Esperava sucesso, mas obteve erro: %v", err)
	}
}
