package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/seunome/go-microservices/api-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("usuário já existe")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
)

type AuthService struct {
	repo   domain.UserRepository
	pub    domain.EventPublisher
	logger *slog.Logger
}

// Construtor (Injeção de Dependência)
func NewAuthService(r domain.UserRepository, p domain.EventPublisher, l *slog.Logger) *AuthService {
	return &AuthService{repo: r, pub: p, logger: l}
}

func (s *AuthService) Register(ctx context.Context, email, password string) error {
	s.logger.Info("Iniciando registro de usuário", "email", email)

	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Erro ao gerar hash da senha", "erro", err)
		return fmt.Errorf("falha ao processar senha: %w", err)
	}

	user := &domain.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	// Salva no banco (via interface)
	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error("Erro ao salvar usuário no banco", "erro", err)
		return ErrUserAlreadyExists
	}

	// Publica evento na fila (via interface)
	if err := s.pub.PublishUserRegistered(ctx, email); err != nil {
		// Logamos o erro, mas não falhamos o registro do usuário
		s.logger.Warn("Erro ao publicar evento no SQS", "erro", err)
	}

	s.logger.Info("Usuário registrado com sucesso", "email", email)
	return nil
}
